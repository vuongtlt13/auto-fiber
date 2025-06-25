package main

import (
	"time"

	"autofiber"

	"github.com/gofiber/fiber/v2"
)

// Request schemas
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" description:"User email address" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6" description:"User password" example:"password123"`
}

type RegisterRequest struct {
	Email     string    `json:"email" validate:"required,email" description:"User email address"`
	Password  string    `json:"password" validate:"required,min=6" description:"User password"`
	Name      string    `json:"name" validate:"required" description:"User full name"`
	BirthDate time.Time `json:"birth_date" description:"User birth date"`
}

type UserResponse struct {
	ID        int       `json:"id" description:"User ID"`
	Email     string    `json:"email" description:"User email"`
	Name      string    `json:"name" description:"User name"`
	CreatedAt time.Time `json:"created_at" description:"Account creation date"`
}

type UserFilterRequest struct {
	Page     int    `json:"page" validate:"gte=1" description:"Page number" example:"1"`
	Limit    int    `json:"limit" validate:"gte=1,lte=100" description:"Items per page" example:"10"`
	Search   string `json:"search" description:"Search term"`
	SortBy   string `json:"sort_by" description:"Sort field" example:"name"`
	SortDesc bool   `json:"sort_desc" description:"Sort descending"`
}

// Handler
type AuthHandler struct{}

// Handler with request parsing
func (h *AuthHandler) Login(c *fiber.Ctx, req *LoginRequest) error {
	// req is automatically parsed and validated
	return c.JSON(fiber.Map{
		"message": "Login successful",
		"email":   req.Email,
		"token":   "jwt_token_here",
	})
}

// Handler with request parsing and response formatting
func (h *AuthHandler) Register(c *fiber.Ctx, req *RegisterRequest) (interface{}, error) {
	// req is automatically parsed and validated
	// return data and error for automatic response formatting
	return UserResponse{
		ID:        1,
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now(),
	}, nil
}

// Handler with query parameters
func (h *AuthHandler) ListUsers(c *fiber.Ctx, req *UserFilterRequest) (interface{}, error) {
	// req is automatically parsed and validated
	return fiber.Map{
		"users": []UserResponse{
			{ID: 1, Email: "user1@example.com", Name: "User 1", CreatedAt: time.Now()},
			{ID: 2, Email: "user2@example.com", Name: "User 2", CreatedAt: time.Now()},
		},
		"total": 2,
		"page":  req.Page,
		"limit": req.Limit,
	}, nil
}

// Simple handler without request parsing
func (h *AuthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "timestamp": time.Now()})
}

func main() {
	// Create AutoFiber app with docs configuration
	app := autofiber.New().
		WithDocsInfo(autofiber.OpenAPIInfo{
			Title:       "AutoFiber API Example",
			Description: "A sample API demonstrating AutoFiber's auto-docs capabilities",
			Version:     "1.0.0",
			Contact: &autofiber.OpenAPIContact{
				Name:  "AutoFiber Team",
				Email: "team@autofiber.com",
			},
		}).
		WithDocsServer(autofiber.OpenAPIServer{
			URL:         "http://localhost:3000",
			Description: "Development server",
		})

	handler := &AuthHandler{}

	// Register routes with auto-parse and documentation
	app.Post("/login", handler.Login,
		autofiber.WithRequestSchema(LoginRequest{}),
		autofiber.WithDescription("Authenticate user and return JWT token"),
		autofiber.WithTags("auth", "authentication"),
	)

	app.Post("/register", handler.Register,
		autofiber.WithRequestSchema(RegisterRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Register a new user account"),
		autofiber.WithTags("auth", "user"),
	)

	app.Get("/users", handler.ListUsers,
		autofiber.WithRequestSchema(UserFilterRequest{}),
		autofiber.WithDescription("List users with filtering and pagination"),
		autofiber.WithTags("user", "admin"),
	)

	app.Get("/health", handler.Health,
		autofiber.WithDescription("Health check endpoint"),
		autofiber.WithTags("system"),
	)

	// Serve API documentation
	app.ServeDocs("/docs")
	app.ServeSwaggerUI("/swagger", "/docs")

	// Start server
	app.Listen(":3000")
}
