package auth

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key-for-testing"

func TestGenerateAndValidateAccessToken(t *testing.T) {
	svc := NewJWTService(testSecret)
	userID := "user-123"
	email := "test@example.com"

	token, err := svc.GenerateAccessToken(userID, email)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("claims.UserID = %q, want %q", claims.UserID, userID)
	}
	if claims.Email != email {
		t.Errorf("claims.Email = %q, want %q", claims.Email, email)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	svc := NewJWTService(testSecret)
	userID := "user-456"
	tokenVersion := 3

	token, err := svc.GenerateRefreshToken(userID, tokenVersion)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("GenerateRefreshToken() returned empty token")
	}

	claims, err := svc.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("claims.UserID = %q, want %q", claims.UserID, userID)
	}
	if claims.TokenVersion != tokenVersion {
		t.Errorf("claims.TokenVersion = %d, want %d", claims.TokenVersion, tokenVersion)
	}
}

func TestExpiredAccessToken(t *testing.T) {
	svc := NewJWTService(testSecret)
	// Override TTL to a negative duration to produce an already-expired token
	svc.accessTokenTTL = -1 * time.Minute

	token, err := svc.GenerateAccessToken("user-789", "expired@example.com")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	_, err = svc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("ValidateAccessToken() expected error for expired token, got nil")
	}
}

func TestInvalidTokenVersion(t *testing.T) {
	svc := NewJWTService(testSecret)

	// Generate a refresh token with version 1
	token, err := svc.GenerateRefreshToken("user-abc", 1)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	// Validate and check that version is 1 (caller is responsible for comparing versions)
	claims, err := svc.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}

	// Simulate a version mismatch: expected version 2, got 1
	expectedVersion := 2
	if claims.TokenVersion == expectedVersion {
		t.Errorf("expected token version mismatch: got %d == %d", claims.TokenVersion, expectedVersion)
	}
}

func TestWrongSecret(t *testing.T) {
	svc := NewJWTService(testSecret)
	token, err := svc.GenerateAccessToken("user-xyz", "secret@example.com")
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	wrongSvc := NewJWTService("wrong-secret-key")
	_, err = wrongSvc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("ValidateAccessToken() expected error for wrong secret, got nil")
	}
}
