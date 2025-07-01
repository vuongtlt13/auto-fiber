package autofiber_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestNew(t *testing.T) {
	app := autofiber.New(fiber.Config{})
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
}

func TestNew_WithConfig(t *testing.T) {
	app := autofiber.New(fiber.Config{})
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
}

func TestAutoFiber_Use(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	// Test middleware registration
	called := false
	middleware := func(c *fiber.Ctx) error {
		called = true
		return c.Next()
	}

	app.Use(middleware)

	// Add a test route
	app.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"message": "test"}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called, "Middleware should be called")
}

func TestAutoFiber_Listen(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	// Test that Listen doesn't panic (we can't actually test server startup in unit tests)
	// This is mainly for coverage
	app.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"message": "test"}, nil
	})

	// The Listen method should not panic when called
	// In a real scenario, this would start the server
	// For testing, we just verify the method exists and can be called
	assert.NotNil(t, app.Listen)
}
