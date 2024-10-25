package handlers

import (
	"errors"
	"github.com/dgraph-io/badger/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/gurkengewuerz/GitCodeJudge/db"
)

func HandleCommitResults() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get commit hash from path parameters
		commitHash := c.Params("commit")
		if commitHash == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Commit hash is required",
			})
		}

		var htmlContent []byte
		err := db.DB.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(commitHash))
			if errors.Is(err, badger.ErrKeyNotFound) {
				return fiber.NewError(404, "Results not found for this commit")
			}
			if err != nil {
				return err
			}

			htmlContent, err = item.ValueCopy(nil)
			return err
		})

		if err != nil {
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

		// Set content type to HTML and send the response
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(htmlContent)
	}
}