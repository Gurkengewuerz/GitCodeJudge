package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gofiber/fiber/v3"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/handlers/templates"
	"github.com/gurkengewuerz/GitCodeJudge/internal/db"
	"github.com/gurkengewuerz/GitCodeJudge/internal/markdown"
	log "github.com/sirupsen/logrus"
)

func HandleCommitResults() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get commit hash from path parameters
		commitHash := c.Params("commit")
		if commitHash == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Commit hash is required",
			})
		}

		var mdContent []byte
		err := db.DB.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(commitHash))
			if errors.Is(err, badger.ErrKeyNotFound) {
				return fiber.NewError(404, "Results not found for this commit")
			}
			if err != nil {
				return err
			}

			mdContent, err = item.ValueCopy(nil)
			return err
		})

		if err != nil {
			log.WithError(err).Error("Failed to view database for results")

			var e *fiber.Error
			if errors.As(err, &e) {
				return c.Status(e.Code).JSON(fiber.Map{
					"error": e.Message,
				})
			}
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		content, err := markdown.FormatMarkdownToHTML(string(mdContent))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to generate HTML content",
			})
		}

		// Prepare template data
		data := templates.TemplateDataResult{
			Title:   fmt.Sprintf("Commit Results - %s", commitHash),
			Content: content, // Convert to template.HTML to prevent escaping
		}

		var buf bytes.Buffer
		if err := templates.GetResultTemplate().Execute(&buf, data); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to render template",
			})
		}

		// Set content type to HTML and send the response
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(buf.Bytes())
	}
}
