package autofiber_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	app := newTestApp()

	app.Put("/users/:id", func(c *fiber.Ctx) (interface{}, error) {
		return c.JSON(fiber.Map{
			"method":  "PUT",
			"id":      c.Params("id"),
			"message": "User updated",
		}), nil
	}, autofiber.WithDescription("Update user"),
		autofiber.WithTags("users", "update"))

	req := httptest.NewRequest(http.MethodPut, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoFiber_Delete(t *testing.T) {
	app := newTestApp()

	app.Delete("/users/:id", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{
			"method":  "DELETE",
			"id":      c.Params("id"),
			"message": "User deleted",
		}, nil
	}, autofiber.WithDescription("Delete user"),
		autofiber.WithTags("users", "delete"))

	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoFiber_Patch(t *testing.T) {
	app := newTestApp()

	app.Patch("/users/:id", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{
			"method":  "PATCH",
			"id":      c.Params("id"),
			"message": "User partially updated",
		}, nil
	}, autofiber.WithDescription("Partially update user"),
		autofiber.WithTags("users", "update"))

	req := httptest.NewRequest(http.MethodPatch, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoFiber_Head(t *testing.T) {
	app := newTestApp()

	app.Head("/users/:id", func(c *fiber.Ctx) (interface{}, error) {
		c.Set("X-User-ID", c.Params("id"))
		return nil, nil
	}, autofiber.WithDescription("Get user headers"),
		autofiber.WithTags("users", "headers"))

	req := httptest.NewRequest(http.MethodHead, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "123", resp.Header.Get("X-User-ID"))
}

func TestAutoFiber_Options(t *testing.T) {
	app := newTestApp()

	app.Options("/users", func(c *fiber.Ctx) (interface{}, error) {
		c.Set("Allow", "GET, POST, PUT, DELETE")
		c.Set("Access-Control-Allow-Origin", "*")
		return nil, nil
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
	app := newTestApp()

	app.All("/health", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{
			"method":  c.Method(),
			"message": "Health check",
			"status":  "ok",
		}, nil
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

	app := newTestApp()

	app.Put("/users/:id", func(c *fiber.Ctx, req *UpdateUserRequest) (interface{}, error) {
		return fiber.Map{
			"method": "PUT",
			"id":     c.Params("id"),
			"name":   req.Name,
			"email":  req.Email,
		}, nil
	}, autofiber.WithRequestSchema(UpdateUserRequest{}),
		autofiber.WithDescription("Update user with validation"))

	// Test without request body (should fail parsing)
	req := httptest.NewRequest(http.MethodPut, "/users/123", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAutoFiber_Delete_WithRequestSchema(t *testing.T) {
	type DeleteUserRequest struct {
		Reason string `json:"reason" validate:"required"`
	}

	app := newTestApp()

	app.Delete("/users/:id", func(c *fiber.Ctx, req *DeleteUserRequest) (interface{}, error) {
		return fiber.Map{
			"method": "DELETE",
			"id":     c.Params("id"),
			"reason": req.Reason,
		}, nil
	}, autofiber.WithRequestSchema(DeleteUserRequest{}),
		autofiber.WithDescription("Delete user with reason"))

	// Test without request body (should fail validation)
	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

// =============================================================================
// ROUTES WITH MIDDLEWARE TESTS
// =============================================================================

func TestAutoFiber_Put_WithMiddleware(t *testing.T) {
	app := newTestApp()

	// Custom middleware
	customMiddleware := func(c *fiber.Ctx) error {
		c.Set("X-Custom-Header", "custom-value")
		return c.Next()
	}

	app.Put("/users/:id", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{
			"method": "PUT",
			"id":     c.Params("id"),
			"custom": c.Get("X-Custom-Header"),
		}, nil
	}, autofiber.WithMiddleware(customMiddleware),
		autofiber.WithDescription("Update user with custom middleware"))

	req := httptest.NewRequest(http.MethodPut, "/users/123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "custom-value", resp.Header.Get("X-Custom-Header"))
}

func TestAutoFiber_Delete_WithMiddleware(t *testing.T) {
	app := newTestApp()

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

	app.Delete("/users/:id", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{
			"method":  "DELETE",
			"id":      c.Params("id"),
			"message": "User deleted",
		}, nil
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

func TestAutoFiber_JWTAuth_NoSchema(t *testing.T) {
	app := newTestApp()

	app.Get("/secure-no-schema", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"ok": true}, nil
	}, autofiber.WithJwtAuth())

	// Missing Authorization -> 401
	req := httptest.NewRequest(http.MethodGet, "/secure-no-schema", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// With Authorization -> 200
	req = httptest.NewRequest(http.MethodGet, "/secure-no-schema", nil)
	req.Header.Set("Authorization", "Bearer token123")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoFiber_JWTAuth_RequestSchema_Inferred(t *testing.T) {
	app := newTestApp()

	type JwtHeaderRequest struct {
		Authorization string `parse:"header:Authorization" validate:"required"`
	}

	app.Get("/secure-schema", func(c *fiber.Ctx, req *JwtHeaderRequest) (interface{}, error) {
		return fiber.Map{"auth": req.Authorization}, nil
	}, autofiber.WithRequestSchema(JwtHeaderRequest{}))

	// Missing Authorization -> 401 (inferred RequireJWTAuth)
	req := httptest.NewRequest(http.MethodGet, "/secure-schema", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// With Authorization -> 200
	req = httptest.NewRequest(http.MethodGet, "/secure-schema", nil)
	req.Header.Set("Authorization", "Bearer token123")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoFiber_Delete_WithBodyParsing(t *testing.T) {
	app := newTestApp()

	type BulkDeleteRequest struct {
		IDs []int `parse:"body:ids" json:"ids" validate:"required,len=2"`
	}

	app.Delete("/bulk-delete", func(c *fiber.Ctx, req *BulkDeleteRequest) (interface{}, error) {
		return fiber.Map{"deleted": req.IDs}, nil
	}, autofiber.WithRequestSchema(BulkDeleteRequest{}))

	// With JSON body (should parse)
	body := `{"ids":[1,2]}`
	req := httptest.NewRequest(http.MethodDelete, "/bulk-delete", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var result map[string][]int
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, result["deleted"])
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

	app := newTestApp()

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
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestAutoFiber_AllMethods_Integration(t *testing.T) {
	app := newTestApp()

	// Test all HTTP methods
	app.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"method": "GET", "message": "success"}, nil
	})

	app.Post("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"method": "POST", "message": "success"}, nil
	})

	app.Put("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"method": "PUT", "message": "success"}, nil
	})

	app.Delete("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"method": "DELETE", "message": "success"}, nil
	})

	app.Patch("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"method": "PATCH", "message": "success"}, nil
	})

	app.Head("/test", func(c *fiber.Ctx) (interface{}, error) {
		return nil, nil
	})

	app.Options("/test", func(c *fiber.Ctx) (interface{}, error) {
		return nil, nil
	})

	// Test all methods
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/test", nil)
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

	app := newTestApp()

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
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
