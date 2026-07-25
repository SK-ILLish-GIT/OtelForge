package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/otelforge/otelforge/internal/config"
	"github.com/otelforge/otelforge/internal/db"
	"github.com/otelforge/otelforge/internal/queue"
	"github.com/otelforge/otelforge/worker/internal/runner"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}

	q, err := queue.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()
	if err := q.Declare(ctx); err != nil {
		log.Fatal(err)
	}

	store := db.NewStore(pool)
	run := runner.New(store, cfg.EncryptionKey)

	var wg sync.WaitGroup
	for i := 0; i < cfg.WorkerConcurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			log.Printf("worker goroutine %d started", id)
			err := q.Consume(ctx, func(msg queue.JobMessage) error {
				return run.Handle(ctx, msg.JobID)
			})
			if err != nil && ctx.Err() == nil {
				log.Printf("consumer error: %v", err)
			}
		}(i + 1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	cancel()
	wg.Wait()
	log.Println("worker stopped")
}
