package api

import (
	"net/http"

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

	// Load visible photos in this collection
	rows, err := s.DB.Query(`
		SELECT p.id, p.title, p.description, p.slug, p.film_stock, p.camera, p.lens,
			p.location, p.taken_at, p.roll_id, p.hidden, p.variants, p.width, p.height,
			p.file_size, p.blur_hash, p.sort_order, p.created_at, p.updated_at
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
		c.Photos = append(c.Photos, p)
	}

	c.PhotoCount = len(c.Photos)
	JSON(w, http.StatusOK, c)
}
