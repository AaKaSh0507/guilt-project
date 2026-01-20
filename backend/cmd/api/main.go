package main

import (
	"context"
	"log"

	"guiltmachine/internal/db"
	v1 "guiltmachine/internal/proto/gen"
	reposqlc "guiltmachine/internal/repository/sqlc"
	"guiltmachine/internal/services"
	grpchandlers "guiltmachine/internal/transport/grpc"

	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	// init DB + queries
	database := db.MustDB(ctx, "postgres://localhost/guilt")
	repos := reposqlc.New(database)

	// service
	userService := services.NewUserService(repos.Users)

	// handler
	userHandler := grpchandlers.NewUserHandler(userService)

	StartGRPCServer(func(s *grpc.Server) {
		v1.RegisterUserServiceServer(s, userHandler)
	})

	log.Println("api ready")
}
