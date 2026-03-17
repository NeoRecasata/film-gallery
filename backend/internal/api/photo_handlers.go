package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/NeoRecasata/film-gallery/backend/internal/models"
)

const maxUploadSize = 100 << 20 // 100MB

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
		"location": "location",
	}

	for jsonKey, dbCol := range fields {
		if val, ok := req[jsonKey]; ok {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", dbCol, argIdx))
			args = append(args, val)
			argIdx++
		}
	}

	if val, ok := req["hidden"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("hidden = $%d", argIdx))
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
		"UPDATE photos SET %s WHERE id = $%d RETURNING id, title, description, slug, film_stock, camera, lens, location, taken_at, roll_id, hidden, width, height, file_size, blur_hash, sort_order, created_at, updated_at",
		strings.Join(setClauses, ", "), argIdx,
	)

	var photo models.Photo
	err := s.DB.QueryRow(query, args...).Scan(
		&photo.ID, &photo.Title, &photo.Description, &photo.Slug,
		&photo.FilmStock, &photo.Camera, &photo.Lens, &photo.Location, &photo.TakenAt,
		&photo.RollID, &photo.Hidden, &photo.Width, &photo.Height,
		&photo.FileSize, &photo.BlurHash, &photo.SortOrder, &photo.CreatedAt, &photo.UpdatedAt,
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
