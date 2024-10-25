package api

import (
	"github.com/gofiber/fiber/v2/middleware/rewrite"
	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/handlers"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/middleware"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRouter(cfg *config.Config, pool *judge.Pool, scoreboardManager *scoreboard.ScoreboardManager) *fiber.App {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	app.Use(rewrite.New(rewrite.Config{
		Rules: map[string]string{
			"/": "/leaderboard",
		},
	}))

	// Routes
	app.Get("/health", handlers.HealthCheck)

	// Webhook route with authentication
	app.Post("/webhook", middleware.ValidateGiteaWebhook(cfg.GiteaWebhookSecret), handlers.HandleWebhook(cfg, pool))

	// PDF for each problem
	app.Get("/pdf", handlers.HandlePDF(cfg))

	// commit results
	app.Get("/results/:commit", handlers.HandleCommitResults())

	// Scoreboard
	app.Get("/user/:username", handlers.HandleUserProgress(scoreboardManager))
	app.Get("/workshop/:workshop/:task", handlers.HandleWorkshopStats(scoreboardManager))
	app.Get("/leaderboard", handlers.HandleLeaderboard(scoreboardManager))

	return app
}
