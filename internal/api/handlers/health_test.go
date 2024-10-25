package handlers_test

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gurkengewuerz/GitCodeJudge/internal/api/handlers"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	app := fiber.New()
	app.Get("/health", handlers.HealthCheck)

	req, _ := app.Test(httptest.NewRequest("GET", "/health", nil))
	assert.Equal(t, 200, req.StatusCode)
}
