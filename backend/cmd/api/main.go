package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"guiltmachine/internal/db"
)

func main() {
	godotenv.Load()
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db.MustDB(ctx, dsn)

	// we will inject repos + services here later

	go func() {
		startGRPCServer(func(s *grpc.Server) {
			// Register services here
		})
	}()

	select {} // block forever for now
}
