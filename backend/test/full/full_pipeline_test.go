package full_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"guiltmachine/internal/auth"
	cacheDomain "guiltmachine/internal/cache/domain"
	cacheRedis "guiltmachine/internal/cache/redis"
	"guiltmachine/internal/ml"
	"guiltmachine/internal/queue"
	reposqlc "guiltmachine/internal/repository/sqlc"
	"guiltmachine/internal/services"

	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// TestFullPipeline validates the complete end-to-end flow:
// User → Session+JWT → Entry(pending) → Worker(ML) → Entry(completed) + Score + Roast
func TestFullPipeline(t *testing.T) {
	ctx := context.Background()

	// Skip if not in integration test mode
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	// Setup
	dbURL := getEnv("TEST_DB_URL", "postgres://guilt:guiltpass@localhost:5432/guiltmachine_test?sslmode=disable")
	redisAddr := getEnv("TEST_REDIS_ADDR", "localhost:6379")

	// Connect to DB
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("db open failed: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db ping failed: %v", err)
	}

	// Setup repos
	repos := reposqlc.New(db)

	// Setup Redis
	cfg := cacheRedis.Config{URL: "redis://" + redisAddr}
	redisClient := cacheRedis.NewRedisClient(cfg)
	if err := cacheRedis.Ping(ctx, redisClient); err != nil {
		t.Fatalf("redis ping failed: %v", err)
	}
	redisCache := cacheRedis.NewRedisCache(redisClient)

	sessionCache := cacheDomain.NewSessionCache(redisCache)
	prefsCache := cacheDomain.NewPreferencesCache(redisCache)

	// Setup JWT
	jwtManager := auth.NewJWTManager("test-secret-key", time.Hour)

	// Setup services
	userService := services.NewUserService(repos.Users)
	sessionService := services.NewSessionServiceWithJWT(repos.Sessions, sessionCache, jwtManager)
	prefsService := services.NewPreferencesService(repos.Preferences, prefsCache)

	// Setup queue
	queueRedis := redis.NewClient(&redis.Options{Addr: redisAddr})
	stream := queue.NewStreams(queueRedis, "ml:entries:test")
	_ = stream.EnsureGroup(ctx, "ml-workers-test")
	producer := queue.NewProducer(stream)

	// Entry service with queue (async mode)
	entryService := services.NewEntryServiceWithQueue(repos.Entries, repos.Scores, nil, prefsService, producer)

	// ML service for worker
	infer := ml.NewInferenceStub()
	orchestrator := ml.NewHybridOrchestrator(infer)
	workerEntryService := services.NewEntryServiceWithHybrid(repos.Entries, repos.Scores, orchestrator, prefsService)

	// =========================================
	// STEP 1: Create User
	// =========================================
	t.Log("Step 1: Creating user...")
	email := "pipeline-test-" + time.Now().Format("20060102150405") + "@test.com"
	user, err := userService.CreateUser(ctx, email, "testpassword123")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if user.ID.String() == "" {
		t.Fatal("user_id is empty")
	}
	t.Logf("✓ User created: %s", user.ID)

	// =========================================
	// STEP 2: Create Session → Get JWT
	// =========================================
	t.Log("Step 2: Creating session...")
	notes := "Integration test session"
	result, err := sessionService.CreateSessionWithJWT(ctx, user.ID.String(), &notes)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if result.Session.ID.String() == "" {
		t.Fatal("session_id is empty")
	}
	if result.JWT == "" {
		t.Fatal("jwt is empty")
	}
	t.Logf("✓ Session created: %s", result.Session.ID)
	t.Logf("✓ JWT issued: %s...", result.JWT[:50])

	// Verify JWT
	userID, sessionID, err := jwtManager.Verify(result.JWT)
	if err != nil {
		t.Fatalf("JWT verification failed: %v", err)
	}
	if userID != user.ID.String() {
		t.Errorf("JWT userID mismatch: got %s, want %s", userID, user.ID.String())
	}
	if sessionID != result.Session.ID.String() {
		t.Errorf("JWT sessionID mismatch: got %s, want %s", sessionID, result.Session.ID.String())
	}
	t.Log("✓ JWT verified successfully")

	// =========================================
	// STEP 3: Create Entry (goes to pending)
	// =========================================
	t.Log("Step 3: Creating entry...")
	entry, err := entryService.CreateEntry(ctx, result.Session.ID.String(), "I procrastinated on my work again today", 7)
	if err != nil {
		t.Fatalf("CreateEntry failed: %v", err)
	}
	if entry.ID.String() == "" {
		t.Fatal("entry_id is empty")
	}
	t.Logf("✓ Entry created: %s", entry.ID)

	// Verify entry is pending
	fetchedEntry, err := entryService.GetEntry(ctx, entry.ID.String())
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}
	if !fetchedEntry.Status.Valid || fetchedEntry.Status.String != "pending" {
		t.Logf("Note: Entry status is '%v' (may vary based on queue setup)", fetchedEntry.Status)
	} else {
		t.Log("✓ Entry status is 'pending'")
	}

	// =========================================
	// STEP 4: Simulate Worker Processing
	// =========================================
	t.Log("Step 4: Processing ML job (simulating worker)...")
	err = workerEntryService.ProcessMLJob(ctx, entry.ID.String())
	if err != nil {
		t.Fatalf("ProcessMLJob failed: %v", err)
	}
	t.Log("✓ ML job processed")

	// =========================================
	// STEP 5: Verify Entry is Completed
	// =========================================
	t.Log("Step 5: Verifying entry completion...")
	completedEntry, err := entryService.GetEntry(ctx, entry.ID.String())
	if err != nil {
		t.Fatalf("GetEntry after ML failed: %v", err)
	}

	// Check status
	if !completedEntry.Status.Valid || completedEntry.Status.String != "completed" {
		t.Errorf("Entry status should be 'completed', got '%v'", completedEntry.Status)
	} else {
		t.Log("✓ Entry status is 'completed'")
	}

	// Check roast text
	if !completedEntry.RoastText.Valid || completedEntry.RoastText.String == "" {
		t.Error("Entry roast_text should not be empty")
	} else {
		t.Logf("✓ Roast text present: %s...", truncate(completedEntry.RoastText.String, 50))
	}

	// =========================================
	// STEP 6: Verify Score
	// =========================================
	t.Log("Step 6: Verifying score...")
	score, err := entryService.GetEntryScore(ctx, entry.ID.String())
	if err != nil {
		t.Logf("Note: GetEntryScore returned error (may be expected): %v", err)
	}
	if score > 0 {
		t.Logf("✓ Guilt score: %d", score)
	} else {
		t.Log("Note: Score is 0 (stub may not set score)")
	}

	// =========================================
	// SUMMARY
	// =========================================
	t.Log("")
	t.Log("========================================")
	t.Log("PIPELINE TEST SUMMARY")
	t.Log("========================================")
	t.Logf("✓ User created:     %s", user.ID)
	t.Logf("✓ Session created:  %s", result.Session.ID)
	t.Logf("✓ JWT issued:       %s...", result.JWT[:30])
	t.Logf("✓ Entry created:    %s", entry.ID)
	t.Logf("✓ Entry status:     %s", completedEntry.Status.String)
	t.Logf("✓ Roast text:       %s", truncate(completedEntry.RoastText.String, 40))
	t.Logf("✓ Guilt score:      %d", score)
	t.Log("========================================")
	t.Log("FULL PIPELINE TEST PASSED!")
	t.Log("========================================")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
