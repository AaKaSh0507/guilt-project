package services_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

func TestUsersService_CreateUser(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)

	// Test successful user creation
	u, err := users.CreateUser(ctx, "test-user@example.com", "hashed-password")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	if u.Email != "test-user@example.com" {
		t.Fatalf("email mismatch: expected test-user@example.com, got %s", u.Email)
	}
	if u.PasswordHash != "hashed-password" {
		t.Fatalf("password hash mismatch: expected hashed-password, got %s", u.PasswordHash)
	}
	if u.ID.String() == "" {
		t.Fatal("user ID should not be empty")
	}
}

func TestUsersService_CreateUserInvalidEmail(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)

	// Test invalid email formats
	invalidEmails := []string{
		"notanemail",
		"@example.com",
		"user@",
		"user @example.com",
		"user@example",
	}

	for _, email := range invalidEmails {
		_, err := users.CreateUser(ctx, email, "password")
		if err == nil {
			t.Fatalf("expected error for invalid email %s, but got none", email)
		}
	}
}

func TestUsersService_GetUser(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)

	// Create a user
	u, err := users.CreateUser(ctx, "get-user@example.com", "hashed-password")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	// Retrieve the user
	u2, err := users.GetUser(ctx, u.ID.String())
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}

	if u2.Email != u.Email {
		t.Fatalf("email mismatch: expected %s, got %s", u.Email, u2.Email)
	}
	if u2.ID != u.ID {
		t.Fatalf("ID mismatch: expected %s, got %s", u.ID, u2.ID)
	}
}

func TestUsersService_GetUserInvalidUUID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)

	// Test invalid UUID
	_, err := users.GetUser(ctx, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID, but got none")
	}
}
