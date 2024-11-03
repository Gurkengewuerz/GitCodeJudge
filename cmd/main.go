package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/db"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "gitcodejudge",
		Short: "GitCodeJudge is a code evaluation system",
		Long: `GitCodeJudge is a system that evaluates code submissions 
against predefined test cases using Docker containers.`,
	}

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the GitCodeJudge server",
		Long:  `Start the GitCodeJudge server with the specified configuration`,
		Run:   serve,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Add commands
	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)
}

func serve(cmd *cobra.Command, args []string) {
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
	defer cleanup()

	// Initialize judge pool
	scoreboardManager := scoreboard.NewScoreboardManager(db.DB)
	docker, err := judge.NewDockerExecutor(cfg.DockerNetwork, cfg.DockerTimeout)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize docker executor")
	}
	executor := judge.NewExecutor(docker, cfg.TestPath)
	pool := judge.NewPool(executor, scoreboardManager, cfg.MaxParallelJudges)

	// Setup router
	router := api.SetupRouter(cfg, pool, scoreboardManager)

	// Start server
	go func() {
		if err := router.Listen(cfg.ServerAddress); err != nil {
			log.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Failed to execute root command")
		os.Exit(1)
	}
}
