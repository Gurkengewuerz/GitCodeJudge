package main

import (
	"context"
	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/db"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = db.Load(cfg)
	if err != nil {
		log.Fatalf("Failed to load db: %v", err)
	}
	defer db.DB.Close()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the GC process
	cleanup := db.StartValueLogGC(ctx)
	defer cleanup() // Will be called when main exits

	// Initialize judge pool
	docker, err := judge.NewDockerExecutor(cfg.DockerNetwork, cfg.DockerTimeout)
	executor := judge.NewExecutor(docker, cfg.TestPath)
	pool := judge.NewPool(executor, cfg.MaxParallelJudges)

	// Setup router
	router := api.SetupRouter(cfg, pool)

	// Start server
	go func() {
		if err := router.Listen(cfg.ServerAddress); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
