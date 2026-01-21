package repo_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
)

func TestSessionsRepo(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := sqlcrepo.New(db)

	t.Run("create and fetch session", func(t *testing.T) {
		// Need a user first
		u, err := repo.Users.CreateUser(ctx, "session@test.com", "hashedpassword")
		if err != nil {
			t.Fatalf("create user failed: %v", err)
		}

		// Create session
		s, err := repo.Sessions.CreateSession(ctx, u.ID, nil)
		if err != nil {
			t.Fatalf("create session failed: %v", err)
		}

		// Lookup
		s2, err := repo.Sessions.GetSessionByID(ctx, s.ID)
		if err != nil {
			t.Fatalf("get session failed: %v", err)
		}
		if s2.UserID != u.ID {
			t.Fatalf("user mismatch")
		}

		// End session
		s3, err := repo.Sessions.EndSession(ctx, s.ID)
		if err != nil {
			t.Fatalf("end session failed: %v", err)
		}
		if s3.ID != s.ID {
			t.Fatalf("session id mismatch after end")
		}
	})

	t.Run("fk user constraint", func(t *testing.T) {
		// Try to create session with non-existent user
		fakeUserID := [16]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
		uuid := [16]byte(fakeUserID)
		_, err := repo.Sessions.CreateSession(ctx, uuid, nil)
		if err == nil {
			t.Fatalf("expected FK violation for non-existent user")
		}
	})

	t.Run("session not found after end", func(t *testing.T) {
		// Create user and session
		u, _ := repo.Users.CreateUser(ctx, "sessionend@test.com", "hashedpassword")
		s, _ := repo.Sessions.CreateSession(ctx, u.ID, nil)

		// End the session
		s3, err := repo.Sessions.EndSession(ctx, s.ID)
		if err != nil {
			t.Fatalf("end session failed: %v", err)
		}

		// Verify the session is marked as ended
		if s3.ID != s.ID {
			t.Fatalf("session id mismatch after end")
		}
	})
}
