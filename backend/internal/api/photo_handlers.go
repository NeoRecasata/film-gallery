package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/NeoRecasata/film-gallery/backend/internal/models"
	slugpkg "github.com/NeoRecasata/film-gallery/backend/internal/slug"
)

const maxUploadSize = 100 << 20 // 100MB

func (s *Server) handleUploadPhoto(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		Error(w, http.StatusRequestEntityTooLarge, "file too large (max 100MB)")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		Error(w, http.StatusBadRequest, "missing photo field")
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".tiff" && ext != ".tif" {
		Error(w, http.StatusBadRequest, "unsupported file type (use JPEG, PNG, or TIFF)")
		return
	}

	// Read file into memory for processing
	data, err := io.ReadAll(file)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	// Extract dimensions
	origWidth, origHeight, err := s.Processor.ExtractDimensions(bytes.NewReader(data))
	if err != nil {
		Error(w, http.StatusBadRequest, "failed to read image dimensions")
		return
	}

	// Generate variants
	variants, err := s.Processor.GenerateVariants(bytes.NewReader(data))
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to process image")
		return
	}

	// Generate BlurHash from thumbnail
	blurHash, err := s.Processor.GenerateBlurHash(bytes.NewReader(variants["thumb"].Data))
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate blur hash")
		return
	}

	// Generate photo ID first for storage keys
	var photoID string
	err = s.DB.QueryRow("SELECT gen_random_uuid()").Scan(&photoID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate ID")
		return
	}

	// Generate slug
	title := r.FormValue("title")
	fallback := photoID[:8]
	photoSlug := slugpkg.Generate(title, fallback)

	// Ensure slug uniqueness
	photoSlug, err = s.ensureUniqueSlug(r.Context(), "photos", photoSlug)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate slug")
		return
	}

	ctx := r.Context()

	// Store original
	originalKey := fmt.Sprintf("photos/%s/original%s", photoID, ext)
	if err := s.Storage.Put(ctx, originalKey, bytes.NewReader(data)); err != nil {
		Error(w, http.StatusInternalServerError, "failed to store original")
		return
	}

	// Store variants
	variantKeys := make(models.PhotoVariants)
	for name, v := range variants {
		key := fmt.Sprintf("photos/%s/%s.webp", photoID, name)
		if err := s.Storage.Put(ctx, key, bytes.NewReader(v.Data)); err != nil {
			Error(w, http.StatusInternalServerError, fmt.Sprintf("failed to store %s variant", name))
			return
		}
		variantKeys[name] = key
	}

	variantsJSON, err := json.Marshal(variantKeys)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to serialize variants")
		return
	}

	// Get next sort order
	var nextOrder int
	s.DB.QueryRow("SELECT COALESCE(MAX(sort_order), 0) + 1 FROM photos").Scan(&nextOrder)

	// Optional metadata from form
	var titlePtr, descPtr, filmStockPtr, cameraPtr, lensPtr *string
	var takenAtPtr *time.Time
	if title != "" {
		titlePtr = &title
	}
	if desc := r.FormValue("description"); desc != "" {
		descPtr = &desc
	}
	if fs := r.FormValue("film_stock"); fs != "" {
		filmStockPtr = &fs
	}
	if cam := r.FormValue("camera"); cam != "" {
		cameraPtr = &cam
	}
	if l := r.FormValue("lens"); l != "" {
		lensPtr = &l
	}
	if ta := r.FormValue("taken_at"); ta != "" {
		if t, err := time.Parse(time.RFC3339, ta); err == nil {
			takenAtPtr = &t
		}
	}

	// Insert into database
	var photo models.Photo
	err = s.DB.QueryRow(`
		INSERT INTO photos (id, title, description, slug, film_stock, camera, lens, taken_at,
			original_key, variants, width, height, file_size, blur_hash, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, title, description, slug, film_stock, camera, lens, taken_at,
			published, width, height, file_size, blur_hash, sort_order, created_at, updated_at`,
		photoID, titlePtr, descPtr, photoSlug, filmStockPtr, cameraPtr, lensPtr, takenAtPtr,
		originalKey, variantsJSON, origWidth, origHeight, len(data), blurHash, nextOrder,
	).Scan(
		&photo.ID, &photo.Title, &photo.Description, &photo.Slug,
		&photo.FilmStock, &photo.Camera, &photo.Lens, &photo.TakenAt,
		&photo.Published, &photo.Width, &photo.Height, &photo.FileSize,
		&photo.BlurHash, &photo.SortOrder, &photo.CreatedAt, &photo.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to save photo")
		return
	}

	// Populate URLs
	photo.URLs = make(map[string]string)
	for name, key := range variantKeys {
		url, _ := s.Storage.URL(ctx, key)
		photo.URLs[name] = url
	}

	JSON(w, http.StatusCreated, photo)
}

