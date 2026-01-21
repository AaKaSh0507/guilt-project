package main

import (
	"context"
	"log"

	"guiltmachine/internal/db"
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

	// service
	userService := services.NewUserService(repos.Users)
	userHandler := grpchandlers.NewUserHandler(userService)

	sessionService := services.NewSessionService(repos.Sessions)
	sessionHandler := grpchandlers.NewSessionHandler(sessionService)

	entryService := services.NewEntryService(repos.Entries)
	entryHandler := grpchandlers.NewEntryHandler(entryService)

	StartGRPCServer(func(s *grpc.Server) {
		v1.RegisterUserServiceServer(s, userHandler)
		sessionv1.RegisterSessionServiceServer(s, sessionHandler)
		v1.RegisterEntryServiceServer(s, entryHandler)
	})

	log.Println("api ready")
}
