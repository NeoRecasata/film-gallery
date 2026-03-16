package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/NeoRecasata/film-gallery/backend/internal/auth"
	"github.com/NeoRecasata/film-gallery/backend/internal/config"
	"github.com/NeoRecasata/film-gallery/backend/internal/media"
	"github.com/NeoRecasata/film-gallery/backend/internal/models"
	"github.com/NeoRecasata/film-gallery/backend/internal/storage"
)

// Server holds all shared dependencies for HTTP handlers.
type Server struct {
	DB        *sql.DB
	Config    *config.Config
	Storage   storage.Storage
	JWT       *auth.JWTService
	Processor *media.Processor
}

// NewRouter builds the chi router with all middleware and route registrations.
func NewRouter(s *Server) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		r.Get("/photos", s.handleListPhotos)
		r.Get("/photos/{slug}", s.handleGetPhoto)
		r.Get("/collections", s.handleListCollections)
		r.Get("/collections/{slug}", s.handleGetCollection)
		r.Get("/site", s.handleGetSiteSettings)
	})

	// Auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/setup", s.handleSetup)
		r.Post("/login", s.handleLogin)
		r.Post("/refresh", s.handleRefresh)
		r.Post("/logout", s.handleLogout)

		// Protected auth routes
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(s.JWT))
			r.Post("/change-password", s.handleChangePassword)
		})
	})

	// Admin routes (protected)
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(auth.RequireAuth(s.JWT))

		// Photos
		r.Get("/photos", s.handleAdminListPhotos)
		r.Post("/photos", s.handleUploadPhoto)
		r.Get("/photos/{id}", s.handleAdminGetPhoto)
		r.Patch("/photos/{id}", s.handleUpdatePhoto)
		r.Delete("/photos/{id}", s.handleDeletePhoto)
		r.Post("/photos/reorder", s.handleReorderPhotos)

		// Collections
		r.Post("/collections", s.handleCreateCollection)
		r.Patch("/collections/{id}", s.handleUpdateCollection)
		r.Delete("/collections/{id}", s.handleDeleteCollection)
		r.Put("/collections/{id}/photos", s.handleSetCollectionPhotos)

		// Site settings
		r.Get("/settings", s.handleGetSiteSettings)
		r.Patch("/settings", s.handleUpdateSiteSettings)
	})

	// Local file server for /media/* when using local storage
	if s.Config.StorageType == "local" {
		localPath := s.Config.StorageLocalPath
		fs := http.StripPrefix("/media/", http.FileServer(http.Dir(localPath)))
		r.Get("/media/*", func(w http.ResponseWriter, r *http.Request) {
			fs.ServeHTTP(w, r)
		})
	}

	return r
}

// --- Public handlers ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Admin photo handlers ---

func (s *Server) handleAdminListPhotos(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query(`SELECT id, title, description, slug, film_stock, camera, lens, taken_at,
		published, variants, width, height, file_size, blur_hash, sort_order, created_at, updated_at
		FROM photos ORDER BY sort_order ASC, created_at DESC`)
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

	JSON(w, http.StatusOK, photosResponse{Data: photos})
}

func (s *Server) handleAdminGetPhoto(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}
