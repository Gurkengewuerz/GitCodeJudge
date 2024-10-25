package handlers

import (
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/handlers/templates"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"
	"github.com/gurkengewuerz/GitCodeJudge/internal/markdown"
	"github.com/gurkengewuerz/GitCodeJudge/internal/models"
)

func HandleUserProgress(scoreboardManager *scoreboard.ScoreboardManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		username := c.Params("username")
		if username == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Username is required",
			})
		}

		progress, err := scoreboardManager.GetUserProgress(username)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to fetch user progress: %v", err),
			})
		}

		if progress == nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		content, err := markdown.FormatMarkdownToHTML(models.FormatUserStats(c.BaseURL(), progress))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to generate HTML content",
			})
		}

		data := templates.TemplateDataResult{
			Title:   fmt.Sprintf("User Progress - %s", username),
			Content: content,
		}

		var buf bytes.Buffer
		if err := templates.GetResultTemplate().Execute(&buf, data); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to render template",
			})
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(buf.Bytes())
	}
}

func HandleWorkshopStats(scoreboardManager *scoreboard.ScoreboardManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workshop := c.Params("workshop")
		task := c.Params("task")
		if workshop == "" || task == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Workshop and task are required",
			})
		}

		stats, err := scoreboardManager.GetWorkshopStats(workshop, task)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to fetch workshop stats: %v", err),
			})
		}

		if stats == nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Workshop/task not found",
			})
		}

		content, err := markdown.FormatMarkdownToHTML(models.FormatWorkshopStats(c.BaseURL(), workshop, task, stats))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to generate HTML content",
			})
		}

		data := templates.TemplateDataResult{
			Title:   fmt.Sprintf("Workshop Stats - %s/%s", workshop, task),
			Content: content,
		}

		var buf bytes.Buffer
		if err := templates.GetResultTemplate().Execute(&buf, data); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to render template",
			})
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(buf.Bytes())
	}
}

func HandleLeaderboard(scoreboardManager *scoreboard.ScoreboardManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit := 50

		leaderboard, err := scoreboardManager.GetLeaderboard(limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to fetch leaderboard: %v", err),
			})
		}

		content, err := markdown.FormatMarkdownToHTML(models.FormatLeaderboard(leaderboard))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to generate HTML content",
			})
		}

		data := templates.TemplateDataResult{
			Title:   "🏆 Workshop Leaderboard",
			Content: content,
		}

		var buf bytes.Buffer
		if err := templates.GetResultTemplate().Execute(&buf, data); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to render template",
			})
		}

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(buf.Bytes())
	}
}
