package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestJWTService() *JWTService {
	return NewJWTService(testSecret)
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		http.Error(w, "no user in context", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func TestMissingHeader(t *testing.T) {
	svc := newTestJWTService()
	handler := RequireAuth(svc)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestInvalidFormat(t *testing.T) {
	svc := newTestJWTService()
	handler := RequireAuth(svc)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestInvalidToken(t *testing.T) {
	svc := newTestJWTService()
	handler := RequireAuth(svc)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.a.valid.jwt")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestValidToken(t *testing.T) {
	svc := newTestJWTService()
	handler := RequireAuth(svc)(http.HandlerFunc(okHandler))

	token, err := svc.GenerateAccessToken("user-001", "valid@example.com")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
