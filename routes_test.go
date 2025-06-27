package autofiber_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestRouteInfo_Struct(t *testing.T) {
	route := &autofiber.RouteInfo{
		Path:        "/users",
		Method:      "GET",
		Handler:     nil,
		Options:     nil,
		OperationID: "getUsers",
	}

	assert.Equal(t, "/users", route.Path)
	assert.Equal(t, "GET", route.Method)
	assert.Equal(t, "getUsers", route.OperationID)
	assert.Nil(t, route.Handler)
	assert.Nil(t, route.Options)
}

// =============================================================================
// HTTP METHODS TESTS
// =============================================================================

func TestAutoFiber_Put(t *testing.T) {
	app := autofiber.New()

	app.Put("/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"method":  "PUT",
			"id":      c.Params("id"),
			"message": "User updated",
		})
	}, autofiber.WithDescription("Update user"),
		autofiber.WithTags("users", "update"))

	req := httptest.NewRequest(http.MethodPut, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoFiber_Delete(t *testing.T) {
	app := autofiber.New()

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"method":  "DELETE",
			"id":      c.Params("id"),
			"message": "User deleted",
		})
	}, autofiber.WithDescription("Delete user"),
		autofiber.WithTags("users", "delete"))

	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoFiber_Patch(t *testing.T) {
	app := autofiber.New()

	app.Patch("/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"method":  "PATCH",
			"id":      c.Params("id"),
			"message": "User partially updated",
		})
	}, autofiber.WithDescription("Partially update user"),
		autofiber.WithTags("users", "update"))

	req := httptest.NewRequest(http.MethodPatch, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoFiber_Head(t *testing.T) {
	app := autofiber.New()

	app.Head("/users/:id", func(c *fiber.Ctx) error {
		c.Set("X-User-ID", c.Params("id"))
		return c.SendStatus(http.StatusOK)
	}, autofiber.WithDescription("Get user headers"),
		autofiber.WithTags("users", "headers"))

	req := httptest.NewRequest(http.MethodHead, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "123", resp.Header.Get("X-User-ID"))
}

func TestAutoFiber_Options(t *testing.T) {
	app := autofiber.New()

	app.Options("/users", func(c *fiber.Ctx) error {
		c.Set("Allow", "GET, POST, PUT, DELETE")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.SendStatus(http.StatusOK)
	}, autofiber.WithDescription("Get allowed methods"),
		autofiber.WithTags("users", "cors"))

	req := httptest.NewRequest(http.MethodOptions, "/users", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "GET, POST, PUT, DELETE", resp.Header.Get("Allow"))
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
}

func TestAutoFiber_All(t *testing.T) {
	app := autofiber.New()

	app.All("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"method":  c.Method(),
			"message": "Health check",
			"status":  "ok",
		})
	}, autofiber.WithDescription("Health check endpoint"),
		autofiber.WithTags("system", "health"))

	// Test with GET
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with POST
	req = httptest.NewRequest(http.MethodPost, "/health", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with PUT
	req = httptest.NewRequest(http.MethodPut, "/health", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// ROUTES WITH REQUEST SCHEMA TESTS
// =============================================================================

func TestAutoFiber_Put_WithRequestSchema(t *testing.T) {
	type UpdateUserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	app := autofiber.New()

	app.Put("/users/:id", func(c *fiber.Ctx, req *UpdateUserRequest) error {
		return c.JSON(fiber.Map{
			"method": "PUT",
			"id":     c.Params("id"),
			"name":   req.Name,
			"email":  req.Email,
		})
	}, autofiber.WithRequestSchema(UpdateUserRequest{}),
		autofiber.WithDescription("Update user with validation"))

	// Test without request body (should fail parsing)
	req := httptest.NewRequest(http.MethodPut, "/users/123", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Should fail parsing
}

func TestAutoFiber_Delete_WithRequestSchema(t *testing.T) {
	type DeleteUserRequest struct {
		Reason string `json:"reason" validate:"required"`
	}

	app := autofiber.New()

	app.Delete("/users/:id", func(c *fiber.Ctx, req *DeleteUserRequest) error {
		return c.JSON(fiber.Map{
			"method": "DELETE",
			"id":     c.Params("id"),
			"reason": req.Reason,
		})
	}, autofiber.WithRequestSchema(DeleteUserRequest{}),
		autofiber.WithDescription("Delete user with reason"))

	// Test without request body (should fail validation)
	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode) // Should fail validation
}

// =============================================================================
// ROUTES WITH MIDDLEWARE TESTS
// =============================================================================

func TestAutoFiber_Put_WithMiddleware(t *testing.T) {
	app := autofiber.New()

	// Custom middleware
	customMiddleware := func(c *fiber.Ctx) error {
		c.Set("X-Custom-Header", "custom-value")
		return c.Next()
	}

	app.Put("/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"method": "PUT",
			"id":     c.Params("id"),
			"custom": c.Get("X-Custom-Header"),
		})
	}, autofiber.WithMiddleware(customMiddleware),
		autofiber.WithDescription("Update user with custom middleware"))

	req := httptest.NewRequest(http.MethodPut, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "custom-value", resp.Header.Get("X-Custom-Header"))
}

