package repo_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
)

func TestUsersRepo(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)

	t.Run("create and fetch user", func(t *testing.T) {
		// Create user
		u, err := repo.Users.CreateUser(ctx, "test@example.com", "hashedpassword")
		if err != nil {
			t.Fatalf("create user failed: %v", err)
		}

		// Fetch by id
		u2, err := repo.Users.GetUserByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("get user by id failed: %v", err)
		}
		if u2.Email != "test@example.com" {
			t.Fatalf("email mismatch: expected %s got %s", "test@example.com", u2.Email)
		}

		// Fetch by email
		u3, err := repo.Users.GetUserByEmail(ctx, "test@example.com")
		if err != nil {
			t.Fatalf("get user by email failed: %v", err)
		}
		if u3.ID != u.ID {
			t.Fatalf("id mismatch: %v vs %v", u3.ID, u.ID)
		}
	})

	t.Run("unique email constraint", func(t *testing.T) {
		// Create first user
		_, err := repo.Users.CreateUser(ctx, "unique@example.com", "hashedpassword")
		if err != nil {
			t.Fatalf("create first user failed: %v", err)
		}

		// Try to create another user with same email
		_, err = repo.Users.CreateUser(ctx, "unique@example.com", "differentpassword")
		if err == nil {
			t.Fatalf("expected duplicate email to fail, got nil")
		}
	})
}
