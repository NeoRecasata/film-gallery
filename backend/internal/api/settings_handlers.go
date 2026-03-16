package api

import (
	"encoding/json"
	"net/http"

	"github.com/NeoRecasata/film-gallery/backend/internal/models"
)

func (s *Server) handleGetSiteSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query("SELECT key, value FROM site_settings")
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	settings := make(models.SiteSettings)
	for rows.Next() {
		var key string
		var value json.RawMessage
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		settings[key] = value
	}

	JSON(w, http.StatusOK, settings)
}

func (s *Server) handleUpdateSiteSettings(w http.ResponseWriter, r *http.Request) {
	var updates map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx, err := s.DB.Begin()
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	defer tx.Rollback()

	for key, value := range updates {
		_, err := tx.Exec(
			"INSERT INTO site_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2",
			key, value,
		)
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to update settings")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		Error(w, http.StatusInternalServerError, "failed to save settings")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
