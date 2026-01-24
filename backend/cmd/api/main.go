package main

import (
	"context"
	"log"
	"os"
	"time"

	"guiltmachine/internal/auth"
	cacheDomain "guiltmachine/internal/cache/domain"
	cacheRedis "guiltmachine/internal/cache/redis"
	"guiltmachine/internal/db"
	"guiltmachine/internal/ml"
	v1 "guiltmachine/internal/proto/gen"
	sessionv1 "guiltmachine/internal/proto/gen/v1"
	"guiltmachine/internal/queue"
	reposqlc "guiltmachine/internal/repository/sqlc"
	"guiltmachine/internal/services"
	grpchandlers "guiltmachine/internal/transport/grpc"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	ctx := context.Background()

	// Get config from environment with fallback defaults
	dbURL := getEnv("DB_URL", "postgres://guilt:guiltpass@localhost:5432/guiltmachine?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-key-change-in-production")
	jwtTTLHours := getEnv("JWT_TTL_HOURS", "24")

	// Parse JWT TTL
	ttlHours := 24
	if _, err := time.ParseDuration(jwtTTLHours + "h"); err == nil {
		ttlHours = int(mustParseInt(jwtTTLHours))
	}
	jwtTTL := time.Duration(ttlHours) * time.Hour

	// init JWT manager
	jwtManager := auth.NewJWTManager(jwtSecret, jwtTTL)

	// init DB + queries
	database := db.MustDB(ctx, dbURL)
	repos := reposqlc.New(database)

	// init Redis cache
	cfg := cacheRedis.Config{URL: "redis://" + redisAddr}
	redisClient := cacheRedis.NewRedisClient(cfg)
	if err := cacheRedis.Ping(ctx, redisClient); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	redisCache := cacheRedis.NewRedisCache(redisClient)

	sessionCache := cacheDomain.NewSessionCache(redisCache)
	prefsCache := cacheDomain.NewPreferencesCache(redisCache)

	// init ML layer (kept for fallback if needed)
	infer := ml.NewInferenceStub()
	_ = ml.NewHybridOrchestrator(infer) // orchestrator available if needed

	// init Redis queue for async ML processing
	queueRedis := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	stream := queue.NewStreams(queueRedis, "ml:entries")
	producer := queue.NewProducer(stream)

	// service
	userService := services.NewUserService(repos.Users)
	userHandler := grpchandlers.NewUserHandler(userService)

	sessionService := services.NewSessionServiceWithJWT(repos.Sessions, sessionCache, jwtManager)
	sessionHandler := grpchandlers.NewSessionHandler(sessionService)

	preferencesService := services.NewPreferencesService(repos.Preferences, prefsCache)

	// Use queue-based async ML processing
	entryService := services.NewEntryServiceWithQueue(repos.Entries, repos.Scores, nil, preferencesService, producer)
	entryHandler := grpchandlers.NewEntryHandler(entryService)

	scoreService := services.NewScoreService(repos.Scores)
	scoreHandler := grpchandlers.NewScoreHandler(scoreService)

	preferencesHandler := grpchandlers.NewPreferencesHandler(preferencesService)

	StartGRPCServerWithAuth(jwtManager, func(s *grpc.Server) {
		v1.RegisterUserServiceServer(s, userHandler)
		sessionv1.RegisterSessionServiceServer(s, sessionHandler)
		v1.RegisterEntryServiceServer(s, entryHandler)
		v1.RegisterScoreServiceServer(s, scoreHandler)
		v1.RegisterPreferencesServiceServer(s, preferencesHandler)
	})

	log.Println("api ready")
}

func mustParseInt(s string) int64 {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	if result == 0 {
		return 24 // default
	}
	return result
}
