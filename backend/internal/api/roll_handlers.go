package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lib/pq"

	"github.com/NeoRecasata/film-gallery/backend/internal/models"
	slugpkg "github.com/NeoRecasata/film-gallery/backend/internal/slug"
)

func (s *Server) handleCreateRoll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string     `json:"title"`
		Description *string    `json:"description"`
		Camera      *string    `json:"camera"`
		FilmStock   *string    `json:"film_stock"`
		Lens        *string    `json:"lens"`
		Location    *string    `json:"location"`
		ShotAt      *time.Time `json:"shot_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		Error(w, http.StatusBadRequest, "title is required")
		return
	}

	rollSlug := slugpkg.Generate(req.Title, "roll")
	rollSlug, err := s.ensureUniqueSlug(r.Context(), "rolls", rollSlug)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate slug")
		return
	}

	var nextOrder int
	s.DB.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM rolls").Scan(&nextOrder)

	var roll models.Roll
	err = s.DB.QueryRow(`
		INSERT INTO rolls (title, slug, description, camera, film_stock, lens, location, shot_at, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, title, slug, description, camera, film_stock, lens, location, shot_at,
			published, cover_photo_id, sort_order, created_at, updated_at`,
		req.Title, rollSlug, req.Description, req.Camera, req.FilmStock, req.Lens,
		req.Location, req.ShotAt, nextOrder,
	).Scan(
		&roll.ID, &roll.Title, &roll.Slug, &roll.Description,
		&roll.Camera, &roll.FilmStock, &roll.Lens, &roll.Location, &roll.ShotAt,
		&roll.Published, &roll.CoverPhotoID, &roll.SortOrder, &roll.CreatedAt, &roll.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create roll")
		return
	}

	JSON(w, http.StatusCreated, roll)
}

func (s *Server) handleListRolls(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`
		SELECT r.id, r.title, r.slug, r.description, r.camera, r.film_stock, r.lens,
			r.location, r.shot_at, r.published, r.cover_photo_id, r.sort_order,
			r.created_at, r.updated_at,
			COUNT(p.id) AS photo_count,
			cp.variants AS cover_variants
		FROM rolls r
		LEFT JOIN photos p ON p.roll_id = r.id
		LEFT JOIN photos cp ON cp.id = r.cover_photo_id
		GROUP BY r.id, cp.variants
		ORDER BY r.sort_order ASC, r.created_at DESC`)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	rolls := []models.Roll{}
	ctx := r.Context()

	for rows.Next() {
		var roll models.Roll
		var coverVariantsJSON []byte
		err := rows.Scan(
			&roll.ID, &roll.Title, &roll.Slug, &roll.Description,
			&roll.Camera, &roll.FilmStock, &roll.Lens, &roll.Location, &roll.ShotAt,
			&roll.Published, &roll.CoverPhotoID, &roll.SortOrder,
			&roll.CreatedAt, &roll.UpdatedAt,
			&roll.PhotoCount,
			&coverVariantsJSON,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to scan roll")
			return
		}

		if coverVariantsJSON != nil {
			var variants models.PhotoVariants
			json.Unmarshal(coverVariantsJSON, &variants)
			if thumbKey, ok := variants["thumb"]; ok {
				if url, err := s.Storage.URL(ctx, thumbKey); err == nil && url != "" {
					roll.CoverURL = &url
				}
			}
		}

		rolls = append(rolls, roll)
	}

	JSON(w, http.StatusOK, rolls)
}


func (s *Server) handleGetRoll(w http.ResponseWriter, r *http.Request) {
	rollID := chi.URLParam(r, "id")

	var roll models.Roll
	err := s.DB.QueryRow(`
		SELECT id, title, slug, description, camera, film_stock, lens, location, shot_at,
			published, cover_photo_id, sort_order, created_at, updated_at
		FROM rolls WHERE id = $1`, rollID,
	).Scan(
		&roll.ID, &roll.Title, &roll.Slug, &roll.Description,
		&roll.Camera, &roll.FilmStock, &roll.Lens, &roll.Location, &roll.ShotAt,
		&roll.Published, &roll.CoverPhotoID, &roll.SortOrder, &roll.CreatedAt, &roll.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "roll not found")
		return
	}

	// Load all photos for this roll (including hidden)
	rows, err := s.DB.Query(`
		SELECT id, title, description, slug, film_stock, camera, lens, location, taken_at,
			roll_id, hidden, featured, variants, width, height, file_size, blur_hash,
			sort_order, created_at, updated_at
		FROM photos WHERE roll_id = $1
		ORDER BY sort_order ASC, created_at DESC`, rollID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	ctx := r.Context()
	roll.Photos = []models.Photo{}

	for rows.Next() {
		var p models.Photo
		var variantsJSON []byte
		err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Slug,
			&p.FilmStock, &p.Camera, &p.Lens, &p.Location, &p.TakenAt,
			&p.RollID, &p.Hidden, &p.Featured, &variantsJSON, &p.Width, &p.Height,
			&p.FileSize, &p.BlurHash, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			log.Printf("WARNING: failed to scan photo row in roll %s: %v", rollID, err)
			continue
		}
		var variants models.PhotoVariants
		variants.Scan(variantsJSON)
		p.URLs = make(map[string]string)
		for name, key := range variants {
			url, _ := s.Storage.URL(ctx, key)
			p.URLs[name] = url
		}
		roll.Photos = append(roll.Photos, p)
	}

	roll.PhotoCount = len(roll.Photos)
	JSON(w, http.StatusOK, roll)
}

func (s *Server) handleUpdateRoll(w http.ResponseWriter, r *http.Request) {
	rollID := chi.URLParam(r, "id")

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
		"title": "title", "slug": "slug", "description": "description",
		"camera": "camera", "film_stock": "film_stock", "lens": "lens",
		"location": "location",
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

	if val, ok := req["cover_photo_id"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("cover_photo_id = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if val, ok := req["shot_at"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("shot_at = $%d", argIdx))
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

	args = append(args, rollID)
	query := fmt.Sprintf(
		`UPDATE rolls SET %s WHERE id = $%d
		RETURNING id, title, slug, description, camera, film_stock, lens, location, shot_at,
			published, cover_photo_id, sort_order, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var roll models.Roll
	err := s.DB.QueryRow(query, args...).Scan(
		&roll.ID, &roll.Title, &roll.Slug, &roll.Description,
		&roll.Camera, &roll.FilmStock, &roll.Lens, &roll.Location, &roll.ShotAt,
		&roll.Published, &roll.CoverPhotoID, &roll.SortOrder, &roll.CreatedAt, &roll.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			Error(w, http.StatusConflict, "a roll with that slug already exists")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			Error(w, http.StatusNotFound, "roll not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to update roll")
		return
	}

	JSON(w, http.StatusOK, roll)
}

