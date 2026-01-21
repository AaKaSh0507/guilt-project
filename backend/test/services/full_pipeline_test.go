package services

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	cacheDomain "guiltmachine/internal/cache/domain"
	cacheRedis "guiltmachine/internal/cache/redis"
	ml "guiltmachine/internal/ml"
	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		t.Fatalf("TEST_DB_URL not set")
	}

	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		t.Fatalf("failed to parse db config: %v", err)
	}

	dsn := stdlib.RegisterConnConfig(cfg)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db failed: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping db failed: %v", err)
	}

	return db
}

func openSessionCache(t *testing.T) *cacheDomain.SessionCache {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Fatalf("TEST_REDIS_URL not set")
	}
	cfg := cacheRedis.Config{URL: url}
	client := cacheRedis.NewRedisClient(cfg)
	return cacheDomain.NewSessionCache(cacheRedis.NewRedisCache(client))
}

func openPreferencesCache(t *testing.T) *cacheDomain.PreferencesCache {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Fatalf("TEST_REDIS_URL not set")
	}
	cfg := cacheRedis.Config{URL: url}
	client := cacheRedis.NewRedisClient(cfg)
	return cacheDomain.NewPreferencesCache(cacheRedis.NewRedisCache(client))
}

func TestFullPipeline(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer db.Close()

	repo := sqlcrepo.New(db)

	// Create ML service
	infer := ml.NewInferenceStub()
	orchestrator := ml.NewHybridOrchestrator(infer)

	// Get caches
	prefsCache := openPreferencesCache(t)
	sessCache := openSessionCache(t)

	// Create services
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, sessCache)
	prefs := svcs.NewPreferencesService(repo.Preferences, prefsCache)
	entries := svcs.NewEntryServiceWithHybrid(repo.Entries, repo.Scores, orchestrator, prefs)
	scores := svcs.NewScoreService(repo.Scores)

	// Step 1: Create user
	u, err := users.CreateUser(ctx, "final@test.com", "password123")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	if u.Email != "final@test.com" {
		t.Fatalf("email mismatch: expected final@test.com, got %s", u.Email)
	}

	// Step 2: Create preferences
	_, err = prefs.UpsertPreferences(ctx, u.ID.String(), nil, true, "")
	if err != nil {
		t.Fatalf("upsert preferences failed: %v", err)
	}

	// Step 3: Create session
	sess, err := sessions.CreateSession(ctx, u.ID.String(), nil)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	if sess.UserID != u.ID {
		t.Fatalf("session user_id mismatch")
	}

	// Step 4: Create entry (triggers ML scoring)
	entry, err := entries.CreateEntry(ctx, sess.ID.String(), "I procrastinated again but mildly", 3)
	if err != nil {
		t.Fatalf("create entry failed: %v", err)
	}

	if entry.EntryText != "I procrastinated again but mildly" {
		t.Fatalf("entry text mismatch: expected 'I procrastinated again but mildly', got %s", entry.EntryText)
	}

	// Step 5: Retrieve score (created by ML pipeline)
	sc, err := scores.GetScore(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("get score failed: %v", err)
	}

	if sc.AggregateScore <= 0 {
		t.Fatalf("expected positive guilt score, got %d", sc.AggregateScore)
	}

	// Step 6: Retrieve entries
	entries_list, err := entries.ListEntries(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("list entries failed: %v", err)
	}

	if len(entries_list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries_list))
	}

	if entries_list[0].ID != entry.ID {
		t.Fatalf("entry ID mismatch")
	}
}

func TestFullPipelineMultipleEntries(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer db.Close()

	repo := sqlcrepo.New(db)
	infer := ml.NewInferenceStub()
	orchestrator := ml.NewHybridOrchestrator(infer)
	sessCache := openSessionCache(t)
	prefsCache := openPreferencesCache(t)

	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, sessCache)
	prefs := svcs.NewPreferencesService(repo.Preferences, prefsCache)
	entries := svcs.NewEntryServiceWithHybrid(repo.Entries, repo.Scores, orchestrator, prefs)

	// Create user and session
	u, _ := users.CreateUser(ctx, "multi@test.com", "password123")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	// Create multiple entries
	e1, _ := entries.CreateEntry(ctx, sess.ID.String(), "lazy", 2)
	e2, _ := entries.CreateEntry(ctx, sess.ID.String(), "I procrastinated all day", 4)
	e3, _ := entries.CreateEntry(ctx, sess.ID.String(), "procrastinated once more and again again", 5)

	// Verify all entries were created
	entries_list, err := entries.ListEntries(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("list entries failed: %v", err)
	}

	if len(entries_list) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries_list))
	}

	// Verify entry IDs
	entryIDs := make(map[string]bool)
	for _, e := range entries_list {
		entryIDs[e.ID.String()] = true
	}

	if !entryIDs[e1.ID.String()] || !entryIDs[e2.ID.String()] || !entryIDs[e3.ID.String()] {
		t.Fatalf("not all created entries found in list")
	}
}

func TestFullPipelineMLIntegration(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	defer db.Close()

	repo := sqlcrepo.New(db)
	infer := ml.NewInferenceStub()
	orchestrator := ml.NewHybridOrchestrator(infer)
	sessCache := openSessionCache(t)
	prefsCache := openPreferencesCache(t)

	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, sessCache)
	prefs := svcs.NewPreferencesService(repo.Preferences, prefsCache)
	entries := svcs.NewEntryServiceWithHybrid(repo.Entries, repo.Scores, orchestrator, prefs)
	scores := svcs.NewScoreService(repo.Scores)

	// Create user and session
	u, _ := users.CreateUser(ctx, "ml@test.com", "password123")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	// Create short entry
	_, _ = entries.CreateEntry(ctx, sess.ID.String(), "lazy", 1)

	// Create long entry
	_, _ = entries.CreateEntry(ctx, sess.ID.String(), "I procrastinated the entire day and accomplished nothing productive at all", 5)

	// Get score
	sc, err := scores.GetScore(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("get score failed: %v", err)
	}

	// Verify score is positive
	if sc.AggregateScore <= 0 {
		t.Fatalf("expected positive score for entry, got %d", sc.AggregateScore)
	}
}
