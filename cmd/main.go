package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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