func TestAutoFiber_Delete_WithMiddleware(t *testing.T) {
	app := autofiber.New()

	// Auth middleware
	authMiddleware := func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization required",
			})
		}
		return c.Next()
	}

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"method":  "DELETE",
			"id":      c.Params("id"),
			"message": "User deleted",
		})
	}, autofiber.WithMiddleware(authMiddleware),
		autofiber.WithDescription("Delete user with auth"))

	// Test without auth header
	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Test with auth header
	req = httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	req.Header.Set("Authorization", "Bearer token123")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// ROUTES WITH RESPONSE SCHEMA TESTS
// =============================================================================

func TestAutoFiber_Put_WithResponseSchema(t *testing.T) {
	type UpdateUserRequest struct {
		Name string `json:"name" validate:"required"`
	}

	type UpdateUserResponse struct {
		ID   string `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	}

	app := autofiber.New()

	app.Put("/users/:id", func(c *fiber.Ctx, req *UpdateUserRequest) (interface{}, error) {
		return &UpdateUserResponse{
			ID:   c.Params("id"),
			Name: req.Name,
		}, nil
	}, autofiber.WithRequestSchema(UpdateUserRequest{}),
		autofiber.WithResponseSchema(UpdateUserResponse{}),
		autofiber.WithDescription("Update user with response validation"))

	// Test without request body (should fail parsing)
	req := httptest.NewRequest(http.MethodPut, "/users/123", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Should fail parsing
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestAutoFiber_AllMethods_Integration(t *testing.T) {
	app := autofiber.New()

	// Register all methods for the same path
	app.Get("/api/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"method": "GET", "action": "list"})
	}, autofiber.WithTags("users", "list"))

	app.Post("/api/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"method": "POST", "action": "create"})
	}, autofiber.WithTags("users", "create"))

	app.Put("/api/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"method": "PUT", "action": "update", "id": c.Params("id")})
	}, autofiber.WithTags("users", "update"))

	app.Delete("/api/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"method": "DELETE", "action": "delete", "id": c.Params("id")})
	}, autofiber.WithTags("users", "delete"))

	app.Patch("/api/users/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"method": "PATCH", "action": "partial_update", "id": c.Params("id")})
	}, autofiber.WithTags("users", "update"))

	// Test all methods
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		path := "/api/users"
		if method == "GET" || method == "POST" {
			path = "/api/users"
		} else {
			path = "/api/users/123"
		}

		req := httptest.NewRequest(method, path, nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestAutoFiber_ComplexRoute_WithAllOptions(t *testing.T) {
	type UserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"gte=18"`
	}

	type UserResponse struct {
		ID    string `json:"id" validate:"required"`
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"gte=18"`
	}

	app := autofiber.New()

	// Custom middleware
	logMiddleware := func(c *fiber.Ctx) error {
		c.Set("X-Request-ID", "req-123")
		return c.Next()
	}

	app.Put("/api/users/:id", func(c *fiber.Ctx, req *UserRequest) (interface{}, error) {
		return &UserResponse{
			ID:    c.Params("id"),
			Name:  req.Name,
			Email: req.Email,
			Age:   req.Age,
		}, nil
	}, autofiber.WithRequestSchema(UserRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithMiddleware(logMiddleware),
		autofiber.WithDescription("Update user with full validation"),
		autofiber.WithTags("users", "update", "api"))

	// Test without request body (should fail parsing)
	req := httptest.NewRequest(http.MethodPut, "/api/users/123", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // Should fail parsing
}
