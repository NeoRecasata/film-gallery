package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/NeoRecasata/film-gallery/backend/internal/models"
)

func (s *Server) handleListPublicRolls(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`
		SELECT r.id, r.title, r.slug, r.description, r.camera, r.film_stock, r.lens,
			r.location, r.shot_at, r.published, r.cover_photo_id, r.sort_order,
			r.created_at, r.updated_at,
			COUNT(p.id) FILTER (WHERE p.hidden = false) AS photo_count,
			cp.variants AS cover_variants
		FROM rolls r
		LEFT JOIN photos p ON p.roll_id = r.id
		LEFT JOIN photos cp ON cp.id = r.cover_photo_id
		WHERE r.published = true
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

func (s *Server) handleGetPublicRoll(w http.ResponseWriter, r *http.Request) {
	rollSlug := chi.URLParam(r, "slug")

	var roll models.Roll
	err := s.DB.QueryRow(`
		SELECT id, title, slug, description, camera, film_stock, lens, location, shot_at,
			published, cover_photo_id, sort_order, created_at, updated_at
		FROM rolls WHERE slug = $1 AND published = true`, rollSlug,
	).Scan(
		&roll.ID, &roll.Title, &roll.Slug, &roll.Description,
		&roll.Camera, &roll.FilmStock, &roll.Lens, &roll.Location, &roll.ShotAt,
		&roll.Published, &roll.CoverPhotoID, &roll.SortOrder, &roll.CreatedAt, &roll.UpdatedAt,
	)
	if err != nil {
		Error(w, http.StatusNotFound, "roll not found")
		return
	}

	// Load visible photos (hidden = false)
	rows, err := s.DB.Query(`
		SELECT id, title, description, slug, film_stock, camera, lens, location, taken_at,
			roll_id, hidden, featured, variants, width, height, file_size, blur_hash,
			sort_order, created_at, updated_at
		FROM photos WHERE roll_id = $1 AND hidden = false
		ORDER BY sort_order ASC, created_at DESC`, roll.ID)
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
			log.Printf("WARNING: failed to scan photo row in roll %s: %v", roll.ID, err)
			continue
		}

		// Resolve metadata: inherit from roll if photo doesn't have its own
		if p.Camera == nil && roll.Camera != nil {
			p.Camera = roll.Camera
		}
		if p.FilmStock == nil && roll.FilmStock != nil {
			p.FilmStock = roll.FilmStock
		}
		if p.Lens == nil && roll.Lens != nil {
			p.Lens = roll.Lens
		}
		if p.Location == nil && roll.Location != nil {
			p.Location = roll.Location
		}
		if p.TakenAt == nil && roll.ShotAt != nil {
			p.TakenAt = roll.ShotAt
		}

		var variants models.PhotoVariants
		variants.Scan(variantsJSON)
		p.URLs = make(map[string]string)
		for name, key := range variants {
			url, _ := s.Storage.URL(ctx, key)
			p.URLs[name] = url
		}

		p.RollSlug = roll.Slug
		p.RollTitle = roll.Title

		roll.Photos = append(roll.Photos, p)
	}

	roll.PhotoCount = len(roll.Photos)
	JSON(w, http.StatusOK, roll)
}
