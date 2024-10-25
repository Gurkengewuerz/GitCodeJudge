package main

import (
	"context"
	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/db"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	cfg, err := config.Load()
	if err != nil {
		log.WithError(err).Fatal("Failed to load config")
	}

	log.SetLevel(log.Level(cfg.LogLevel))

	err = db.Load(cfg)
	if err != nil {
		log.WithError(err).Fatal("Failed to load db")
	}
	defer db.DB.Close()

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the GC process
	cleanup := db.StartValueLogGC(ctx)
	defer cleanup() // Will be called when main exits

	// Initialize judge pool
	scoreboardManager := scoreboard.NewScoreboardManager(db.DB)
	docker, err := judge.NewDockerExecutor(cfg.DockerNetwork, cfg.DockerTimeout)
	executor := judge.NewExecutor(docker, cfg.TestPath)
	pool := judge.NewPool(executor, scoreboardManager, cfg.MaxParallelJudges)

	// Setup router
	router := api.SetupRouter(cfg, pool, scoreboardManager)

	// Start server
	go func() {
		if err := router.Listen(cfg.ServerAddress); err != nil {
			log.WithError(err).Fatalf("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
}
