package main

import (
	"context"
	"log"
	"time"
	"database/sql"

	"github.com/redis/go-redis/v9"
	queue "guiltmachine/internal/queue"
	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // adjust docker-compose later
	})

	stream := queue.NewStreams(rdb, "ml:entries")
	_ = stream.EnsureGroup(ctx, "ml-workers")

	dbURL := "postgres://guilt:guiltpass@localhost:5432/guiltmachine?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}
	repo := sqlcrepo.New(db)

	entries := svcs.NewEntryService(repo.Entries) // prefs fetched inside

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
