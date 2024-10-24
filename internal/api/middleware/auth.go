package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
)

func ValidateGiteaWebhook(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		signature := c.Get("X-Gitea-Signature")
		if signature == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Missing signature",
			})
		}

		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(c.Body())
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid signature",
			})
		}

		return c.Next()
	}
}
