package handlers

import (
	"bytes"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/handlers/templates"
	"github.com/gurkengewuerz/GitCodeJudge/internal/judge/scoreboard"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
	"sort"
	"strings"
	"time"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

func formatMarkdownToHTML(markdown string) (template.HTML, error) {
	var htmlBuf bytes.Buffer
	if err := md.Convert([]byte(markdown), &htmlBuf); err != nil {
		return "", err
	}
	return template.HTML(htmlBuf.String()), nil
}

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

		var b strings.Builder
		b.WriteString(fmt.Sprintf("# Progress for %s\n\n", username))

		// Task count summary
		b.WriteString(fmt.Sprintf("## Overview\n\n"))
		b.WriteString(fmt.Sprintf("Total Completed Tasks: **%d**\n\n", len(progress.Submissions)))

		// Sort submissions by timestamp (most recent first)
		sort.Slice(progress.Submissions, func(i, j int) bool {
			return progress.Submissions[i].Submission.Timestamp.After(progress.Submissions[j].Submission.Timestamp)
		})

		b.WriteString("## Completed Tasks\n\n")
		b.WriteString("| Workshop | Task | Completion Date | Repository | Commit |\n")
		b.WriteString("|----------|------|-----------------|------------|--------|\n")

		for _, submission := range progress.Submissions {
			b.WriteString(fmt.Sprintf("| %s | [%s](/workshop/%s/%s) | %s | [%s](%s) | [`%s`](%s/results/%s) |\n",
				submission.Workshop,
				submission.Task,
				submission.Workshop,
				submission.Task,
				submission.Submission.Timestamp.Format(time.RFC850),
				submission.Submission.RepoName,
				submission.Submission.CloneURL,
				submission.Submission.CommitID[:8],
				c.BaseURL(),
				submission.Submission.CommitID))
		}

		content, err := formatMarkdownToHTML(b.String())
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

		var b strings.Builder
		b.WriteString(fmt.Sprintf("# Statistics for [%s/%s](/pdf?task=%s/%s)\n\n", workshop, task, workshop, task))

		// Overview section
		b.WriteString("## Overview\n\n")
		b.WriteString(fmt.Sprintf("- Total Completions: **%d**\n", stats.TotalUsers))
		if stats.LatestSubmit.Year() > 1 {
			b.WriteString(fmt.Sprintf("- Latest Completion: **%s**\n", stats.LatestSubmit.Format(time.RFC850)))
		}
		b.WriteString("\n")

		// Sort submissions by timestamp (most recent first)
		sort.Slice(stats.Submissions, func(i, j int) bool {
			return stats.Submissions[i].Timestamp.After(stats.Submissions[j].Timestamp)
		})

		b.WriteString("## Submissions\n\n")
		b.WriteString("| User | Completion Date | Repository | Commit |\n")
		b.WriteString("|------|-----------------|------------|--------|\n")

		for _, submission := range stats.Submissions {
			parts := strings.Split(submission.RepoName, "/")
			username := parts[1]
			b.WriteString(fmt.Sprintf("| [%s](/user/%s) | %s | [%s](%s) | [`%s`](%s/results/%s) |\n",
				username,
				username,
				submission.Timestamp.Format(time.RFC850),
				submission.RepoName,
				submission.CloneURL,
				submission.CommitID[:8],
				c.BaseURL(),
				submission.CommitID))
		}

		content, err := formatMarkdownToHTML(b.String())
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

		var b strings.Builder
		b.WriteString("# 🏆 Leaderboard\n\n")

		// Add total participants info if available
		if len(leaderboard) > 0 {
			b.WriteString(fmt.Sprintf("Showing top %d participants\n\n", len(leaderboard)))
		}

		b.WriteString("| Rank | User | Completed Tasks | Latest Submission | Latest Repository |\n")
		b.WriteString("|------|------|-----------------|-------------------|------------------|\n")

		for i, entry := range leaderboard {
			b.WriteString(fmt.Sprintf("| %d | [%s](/user/%s) | %d | %s | %s |\n",
				i+1,
				entry.Username,
				entry.Username,
				entry.CompletedTasks,
				entry.LastSubmission.Format(time.RFC850),
				entry.LatestRepoName))
		}

		content, err := formatMarkdownToHTML(b.String())
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
