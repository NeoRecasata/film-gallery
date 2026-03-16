package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

	query := `SELECT id, title, description, slug, film_stock, camera, lens, taken_at,
		published, variants, width, height, file_size, blur_hash, sort_order, created_at, updated_at
		FROM photos WHERE published = true`
	args := []interface{}{}
	argIdx := 1

	// Filters
	if collection := r.URL.Query().Get("collection"); collection != "" {
		query += fmt.Sprintf(` AND id IN (
			SELECT photo_id FROM collection_photos cp
			JOIN collections c ON c.id = cp.collection_id
			WHERE c.slug = $%d)`, argIdx)
		args = append(args, collection)
		argIdx++
	}
	if fs := r.URL.Query().Get("film_stock"); fs != "" {
		query += fmt.Sprintf(" AND film_stock = $%d", argIdx)
		args = append(args, fs)
		argIdx++
	}
	if cam := r.URL.Query().Get("camera"); cam != "" {
		query += fmt.Sprintf(" AND camera = $%d", argIdx)
		args = append(args, cam)
		argIdx++
	}

	// Cursor pagination: cursor format is "sort_order:created_at"
	// Matches ORDER BY sort_order ASC, created_at DESC
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		parts := strings.SplitN(cursor, ":", 2)
		if len(parts) == 2 {
			cursorOrder, _ := strconv.Atoi(parts[0])
			cursorTime := parts[1]
			query += fmt.Sprintf(
				" AND (sort_order > $%d OR (sort_order = $%d AND created_at < $%d))",
				argIdx, argIdx+1, argIdx+2,
			)
			args = append(args, cursorOrder, cursorOrder, cursorTime)
			argIdx += 3
		}
	}

	query += " ORDER BY sort_order ASC, created_at DESC"
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
		err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Slug,
			&p.FilmStock, &p.Camera, &p.Lens, &p.TakenAt,
			&p.Published, &variantsJSON, &p.Width, &p.Height,
			&p.FileSize, &p.BlurHash, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to scan photo")
			return
		}

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
		cursor := fmt.Sprintf("%d:%s", last.SortOrder, last.CreatedAt.Format(time.RFC3339Nano))
		resp.NextCursor = &cursor
	}

	JSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetPhoto(w http.ResponseWriter, r *http.Request) {
	photoSlug := chi.URLParam(r, "slug")

	var p models.Photo
	var variantsJSON []byte
	err := s.DB.QueryRow(`
		SELECT id, title, description, slug, film_stock, camera, lens, taken_at,
			published, variants, width, height, file_size, blur_hash, sort_order, created_at, updated_at
		FROM photos WHERE slug = $1 AND published = true`, photoSlug,
	).Scan(
		&p.ID, &p.Title, &p.Description, &p.Slug,
		&p.FilmStock, &p.Camera, &p.Lens, &p.TakenAt,
		&p.Published, &variantsJSON, &p.Width, &p.Height,
		&p.FileSize, &p.BlurHash, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "photo not found")
		return
	}

	var variants models.PhotoVariants
	variants.Scan(variantsJSON)
	p.URLs = make(map[string]string)
	ctx := r.Context()
	for name, key := range variants {
		url, _ := s.Storage.URL(ctx, key)
		p.URLs[name] = url
	}

	JSON(w, http.StatusOK, p)
}
