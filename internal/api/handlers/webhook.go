package handlers

import (
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"github.com/gurkengewuerz/GitCodeJudge/internal/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/gitea"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
	log "github.com/sirupsen/logrus"
)

func HandleWebhook(cfg *config.Config, pool *judge.Pool) fiber.Handler {
	return func(c fiber.Ctx) error {
		var pushEvent models.GiteaPushEvent
		if err := json.Unmarshal(c.Body(), &pushEvent); err != nil {
			log.WithError(err).Warn("Invalid webhook payload")
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid webhook payload",
			})
		}

		// Create submission
		submission := models.Submission{
			RepoName:   pushEvent.Repository.FullName,
			CommitID:   pushEvent.After,
			BranchName: pushEvent.Ref,
			CloneURL:   pushEvent.Repository.CloneURL,
			GitClient:  gitea.NewGiteaClient(cfg.GiteaURL, cfg.GiteaToken),
		}

		// Submit to judge pool
		pool.Submit(submission)

		return c.JSON(fiber.Map{
			"status": "submission accepted",
		})
	}
}
