package transport

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	cacheDomain "guiltmachine/internal/cache/domain"
	cacheRedis "guiltmachine/internal/cache/redis"
	ml "guiltmachine/internal/ml"
	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
	grpchandlers "guiltmachine/internal/transport/grpc"
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

func TestTransportLayerSetup(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// setup repo
	repo := sqlcrepo.New(db)

	// setup ml service
	infer := ml.NewInferenceStub()
	mlService := ml.NewMLService(infer)

	// setup redis caches
	sessionCache := openSessionCache(t)
	preferencesCache := openPreferencesCache(t)

	// setup services
	userService := svcs.NewUserService(repo.Users)
	sessionService := svcs.NewSessionService(repo.Sessions, sessionCache)
	prefsService := svcs.NewPreferencesService(repo.Preferences, preferencesCache)
	entryService := svcs.NewEntryServiceWithML(repo.Entries, repo.Scores, mlService)
	scoreService := svcs.NewScoreService(repo.Scores)

	// Verify all handlers can be created
	_ = grpchandlers.NewUserHandler(userService)
	_ = grpchandlers.NewSessionHandler(sessionService)
	_ = grpchandlers.NewEntryHandler(entryService)
	_ = grpchandlers.NewScoreHandler(scoreService)
	_ = grpchandlers.NewPreferencesHandler(prefsService)

	// Verify server can be started
	s := startTestGRPC(t, func(gs *grpc.Server) {
		// Handlers are created but registration would happen here
	})
	defer s.stop()

	// Verify connection can be made
	conn, err := grpc.Dial(s.getAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Fatalf("connection is nil")
	}
}
