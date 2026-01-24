package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"guiltmachine/internal/ml"
	queue "guiltmachine/internal/queue"
	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"

	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
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

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	stream := queue.NewStreams(rdb, "ml:entries")
	_ = stream.EnsureGroup(ctx, "ml-workers")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}
	repo := sqlcrepo.New(db)

	// init ML layer
	infer := ml.NewInferenceStub()
	orchestrator := ml.NewHybridOrchestrator(infer)

	// init services with orchestrator for ML processing
	prefsService := svcs.NewPreferencesService(repo.Preferences, nil)
	entries := svcs.NewEntryServiceWithHybrid(repo.Entries, repo.Scores, orchestrator, prefsService)

	consumer := queue.NewConsumer(stream, "ml-workers", "ml-consumer-1", 5*time.Second)

	log.Println("ML Worker running...")
	for {
		jobs, err := consumer.Poll(ctx)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		for _, job := range jobs {
			log.Printf("Processing job entry=%s user=%s", job.EntryID, job.UserID)
			if err := entries.ProcessMLJob(ctx, job.EntryID); err != nil {
				log.Printf("job failed: %v", err)
			}
		}
	}
}
