package services_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

func TestScoresService_CreateScore(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	scores := svcs.NewScoreService(repo.Scores)

	u, _ := users.CreateUser(ctx, "score-create@test.com", "password")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	s, err := scores.CreateScore(ctx, sess.ID.String(), 75, "")
	if err != nil {
		t.Fatalf("create score failed: %v", err)
	}

	if s.AggregateScore != 75 {
		t.Fatalf("score mismatch: expected 75, got %d", s.AggregateScore)
	}
	if s.SessionID != sess.ID {
		t.Fatalf("session ID mismatch: expected %s, got %s", sess.ID, s.SessionID)
	}
}

func TestScoresService_CreateScoreWithMetadata(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	scores := svcs.NewScoreService(repo.Scores)

	u, _ := users.CreateUser(ctx, "score-meta@test.com", "password")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	meta := `{"reason":"high guilt","source":"entry_analysis"}`
	s, err := scores.CreateScore(ctx, sess.ID.String(), 85, meta)
	if err != nil {
		t.Fatalf("create score with metadata failed: %v", err)
	}

	if s.AggregateScore != 85 {
		t.Fatalf("score mismatch: expected 85, got %d", s.AggregateScore)
	}
	if !s.Meta.Valid {
		t.Fatal("metadata should be valid")
	}
}

func TestScoresService_GetScore(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	scores := svcs.NewScoreService(repo.Scores)

	u, _ := users.CreateUser(ctx, "score-get@test.com", "password")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	s, err := scores.CreateScore(ctx, sess.ID.String(), 60, "")
	if err != nil {
		t.Fatalf("create score failed: %v", err)
	}

	s2, err := scores.GetScore(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("get score failed: %v", err)
	}

	if s2.AggregateScore != s.AggregateScore {
		t.Fatalf("score mismatch: expected %d, got %d", s.AggregateScore, s2.AggregateScore)
	}
	if s2.SessionID != s.SessionID {
		t.Fatalf("session ID mismatch: expected %s, got %s", s.SessionID, s2.SessionID)
	}
}

func TestScoresService_CreateScoreMinMax(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	scores := svcs.NewScoreService(repo.Scores)

	u, _ := users.CreateUser(ctx, "score-minmax@test.com", "password")

	// Test minimum score
	sess1, _ := sessions.CreateSession(ctx, u.ID.String(), nil)
	s1, err := scores.CreateScore(ctx, sess1.ID.String(), 0, "")
	if err != nil {
		t.Fatalf("create score 0 failed: %v", err)
	}
	if s1.AggregateScore != 0 {
		t.Fatalf("min score mismatch: expected 0, got %d", s1.AggregateScore)
	}

	// Test maximum score
	sess2, _ := sessions.CreateSession(ctx, u.ID.String(), nil)
	s2, err := scores.CreateScore(ctx, sess2.ID.String(), 100, "")
	if err != nil {
		t.Fatalf("create score 100 failed: %v", err)
	}
	if s2.AggregateScore != 100 {
		t.Fatalf("max score mismatch: expected 100, got %d", s2.AggregateScore)
	}
}

func TestScoresService_CreateScoreInvalidSessionID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	scores := svcs.NewScoreService(repo.Scores)

	_, err := scores.CreateScore(ctx, "not-a-uuid", 50, "")
	if err == nil {
		t.Fatal("expected error for invalid session UUID, but got none")
	}
}

func TestScoresService_GetScoreInvalidSessionID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	scores := svcs.NewScoreService(repo.Scores)

	_, err := scores.GetScore(ctx, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid session UUID, but got none")
	}
}
