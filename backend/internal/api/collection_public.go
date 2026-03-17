package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/NeoRecasata/film-gallery/backend/internal/models"
)

func (s *Server) handleListCollections(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`
		SELECT c.id, c.title, c.slug, c.description, c.cover_photo, c.sort_order,
			c.created_at, c.updated_at,
			(SELECT COUNT(*) FROM collection_photos cp2
			 JOIN photos p2 ON p2.id = cp2.photo_id
			 JOIN rolls r2 ON r2.id = p2.roll_id
			 WHERE cp2.collection_id = c.id AND r2.published = true AND p2.hidden = false) AS photo_count
		FROM collections c
		ORDER BY c.sort_order ASC, c.created_at DESC`)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	collections := []models.Collection{}
	ctx := r.Context()

	for rows.Next() {
		var c models.Collection
		err := rows.Scan(
			&c.ID, &c.Title, &c.Slug, &c.Description, &c.CoverPhoto,
			&c.SortOrder, &c.CreatedAt, &c.UpdatedAt, &c.PhotoCount,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to scan collection")
			return
		}

		// Resolve cover photo URL
		if c.CoverPhoto != nil {
			urls := s.getPhotoURLs(ctx, *c.CoverPhoto)
			if thumb, ok := urls["thumb"]; ok {
				c.CoverURL = &thumb
			}
		}

		collections = append(collections, c)
	}

	JSON(w, http.StatusOK, collections)
}

func (s *Server) handleGetCollection(w http.ResponseWriter, r *http.Request) {
	collSlug := chi.URLParam(r, "slug")

	var c models.Collection
	err := s.DB.QueryRow(`
		SELECT id, title, slug, description, cover_photo, sort_order, created_at, updated_at
		FROM collections WHERE slug = $1`, collSlug,
	).Scan(
		&c.ID, &c.Title, &c.Slug, &c.Description, &c.CoverPhoto,
		&c.SortOrder, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "collection not found")
		return
	}

	// Load visible photos in this collection with roll metadata for inheritance
	rows, err := s.DB.Query(`
		SELECT p.id, p.title, p.description, p.slug, p.film_stock, p.camera, p.lens,
			p.location, p.taken_at, p.roll_id, p.hidden, p.variants, p.width, p.height,
			p.file_size, p.blur_hash, p.sort_order, p.created_at, p.updated_at,
			r.camera AS roll_camera, r.film_stock AS roll_film_stock, r.lens AS roll_lens,
			r.location AS roll_location, r.shot_at AS roll_shot_at,
			r.slug AS roll_slug, r.title AS roll_title
		FROM photos p
		JOIN collection_photos cp ON cp.photo_id = p.id
		JOIN rolls r ON r.id = p.roll_id
		WHERE cp.collection_id = $1 AND r.published = true AND p.hidden = false
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
		var rollCamera, rollFilmStock, rollLens, rollLocation, rollSlug, rollTitle *string
		var rollShotAt *time.Time
		err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.Slug,
			&p.FilmStock, &p.Camera, &p.Lens, &p.Location, &p.TakenAt,
			&p.RollID, &p.Hidden, &variantsJSON, &p.Width, &p.Height,
			&p.FileSize, &p.BlurHash, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
			&rollCamera, &rollFilmStock, &rollLens, &rollLocation, &rollShotAt,
			&rollSlug, &rollTitle,
		)
		if err != nil {
			continue
		}
		// Resolve metadata: photo-level overrides roll-level
		if p.Camera == nil { p.Camera = rollCamera }
		if p.FilmStock == nil { p.FilmStock = rollFilmStock }
		if p.Lens == nil { p.Lens = rollLens }
		if p.Location == nil { p.Location = rollLocation }
		if p.TakenAt == nil { p.TakenAt = rollShotAt }
		if rollSlug != nil { p.RollSlug = *rollSlug }
		if rollTitle != nil { p.RollTitle = *rollTitle }

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
