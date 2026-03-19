package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/NeoRecasata/film-gallery/backend/internal/models"
)

type photosResponse struct {
	Data       []models.Photo `json:"data"`
	NextCursor *string        `json:"next_cursor"`
}

func (s *Server) handleListPhotos(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	query := `SELECT p.id, p.title, p.description, p.slug, p.film_stock, p.camera, p.lens,
		p.location, p.taken_at, p.roll_id, p.hidden, p.featured, p.variants, p.width, p.height,
		p.file_size, p.blur_hash, p.sort_order, p.created_at, p.updated_at,
		r.slug AS roll_slug, r.title AS roll_title,
		r.camera AS roll_camera, r.film_stock AS roll_film_stock,
		r.lens AS roll_lens, r.location AS roll_location, r.shot_at AS roll_shot_at
		FROM photos p
		JOIN rolls r ON r.id = p.roll_id
		WHERE r.published = true AND p.hidden = false`
	args := []interface{}{}
	argIdx := 1

	// Filters
	if collection := r.URL.Query().Get("collection"); collection != "" {
		query += fmt.Sprintf(` AND p.id IN (
			SELECT photo_id FROM collection_photos cp
			JOIN collections c ON c.id = cp.collection_id
			WHERE c.slug = $%d)`, argIdx)
		args = append(args, collection)
		argIdx++
	}
	if fs := r.URL.Query().Get("film_stock"); fs != "" {
		query += fmt.Sprintf(" AND (p.film_stock = $%d OR (p.film_stock IS NULL AND r.film_stock = $%d))", argIdx, argIdx)
		args = append(args, fs)
		argIdx++
	}
	if cam := r.URL.Query().Get("camera"); cam != "" {
		query += fmt.Sprintf(" AND (p.camera = $%d OR (p.camera IS NULL AND r.camera = $%d))", argIdx, argIdx)
		args = append(args, cam)
		argIdx++
	}

	if r.URL.Query().Get("featured") == "true" {
		query += " AND p.featured = true"
	}

	// Cursor pagination: cursor is a created_at timestamp
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		query += fmt.Sprintf(" AND p.created_at < $%d", argIdx)
		args = append(args, cursor)
		argIdx++
	}

	query += " ORDER BY p.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit+1) // fetch one extra to detect next page

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	photos := []models.Photo{}
	ctx := r.Context()

	for rows.Next() {
		var p models.Photo
		var variantsJSON []byte
		var rollSlug, rollTitle string
		var rollCamera, rollFilmStock, rollLens, rollLocation *string
		var rollShotAt *time.Time
		err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Slug,
			&p.FilmStock, &p.Camera, &p.Lens, &p.Location, &p.TakenAt,
			&p.RollID, &p.Hidden, &p.Featured, &variantsJSON, &p.Width, &p.Height,
			&p.FileSize, &p.BlurHash, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
			&rollSlug, &rollTitle,
			&rollCamera, &rollFilmStock, &rollLens, &rollLocation, &rollShotAt,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to scan photo")
			return
		}

		// Resolve metadata: photo-level overrides roll-level
		if p.Camera == nil && rollCamera != nil {
			p.Camera = rollCamera
		}
		if p.FilmStock == nil && rollFilmStock != nil {
			p.FilmStock = rollFilmStock
		}
		if p.Lens == nil && rollLens != nil {
			p.Lens = rollLens
		}
		if p.Location == nil && rollLocation != nil {
			p.Location = rollLocation
		}
		if p.TakenAt == nil && rollShotAt != nil {
			p.TakenAt = rollShotAt
		}

		p.RollSlug = rollSlug
		p.RollTitle = rollTitle

		var variants models.PhotoVariants
		variants.Scan(variantsJSON)
		p.URLs = make(map[string]string)
		for name, key := range variants {
			url, _ := s.Storage.URL(ctx, key)
			p.URLs[name] = url
		}

		photos = append(photos, p)
	}

	resp := photosResponse{Data: photos}

	if len(photos) > limit {
		resp.Data = photos[:limit]
		last := photos[limit-1]
		cursor := last.CreatedAt.Format(time.RFC3339Nano)
		resp.NextCursor = &cursor
	}

	JSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetPhoto(w http.ResponseWriter, r *http.Request) {
	photoSlug := chi.URLParam(r, "slug")

	var p models.Photo
	var variantsJSON []byte
	var rollSlug, rollTitle string
	var rollCamera, rollFilmStock, rollLens, rollLocation *string
	var rollShotAt *time.Time
	err := s.DB.QueryRow(`
		SELECT p.id, p.title, p.description, p.slug, p.film_stock, p.camera, p.lens,
			p.location, p.taken_at, p.roll_id, p.hidden, p.featured, p.variants, p.width, p.height,
			p.file_size, p.blur_hash, p.sort_order, p.created_at, p.updated_at,
			r.slug AS roll_slug, r.title AS roll_title,
			r.camera AS roll_camera, r.film_stock AS roll_film_stock,
			r.lens AS roll_lens, r.location AS roll_location, r.shot_at AS roll_shot_at
		FROM photos p
		JOIN rolls r ON r.id = p.roll_id
		WHERE p.slug = $1 AND r.published = true AND p.hidden = false`, photoSlug,
	).Scan(
		&p.ID, &p.Title, &p.Description, &p.Slug,
		&p.FilmStock, &p.Camera, &p.Lens, &p.Location, &p.TakenAt,
		&p.RollID, &p.Hidden, &p.Featured, &variantsJSON, &p.Width, &p.Height,
		&p.FileSize, &p.BlurHash, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
		&rollSlug, &rollTitle,
		&rollCamera, &rollFilmStock, &rollLens, &rollLocation, &rollShotAt,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "photo not found")
		return
	}

	// Resolve metadata: photo-level overrides roll-level
	if p.Camera == nil && rollCamera != nil {
		p.Camera = rollCamera
	}
	if p.FilmStock == nil && rollFilmStock != nil {
		p.FilmStock = rollFilmStock
	}
	if p.Lens == nil && rollLens != nil {
		p.Lens = rollLens
	}
	if p.Location == nil && rollLocation != nil {
		p.Location = rollLocation
	}
	if p.TakenAt == nil && rollShotAt != nil {
		p.TakenAt = rollShotAt
	}

	p.RollSlug = rollSlug
	p.RollTitle = rollTitle

	var variants models.PhotoVariants
	variants.Scan(variantsJSON)
	p.URLs = make(map[string]string)
	ctx := r.Context()
	for name, key := range variants {
		url, _ := s.Storage.URL(ctx, key)
		p.URLs[name] = url
	}

	// Resolve prev/next photo slugs (global feed order: created_at DESC)
	var prevSlug, nextSlug *string
	s.DB.QueryRow(`
		SELECT p2.slug FROM photos p2
		JOIN rolls r2 ON r2.id = p2.roll_id
		WHERE r2.published = true AND p2.hidden = false AND p2.created_at > $1
		ORDER BY p2.created_at ASC LIMIT 1`, p.CreatedAt,
	).Scan(&prevSlug)
	s.DB.QueryRow(`
		SELECT p2.slug FROM photos p2
		JOIN rolls r2 ON r2.id = p2.roll_id
		WHERE r2.published = true AND p2.hidden = false AND p2.created_at < $1
		ORDER BY p2.created_at DESC LIMIT 1`, p.CreatedAt,
	).Scan(&nextSlug)
	p.PrevSlug = prevSlug
	p.NextSlug = nextSlug

	JSON(w, http.StatusOK, p)
}
