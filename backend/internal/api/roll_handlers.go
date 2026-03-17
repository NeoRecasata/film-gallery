package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

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
			COUNT(p.id) AS photo_count
		FROM rolls r
		LEFT JOIN photos p ON p.roll_id = r.id
		GROUP BY r.id
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
		err := rows.Scan(
			&roll.ID, &roll.Title, &roll.Slug, &roll.Description,
			&roll.Camera, &roll.FilmStock, &roll.Lens, &roll.Location, &roll.ShotAt,
			&roll.Published, &roll.CoverPhotoID, &roll.SortOrder,
			&roll.CreatedAt, &roll.UpdatedAt,
			&roll.PhotoCount,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to scan roll")
			return
		}

		if roll.CoverPhotoID != nil {
			url := s.getCoverURL(ctx, *roll.CoverPhotoID)
			if url != "" {
				roll.CoverURL = &url
			}
		}

		rolls = append(rolls, roll)
	}

	JSON(w, http.StatusOK, rolls)
}

// getCoverURL resolves the thumbnail URL for a cover photo.
func (s *Server) getCoverURL(ctx context.Context, photoID string) string {
	var variantsJSON []byte
	err := s.DB.QueryRow("SELECT variants FROM photos WHERE id = $1", photoID).Scan(&variantsJSON)
	if err != nil {
		return ""
	}

	var variants models.PhotoVariants
	json.Unmarshal(variantsJSON, &variants)

	if thumbKey, ok := variants["thumb"]; ok {
		url, _ := s.Storage.URL(ctx, thumbKey)
		return url
	}
	return ""
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
			roll_id, hidden, variants, width, height, file_size, blur_hash,
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
			&p.RollID, &p.Hidden, &variantsJSON, &p.Width, &p.Height,
			&p.FileSize, &p.BlurHash, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
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
		Error(w, http.StatusNotFound, "roll not found")
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
