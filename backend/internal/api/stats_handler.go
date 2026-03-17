package api

import (
	"net/http"
)

func (s *Server) handleAdminStats(w http.ResponseWriter, r *http.Request) {
	var rollCount, photoCount, collectionCount int
	var storageBytes int64

	s.DB.QueryRow("SELECT COUNT(*) FROM rolls").Scan(&rollCount)
	s.DB.QueryRow("SELECT COUNT(*) FROM photos").Scan(&photoCount)
	s.DB.QueryRow("SELECT COUNT(*) FROM collections").Scan(&collectionCount)
	s.DB.QueryRow("SELECT COALESCE(SUM(file_size), 0) FROM photos").Scan(&storageBytes)

	JSON(w, http.StatusOK, map[string]interface{}{
		"roll_count":       rollCount,
		"photo_count":      photoCount,
		"collection_count": collectionCount,
		"storage_bytes":    storageBytes,
	})
}
