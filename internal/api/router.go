package api

import (
	"github.com/gofiber/fiber/v3/middleware/rewrite"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/handlers"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/middleware"
	"github.com/gurkengewuerz/GitCodeJudge/internal/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"

	"github.com/gofiber/fiber/v3"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
)

func SetupRouter(cfg *config.Config, pool *judge.Pool, scoreboardManager *scoreboard.ScoreboardManager) *fiber.App {
	app := fiber.New()

	// Initialize OAuth2 if enabled
	middleware.InitOAuth2(cfg)

	// Middleware
	app.Use(middleware.Logger())
	app.Use(recoverer.New(recoverer.Config{
		EnableStackTrace: true,
	}))

	sessionMiddleware, _ := session.NewWithStore()

	app.Use(sessionMiddleware)

	app.Use(rewrite.New(rewrite.Config{
		Rules: map[string]string{
			"/": func() string {
				if cfg.LeaderboardEnabled {
					return "/leaderboard"
				}
				return "/health"
			}(),
		},
	}))

	// Routes
	app.Get("/health", handlers.HealthCheck)

	// Webhook route with authentication
	app.Post("/webhook", middleware.ValidateGiteaWebhook(cfg.GiteaWebhookSecret), handlers.HandleWebhook(cfg, pool))

	// PDF for each problem
	app.Get("/pdf", handlers.HandlePDF(cfg))

	// Commit results
	app.Get("/results/:commit", handlers.HandleCommitResults())

	// Auth routes
	if cfg.OAuth2Issuer != "" {
		app.Get("/auth/login", middleware.HandleLogin)
		app.Get("/auth/callback", middleware.HandleCallback)
		app.Get("/auth/logout", middleware.HandleLogout)
	}

	// Scoreboard routes - only if enabled
	if cfg.LeaderboardEnabled {
		// Leaderboard requires auth if OAuth2 is enabled
		var oauthHandler []fiber.Handler
		if cfg.OAuth2Issuer != "" {
			oauthHandler = append(oauthHandler, middleware.RequireAuth(cfg))
		}

		// Individual user and workshop stats don't require auth
		app.Get("/user/:username", handlers.HandleUserProgress(scoreboardManager), oauthHandler...)
		app.Get("/workshop/:workshop/:task", handlers.HandleWorkshopStats(scoreboardManager), oauthHandler...)

		app.Get("/leaderboard", handlers.HandleLeaderboard(scoreboardManager), oauthHandler...)
	}

	return app
}
