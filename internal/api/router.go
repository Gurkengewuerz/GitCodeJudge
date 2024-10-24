package api

import (
	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/handlers"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/middleware"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRouter(cfg *config.Config, pool *judge.Pool) *fiber.App {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Routes
	app.Get("/health", handlers.HealthCheck)

	// Webhook route with authentication
	app.Post("/webhook",
		middleware.ValidateGiteaWebhook(cfg.GiteaWebhookSecret),
		handlers.HandleWebhook(cfg, pool))

	return app
}
