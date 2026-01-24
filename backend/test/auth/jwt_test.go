package auth_test

import (
	"testing"
	"time"

	"guiltmachine/internal/auth"
)

func TestJWTManager_IssueAndVerify(t *testing.T) {
	m := auth.NewJWTManager("test-secret-key", time.Hour)

	userID := "user-123"
	sessionID := "session-456"

	// Issue token
	token, err := m.Issue(userID, sessionID)
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Verify token
	gotUserID, gotSessionID, err := m.Verify(token)
	if err != nil {
		t.Fatalf("failed to verify token: %v", err)
	}

	if gotUserID != userID {
		t.Errorf("expected userID %q, got %q", userID, gotUserID)
	}
	if gotSessionID != sessionID {
		t.Errorf("expected sessionID %q, got %q", sessionID, gotSessionID)
	}
}

func TestJWTManager_ExpiredToken(t *testing.T) {
	// Create manager with very short TTL
	m := auth.NewJWTManager("test-secret-key", -time.Hour) // Already expired

	token, err := m.Issue("user-123", "session-456")
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Verify should fail for expired token
	_, _, err = m.Verify(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	m := auth.NewJWTManager("test-secret-key", time.Hour)

	// Test with invalid token
	_, _, err := m.Verify("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	m1 := auth.NewJWTManager("secret-1", time.Hour)
	m2 := auth.NewJWTManager("secret-2", time.Hour)

	// Issue with one secret
	token, err := m1.Issue("user-123", "session-456")
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Verify with different secret should fail
	_, _, err = m2.Verify(token)
	if err == nil {
		t.Fatal("expected error when verifying with wrong secret")
	}
}

func TestJWTManager_TTL(t *testing.T) {
	ttl := 2 * time.Hour
	m := auth.NewJWTManager("test-secret", ttl)

	if m.TTL() != ttl {
		t.Errorf("expected TTL %v, got %v", ttl, m.TTL())
	}
}

func TestJWTManager_EmptyClaims(t *testing.T) {
	m := auth.NewJWTManager("test-secret", time.Hour)

	// Issue with empty claims - should still work
	token, err := m.Issue("", "")
	if err != nil {
		t.Fatalf("failed to issue token: %v", err)
	}

	// Verify should fail because claims are empty
	_, _, err = m.Verify(token)
	if err == nil {
		t.Fatal("expected error for empty claims")
	}
}
