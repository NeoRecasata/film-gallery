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
		r.Get("/photos", s.handleListPublicPhotos)
		r.Get("/photos/{slug}", s.handleGetPublicPhoto)
		r.Get("/collections", s.handleListPublicCollections)
		r.Get("/collections/{slug}", s.handleGetPublicCollection)
	})

	// Auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/login", s.handleLogin)
		r.Post("/refresh", s.handleRefresh)
		r.Post("/logout", s.handleLogout)
	})

	// Admin routes (protected)
	r.Route("/api/admin", func(r chi.Router) {
		r.Use(auth.RequireAuth(s.JWT))

		// Photos
		r.Get("/photos", s.handleAdminListPhotos)
		r.Post("/photos", s.handleAdminUploadPhoto)
		r.Get("/photos/{id}", s.handleAdminGetPhoto)
		r.Patch("/photos/{id}", s.handleAdminUpdatePhoto)
		r.Delete("/photos/{id}", s.handleAdminDeletePhoto)

		// Collections
		r.Get("/collections", s.handleAdminListCollections)
		r.Post("/collections", s.handleAdminCreateCollection)
		r.Get("/collections/{id}", s.handleAdminGetCollection)
		r.Patch("/collections/{id}", s.handleAdminUpdateCollection)
		r.Delete("/collections/{id}", s.handleAdminDeleteCollection)

		// Site settings
		r.Get("/settings", s.handleAdminGetSettings)
		r.Patch("/settings", s.handleAdminUpdateSettings)
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

func (s *Server) handleListPublicPhotos(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleGetPublicPhoto(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleListPublicCollections(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleGetPublicCollection(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

// --- Auth handlers ---

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

// --- Admin photo handlers ---

func (s *Server) handleAdminListPhotos(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminUploadPhoto(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminGetPhoto(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminUpdatePhoto(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminDeletePhoto(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

// --- Admin collection handlers ---

func (s *Server) handleAdminListCollections(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminCreateCollection(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminGetCollection(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminUpdateCollection(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminDeleteCollection(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

// --- Admin settings handlers ---

func (s *Server) handleAdminGetSettings(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}

func (s *Server) handleAdminUpdateSettings(w http.ResponseWriter, r *http.Request) {
	Error(w, http.StatusNotImplemented, "not implemented")
}
