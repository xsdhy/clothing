package auth

import (
	"clothing/internal/entity"
	"strings"
	"testing"
	"time"
)

func TestNewManagerAndTokenLifecycle(t *testing.T) {
	mgr, err := NewManager("test-secret", "issuer", time.Minute*30)
	if err != nil {
		t.Fatalf("unexpected error creating manager: %v", err)
	}

	user := &entity.DbUser{ID: 42, Email: "user@example.com", Role: entity.UserRoleAdmin}
	token, expiresAt, err := mgr.GenerateToken(user)
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("expected future expiry time")
	}

	claims, err := mgr.ParseToken(token)
	if err != nil {
		t.Fatalf("unexpected error parsing token: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, claims.UserID)
	}
	if !strings.EqualFold(claims.Email, user.Email) {
		t.Fatalf("expected email %s, got %s", user.Email, claims.Email)
	}
	if claims.Role != user.Role {
		t.Fatalf("expected role %s, got %s", user.Role, claims.Role)
	}
}

func TestNewManagerRequiresSecret(t *testing.T) {
	if _, err := NewManager("   ", "", time.Hour); err == nil {
		t.Fatal("expected error for empty secret")
	}
}
