package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/NeoRecasata/film-gallery/backend/internal/models"
	slugpkg "github.com/NeoRecasata/film-gallery/backend/internal/slug"
)

func (s *Server) handleGetAdminCollection(w http.ResponseWriter, r *http.Request) {
	collID := chi.URLParam(r, "id")

	var c models.Collection
	err := s.DB.QueryRow(`
		SELECT id, title, slug, description, cover_photo, sort_order, created_at, updated_at
		FROM collections WHERE id = $1`, collID,
	).Scan(
		&c.ID, &c.Title, &c.Slug, &c.Description, &c.CoverPhoto,
		&c.SortOrder, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "collection not found")
		return
	}

	// Load ALL photos in this collection (no visibility filter)
	rows, err := s.DB.Query(`
		SELECT p.id, p.title, p.description, p.slug, p.film_stock, p.camera, p.lens,
			p.location, p.taken_at, p.roll_id, p.hidden, p.featured, p.variants, p.width, p.height,
			p.file_size, p.blur_hash, cp.sort_order, p.created_at, p.updated_at,
			r.title AS roll_title
		FROM photos p
		JOIN collection_photos cp ON cp.photo_id = p.id
		JOIN rolls r ON r.id = p.roll_id
		WHERE cp.collection_id = $1
		ORDER BY cp.sort_order ASC`, c.ID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	ctx := r.Context()
	c.Photos = []models.Photo{}

	for rows.Next() {
		var p models.Photo
		var variantsJSON []byte
		err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Slug,
			&p.FilmStock, &p.Camera, &p.Lens, &p.Location, &p.TakenAt,
			&p.RollID, &p.Hidden, &p.Featured, &variantsJSON, &p.Width, &p.Height,
			&p.FileSize, &p.BlurHash, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
			&p.RollTitle,
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
		c.Photos = append(c.Photos, p)
	}

	c.PhotoCount = len(c.Photos)
	JSON(w, http.StatusOK, c)
}

func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string  `json:"title"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" {
		Error(w, http.StatusBadRequest, "title is required")
		return
	}

	collSlug := slugpkg.Generate(req.Title, "")
	collSlug, err := s.ensureUniqueSlug(r.Context(), "collections", collSlug)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate slug")
		return
	}

	var nextOrder int
	s.DB.QueryRow("SELECT COALESCE(MAX(sort_order), 0) + 1 FROM collections").Scan(&nextOrder)

	var coll models.Collection
	err = s.DB.QueryRow(`
		INSERT INTO collections (title, slug, description, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, slug, description, cover_photo, sort_order, created_at, updated_at`,
		req.Title, collSlug, req.Description, nextOrder,
	).Scan(
		&coll.ID, &coll.Title, &coll.Slug, &coll.Description,
		&coll.CoverPhoto, &coll.SortOrder, &coll.CreatedAt, &coll.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create collection")
		return
	}

	JSON(w, http.StatusCreated, coll)
}

func (s *Server) handleUpdateCollection(w http.ResponseWriter, r *http.Request) {
	collID := chi.URLParam(r, "id")

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if val, ok := req["title"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}
	if val, ok := req["description"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}
	if val, ok := req["cover_photo"]; ok {
		setClauses = append(setClauses, fmt.Sprintf("cover_photo = $%d", argIdx))
		args = append(args, val)
		argIdx++
	}

	if len(setClauses) == 0 {
		Error(w, http.StatusBadRequest, "no fields to update")
		return
	}

	args = append(args, collID)
	query := fmt.Sprintf(
		`UPDATE collections SET %s WHERE id = $%d
		RETURNING id, title, slug, description, cover_photo, sort_order, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var coll models.Collection
	err := s.DB.QueryRow(query, args...).Scan(
		&coll.ID, &coll.Title, &coll.Slug, &coll.Description,
		&coll.CoverPhoto, &coll.SortOrder, &coll.CreatedAt, &coll.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "collection not found")
		return
	}

	JSON(w, http.StatusOK, coll)
}

func (s *Server) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	collID := chi.URLParam(r, "id")

	result, err := s.DB.Exec("DELETE FROM collections WHERE id = $1", collID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to delete collection")
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		Error(w, http.StatusNotFound, "collection not found")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleSetCollectionPhotos(w http.ResponseWriter, r *http.Request) {
	collID := chi.URLParam(r, "id")

	var req struct {
		Photos []struct {
			PhotoID   string `json:"photo_id"`
			SortOrder int    `json:"sort_order"`
		} `json:"photos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Verify collection exists
	var exists bool
	s.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM collections WHERE id = $1)", collID).Scan(&exists)
	if !exists {
		Error(w, http.StatusNotFound, "collection not found")
		return
	}

	tx, err := s.DB.Begin()
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer tx.Rollback()

	// Clear existing
	tx.Exec("DELETE FROM collection_photos WHERE collection_id = $1", collID)

	// Insert new
	for _, p := range req.Photos {
		_, err := tx.Exec(
			"INSERT INTO collection_photos (collection_id, photo_id, sort_order) VALUES ($1, $2, $3)",
			collID, p.PhotoID, p.SortOrder,
		)
		if err != nil {
			Error(w, http.StatusBadRequest, "invalid photo ID")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		Error(w, http.StatusInternalServerError, "failed to update collection photos")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
