package main

import (
	"context"
	"log"

	cacheDomain "guiltmachine/internal/cache/domain"
	cacheRedis "guiltmachine/internal/cache/redis"
	"guiltmachine/internal/db"
	"guiltmachine/internal/ml"
	v1 "guiltmachine/internal/proto/gen"
	sessionv1 "guiltmachine/internal/proto/gen/v1"
	reposqlc "guiltmachine/internal/repository/sqlc"
	"guiltmachine/internal/services"
	grpchandlers "guiltmachine/internal/transport/grpc"

	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	// init DB + queries
	database := db.MustDB(ctx, "postgres://guilt:guiltpass@localhost:5432/guiltmachine?sslmode=disable")
	repos := reposqlc.New(database)

	// init Redis cache
	cfg, err := cacheRedis.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load redis config: %v", err)
	}

	redisClient := cacheRedis.NewRedisClient(cfg)
	if err := cacheRedis.Ping(ctx, redisClient); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	redisCache := cacheRedis.NewRedisCache(redisClient)

	sessionCache := cacheDomain.NewSessionCache(redisCache)
	prefsCache := cacheDomain.NewPreferencesCache(redisCache)

	// init ML layer
	infer := ml.NewInferenceStub()
	mlService := ml.NewMLService(infer)

	// service
	userService := services.NewUserService(repos.Users)
	userHandler := grpchandlers.NewUserHandler(userService)

	sessionService := services.NewSessionService(repos.Sessions, sessionCache)
	sessionHandler := grpchandlers.NewSessionHandler(sessionService)

	entryService := services.NewEntryServiceWithML(repos.Entries, repos.Scores, mlService)
	entryHandler := grpchandlers.NewEntryHandler(entryService)

	scoreService := services.NewScoreService(repos.Scores)
	scoreHandler := grpchandlers.NewScoreHandler(scoreService)

	preferencesService := services.NewPreferencesService(repos.Preferences, prefsCache)
	preferencesHandler := grpchandlers.NewPreferencesHandler(preferencesService)

	StartGRPCServer(func(s *grpc.Server) {
		v1.RegisterUserServiceServer(s, userHandler)
		sessionv1.RegisterSessionServiceServer(s, sessionHandler)
		v1.RegisterEntryServiceServer(s, entryHandler)
		v1.RegisterScoreServiceServer(s, scoreHandler)
		v1.RegisterPreferencesServiceServer(s, preferencesHandler)
	})

	log.Println("api ready")
}
