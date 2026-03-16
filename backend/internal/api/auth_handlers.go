package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/NeoRecasata/film-gallery/backend/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	// Check if a user already exists
	var count int
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	if count > 0 {
		Error(w, http.StatusNotFound, "not found")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var userID string
	err = s.DB.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		req.Email, string(hash),
	).Scan(&userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	accessToken, err := s.JWT.GenerateAccessToken(userID, req.Email)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := s.JWT.GenerateRefreshToken(userID, 0)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	setRefreshCookie(w, refreshToken)
	JSON(w, http.StatusCreated, map[string]string{
		"access_token": accessToken,
		"user_id":      userID,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var userID, passwordHash string
	var tokenVersion int
	err := s.DB.QueryRow(
		"SELECT id, password_hash, token_version FROM users WHERE email = $1",
		req.Email,
	).Scan(&userID, &passwordHash, &tokenVersion)
	if err == sql.ErrNoRows {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	accessToken, err := s.JWT.GenerateAccessToken(userID, req.Email)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := s.JWT.GenerateRefreshToken(userID, tokenVersion)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	setRefreshCookie(w, refreshToken)
	JSON(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		Error(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	claims, err := s.JWT.ValidateRefreshToken(cookie.Value)
	if err != nil {
		Error(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	// Verify token version matches DB
	var email string
	var dbTokenVersion int
	err = s.DB.QueryRow(
		"SELECT email, token_version FROM users WHERE id = $1",
		claims.UserID,
	).Scan(&email, &dbTokenVersion)
	if err != nil {
		Error(w, http.StatusUnauthorized, "user not found")
		return
	}

	if claims.TokenVersion != dbTokenVersion {
		Error(w, http.StatusUnauthorized, "token has been revoked")
		return
	}

	accessToken, err := s.JWT.GenerateAccessToken(claims.UserID, email)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshToken, err := s.JWT.GenerateRefreshToken(claims.UserID, dbTokenVersion)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	setRefreshCookie(w, refreshToken)
	JSON(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	JSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r)
	if user == nil {
		Error(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		Error(w, http.StatusBadRequest, "current_password and new_password are required")
		return
	}

	var passwordHash string
	err := s.DB.QueryRow(
		"SELECT password_hash FROM users WHERE id = $1",
		user.UserID,
	).Scan(&passwordHash)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.CurrentPassword)); err != nil {
		Error(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	_, err = s.DB.Exec(
		"UPDATE users SET password_hash = $1, token_version = token_version + 1 WHERE id = $2",
		string(newHash), user.UserID,
	)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "password changed"})
}

func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})
}

