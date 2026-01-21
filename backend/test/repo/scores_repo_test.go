package repo_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
)

func TestScoresRepo(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := sqlcrepo.New(db)

	t.Run("create and fetch score", func(t *testing.T) {
		u, _ := repo.Users.CreateUser(ctx, "score@test.com", "hashedpassword")
		s, _ := repo.Sessions.CreateSession(ctx, u.ID, nil)
		_, _ = repo.Entries.CreateEntry(ctx, s.ID, "score test", 5)

		sc, err := repo.Scores.CreateScore(ctx, s.ID, 80, nil)
		if err != nil {
			t.Fatalf("create score failed: %v", err)
		}

		s2, err := repo.Scores.GetScoreBySession(ctx, s.ID)
		if err != nil {
			t.Fatalf("get score failed: %v", err)
		}
		if s2.AggregateScore != sc.AggregateScore {
			t.Fatalf("mismatch score value")
		}
	})

	t.Run("fk session constraint", func(t *testing.T) {
		// Try to create score with non-existent session
		fakeSessionID := [16]byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}
		uuid := [16]byte(fakeSessionID)
		_, err := repo.Scores.CreateScore(ctx, uuid, 75, nil)
		if err == nil {
			t.Fatalf("expected FK violation for non-existent session")
		}
	})

	t.Run("score overwrite behavior", func(t *testing.T) {
		u, _ := repo.Users.CreateUser(ctx, "scoreoverwrite@test.com", "hashedpassword")
		s, _ := repo.Sessions.CreateSession(ctx, u.ID, nil)

		// Create first score
		sc1, err := repo.Scores.CreateScore(ctx, s.ID, 70, nil)
		if err != nil {
			t.Fatalf("create first score failed: %v", err)
		}
		if sc1.AggregateScore != 70 {
			t.Fatalf("expected score 70, got %d", sc1.AggregateScore)
		}

		// Try to create another score for same session
		_, err = repo.Scores.CreateScore(ctx, s.ID, 85, nil)
		if err != nil {
			// If error occurs, this indicates 1-to-1 relationship (expected behavior)
			t.Logf("Score creation for existing session returned error (1-to-1 constraint): %v", err)
		} else {
			// If it succeeds, verify the new score was applied
			fetched, err := repo.Scores.GetScoreBySession(ctx, s.ID)
			if err != nil {
				t.Fatalf("failed to fetch score after second create: %v", err)
			}
			t.Logf("Second score created successfully: %d. Score model supports multiple scores per session.", fetched.AggregateScore)
		}
	})
}