func (s *Server) handleDeleteRoll(w http.ResponseWriter, r *http.Request) {
	rollID := chi.URLParam(r, "id")

	// Get all photo storage keys before deleting
	rows, err := s.DB.Query(
		"SELECT original_key, variants FROM photos WHERE roll_id = $1", rollID,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}

	type photoKeys struct {
		originalKey  string
		variantsJSON []byte
	}
	var allKeys []photoKeys
	for rows.Next() {
		var pk photoKeys
		rows.Scan(&pk.originalKey, &pk.variantsJSON)
		allKeys = append(allKeys, pk)
	}
	rows.Close()

	// Delete roll (cascades to photos → collection_photos)
	result, err := s.DB.Exec("DELETE FROM rolls WHERE id = $1", rollID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to delete roll")
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		Error(w, http.StatusNotFound, "roll not found")
		return
	}

	// Best-effort file cleanup from storage
	ctx := r.Context()
	for _, pk := range allKeys {
		s.Storage.Delete(ctx, pk.originalKey)
		var variants models.PhotoVariants
		json.Unmarshal(pk.variantsJSON, &variants)
		for _, key := range variants {
			s.Storage.Delete(ctx, key)
		}
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Upload & Reorder ---

func (s *Server) handleUploadRollPhotos(w http.ResponseWriter, r *http.Request) {
	rollID := chi.URLParam(r, "id")

	// Verify roll exists and get its slug
	var rollSlug string
	err := s.DB.QueryRow("SELECT slug FROM rolls WHERE id = $1", rollID).Scan(&rollSlug)
	if err != nil {
		Error(w, http.StatusNotFound, "roll not found")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		Error(w, http.StatusRequestEntityTooLarge, "upload too large (max 100MB)")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		Error(w, http.StatusBadRequest, "no files provided")
		return
	}

	type failedUpload struct {
		Filename string `json:"filename"`
		Error    string `json:"error"`
	}

	uploaded := []models.Photo{}
	failed := []failedUpload{}

	for _, fh := range files {
		photo, err := s.processAndStorePhoto(r.Context(), fh, rollID, rollSlug)
		if err != nil {
			failed = append(failed, failedUpload{Filename: fh.Filename, Error: err.Error()})
			continue
		}
		uploaded = append(uploaded, *photo)
	}

	JSON(w, http.StatusCreated, map[string]interface{}{
		"uploaded": uploaded,
		"failed":   failed,
	})
}

func (s *Server) processAndStorePhoto(ctx context.Context, fh *multipart.FileHeader, rollID, rollSlug string) (*models.Photo, error) {
	file, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file")
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file")
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".tiff" && ext != ".tif" {
		return nil, fmt.Errorf("unsupported file type (use JPEG, PNG, or TIFF)")
	}

	// Extract dimensions
	origWidth, origHeight, err := s.Processor.ExtractDimensions(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to read image dimensions")
	}

	// Generate variants
	variants, err := s.Processor.GenerateVariants(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to process image")
	}

	// Generate BlurHash from thumbnail
	blurHash, err := s.Processor.GenerateBlurHash(bytes.NewReader(variants["thumb"].Data))
	if err != nil {
		return nil, fmt.Errorf("failed to generate blur hash")
	}

	// Generate photo ID
	var photoID string
	err = s.DB.QueryRow("SELECT gen_random_uuid()").Scan(&photoID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID")
	}

	// Store original
	originalKey := fmt.Sprintf("photos/%s/original%s", photoID, ext)
	if err := s.Storage.Put(ctx, originalKey, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("failed to store original")
	}

	// Store variants
	variantKeys := make(models.PhotoVariants)
	for name, v := range variants {
		key := fmt.Sprintf("photos/%s/%s.webp", photoID, name)
		if err := s.Storage.Put(ctx, key, bytes.NewReader(v.Data)); err != nil {
			return nil, fmt.Errorf("failed to store %s variant", name)
		}
		variantKeys[name] = key
	}

	variantsJSON, err := json.Marshal(variantKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize variants")
	}

	// Generate slug and insert — retry on slug collision from concurrent uploads.
	// Slug generation (COUNT + ensureUniqueSlug) races with other goroutines, so
	// if the INSERT hits a unique constraint violation we re-generate and retry.
	var photo models.Photo
	for attempt := 0; attempt < 3; attempt++ {
		var photoCount int
		s.DB.QueryRow("SELECT COUNT(*) FROM photos WHERE roll_id = $1", rollID).Scan(&photoCount)
		photoSlug := fmt.Sprintf("%s-%d", rollSlug, photoCount+1)
		photoSlug, err = s.ensureUniqueSlug(ctx, "photos", photoSlug)
		if err != nil {
			return nil, fmt.Errorf("failed to generate slug")
		}

		var nextOrder int
		s.DB.QueryRow("SELECT COALESCE(MAX(sort_order), -1) + 1 FROM photos WHERE roll_id = $1", rollID).Scan(&nextOrder)

		err = s.DB.QueryRowContext(ctx, `
			INSERT INTO photos (id, slug, roll_id, original_key, variants, width, height,
				file_size, blur_hash, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, title, description, slug, film_stock, camera, lens, location, taken_at,
				roll_id, hidden, featured, width, height, file_size, blur_hash, sort_order, created_at, updated_at`,
			photoID, photoSlug, rollID, originalKey, variantsJSON,
			origWidth, origHeight, len(data), blurHash, nextOrder,
		).Scan(
			&photo.ID, &photo.Title, &photo.Description, &photo.Slug,
			&photo.FilmStock, &photo.Camera, &photo.Lens, &photo.Location, &photo.TakenAt,
			&photo.RollID, &photo.Hidden, &photo.Featured, &photo.Width, &photo.Height,
			&photo.FileSize, &photo.BlurHash, &photo.SortOrder, &photo.CreatedAt, &photo.UpdatedAt,
		)
		if err == nil {
			break
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			continue // slug collision, retry
		}
		return nil, fmt.Errorf("failed to save photo")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to save photo after retries")
	}

	// Populate URLs
	photo.URLs = make(map[string]string)
	for name, key := range variantKeys {
		url, _ := s.Storage.URL(ctx, key)
		photo.URLs[name] = url
	}

	return &photo, nil
}

func (s *Server) handleReorderRollPhotos(w http.ResponseWriter, r *http.Request) {
	rollID := chi.URLParam(r, "id")

	// Verify roll exists
	var exists bool
	s.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM rolls WHERE id = $1)", rollID).Scan(&exists)
	if !exists {
		Error(w, http.StatusNotFound, "roll not found")
		return
	}

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
		_, err := tx.Exec(
			"UPDATE photos SET sort_order = $1 WHERE id = $2 AND roll_id = $3",
			o.SortOrder, o.ID, rollID,
		)
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
