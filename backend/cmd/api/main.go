package main

import (
	"context"
	"log"

	"guiltmachine/internal/db"
	userproto "guiltmachine/internal/proto/gen"
	v1 "guiltmachine/internal/proto/gen/v1"
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

	// handler
	userHandler := grpchandlers.NewUserHandler(userService)

	sessionService := services.NewSessionService(repos.Sessions)
	sessionHandler := grpchandlers.NewSessionHandler(sessionService)

	StartGRPCServer(func(s *grpc.Server) {
		userproto.RegisterUserServiceServer(s, userHandler)
		v1.RegisterSessionServiceServer(s, sessionHandler)
	})

	log.Println("api ready")
}
