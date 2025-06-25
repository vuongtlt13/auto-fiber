package autofiber

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// Test request schemas
type TestRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type TestResponse struct {
	Message string `json:"message"`
	Data    string `json:"data"`
}

func TestAutoFiberBasic(t *testing.T) {
	app := New()

	// Test handler with request parsing
	handler := func(c *fiber.Ctx, req *TestRequest) error {
		return c.JSON(TestResponse{
			Message: "Success",
			Data:    req.Name + " - " + req.Email,
		})
	}

	// Register route with auto-parse
	app.Post("/test", handler, WithRequestSchema(TestRequest{}))

	t.Run("Valid Request", func(t *testing.T) {
		reqBody := TestRequest{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Invalid Request", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "John Doe",
			// email is missing
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestAutoFiberSimple(t *testing.T) {
	app := New()

	// Test simple handler without request parsing
	handler := func(c *fiber.Ctx) error {
		return c.JSON(TestResponse{
			Message: "Simple handler",
			Data:    "no request parsing",
		})
	}

	// Register route without request schema
	app.Get("/simple", handler)

	t.Run("Simple Handler", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/simple", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAutoFiberGroup(t *testing.T) {
	app := New()

	// Test handler
	handler := func(c *fiber.Ctx, req *TestRequest) error {
		return c.JSON(TestResponse{
			Message: "Group handler",
			Data:    req.Name,
		})
	}

	// Create group and add route
	group := app.Group("/api/v1")
	group.Post("/test", handler, WithRequestSchema(TestRequest{}))

	t.Run("Group Route", func(t *testing.T) {
		reqBody := TestRequest{
			Name:  "John Doe",
			Email: "john@example.com",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAutoFiberMiddleware(t *testing.T) {
	app := New()

	// Test handler
	handler := func(c *fiber.Ctx) error {
		return c.JSON(TestResponse{
			Message: "Middleware test",
			Data:    c.Get("X-Custom-Header"),
		})
	}

	// Custom middleware
	customMiddleware := func(c *fiber.Ctx) error {
		c.Set("X-Custom-Header", "test-value")
		return c.Next()
	}

	// Register route with middleware
	app.Get("/middleware", handler, WithMiddleware(customMiddleware))

	t.Run("Middleware Route", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/middleware", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "test-value", resp.Header.Get("X-Custom-Header"))
	})
}