func (s *Server) ensureUniqueSlug(ctx context.Context, table, base string) (string, error) {
	candidate := base
	for i := 2; ; i++ {
		var exists bool
		err := s.DB.QueryRowContext(ctx,
			fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE slug = $1)", table),
			candidate,
		).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
}

func (s *Server) handleUpdatePhoto(w http.ResponseWriter, r *http.Request) {
	photoID := chi.URLParam(r, "id")

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Build dynamic UPDATE query
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	fields := map[string]string{
		"title": "title", "description": "description",
		"film_stock": "film_stock", "camera": "camera", "lens": "lens",
	}

	for jsonKey, dbCol := range fields {
		if val, ok := req[jsonKey]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", dbCol, argIdx))
			args = append(args, val)
			argIdx++
		}
	}

	if val, ok := req["published"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("published = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if val, ok := req["taken_at"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("taken_at = $%d", argIdx))
		if val == nil {
			args = append(args, nil)
		} else {
			args = append(args, val)
		}
		argIdx++
	}

	if len(setClauses) == 0 {
		Error(w, http.StatusBadRequest, "no fields to update")
		return
	}

	args = append(args, photoID)
	query := fmt.Sprintf(
		"UPDATE photos SET %s WHERE id = $%d RETURNING id, title, description, slug, film_stock, camera, lens, taken_at, published, width, height, file_size, blur_hash, sort_order, created_at, updated_at",
		strings.Join(setClauses, ", "), argIdx,
	)

	var photo models.Photo
	err := s.DB.QueryRow(query, args...).Scan(
		&photo.ID, &photo.Title, &photo.Description, &photo.Slug,
		&photo.FilmStock, &photo.Camera, &photo.Lens, &photo.TakenAt,
		&photo.Published, &photo.Width, &photo.Height, &photo.FileSize,
		&photo.BlurHash, &photo.SortOrder, &photo.CreatedAt, &photo.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "photo not found")
		return
	}

	// Populate URLs
	photo.URLs = s.getPhotoURLs(r.Context(), photo.ID)

	JSON(w, http.StatusOK, photo)
}

func (s *Server) handleDeletePhoto(w http.ResponseWriter, r *http.Request) {
	photoID := chi.URLParam(r, "id")

	// Get storage keys before deleting from DB
	var originalKey string
	var variantsJSON []byte
	err := s.DB.QueryRow(
		"SELECT original_key, variants FROM photos WHERE id = $1", photoID,
	).Scan(&originalKey, &variantsJSON)
	if err != nil {
		Error(w, http.StatusNotFound, "photo not found")
		return
	}

	// Delete from DB (cascades to collection_photos)
	_, err = s.DB.Exec("DELETE FROM photos WHERE id = $1", photoID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to delete photo")
		return
	}

	// Delete files from storage (best effort)
	ctx := r.Context()
	s.Storage.Delete(ctx, originalKey)
	var variants models.PhotoVariants
	json.Unmarshal(variantsJSON, &variants)
	for _, key := range variants {
		s.Storage.Delete(ctx, key)
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleReorderPhotos(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Orders []struct {
			ID        string `json:"id"`
			SortOrder int    `json:"sort_order"`
		} `json:"orders"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := s.DB.Begin()
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer tx.Rollback()

	for _, o := range req.Orders {
		_, err := tx.Exec("UPDATE photos SET sort_order = $1 WHERE id = $2", o.SortOrder, o.ID)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to update order")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		Error(w, http.StatusInternalServerError, "failed to commit reorder")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}

// getPhotoURLs loads variant keys from DB and resolves storage URLs.
func (s *Server) getPhotoURLs(ctx context.Context, photoID string) map[string]string {
	var variantsJSON []byte
	s.DB.QueryRow("SELECT variants FROM photos WHERE id = $1", photoID).Scan(&variantsJSON)

	var variants models.PhotoVariants
	json.Unmarshal(variantsJSON, &variants)

	urls := make(map[string]string)
	for name, key := range variants {
		url, _ := s.Storage.URL(ctx, key)
		urls[name] = url
	}
	return urls
}
