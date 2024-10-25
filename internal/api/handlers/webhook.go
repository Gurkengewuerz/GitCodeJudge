package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gurkengewuerz/GitCodeJudge/config"
	"github.com/gurkengewuerz/GitCodeJudge/internal/gitea"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"

	"github.com/gofiber/fiber/v2"
)

func HandleWebhook(cfg *config.Config, pool *judge.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var pushEvent models.GiteaPushEvent
		if err := json.Unmarshal(c.Body(), &pushEvent); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid webhook payload",
			})
		}

		baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname())

		// Create submission
		submission := models.Submission{
			RepoName:   pushEvent.Repository.FullName,
			CommitID:   pushEvent.After,
			BranchName: pushEvent.Ref,
			CloneURL:   pushEvent.Repository.CloneURL,
			GitClient:  gitea.NewClient(cfg.GiteaURL, cfg.GiteaToken),
			BaseURL:    baseURL,
		}

		// Submit to judge pool
		pool.Submit(submission)

		return c.JSON(fiber.Map{
			"status": "submission accepted",
		})
	}
}
