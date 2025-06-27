package main

import (
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

// Request schemas with multi-source parsing
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

// UserFilterRequest demonstrates parsing from multiple sources using parse tag
type UserFilterRequest struct {
	// Query parameters
	Page     int    `parse:"query:page" validate:"gte=1" description:"Page number" example:"1"`
	Limit    int    `parse:"query:limit" validate:"gte=1,lte=100" description:"Items per page" example:"10"`
	Search   string `parse:"query:search" description:"Search term"`
	SortBy   string `parse:"query:sort_by" description:"Sort field" example:"name"`
	SortDesc bool   `parse:"query:sort_desc" description:"Sort descending"`

	// Headers
	Authorization string `parse:"header:Authorization" validate:"required" description:"Bearer token"`
	Accept        string `parse:"header:Accept" description:"Accept header"`

	// Cookies
	SessionID string `parse:"cookie:session_id" description:"Session ID from cookie"`
}

// GetUserRequest demonstrates smart parsing (auto-detect source)
type GetUserRequest struct {
	// These will be auto-detected based on HTTP method
	UserID         int  `parse:"auto:user_id" validate:"required" description:"User ID (auto-detected from path/query/body)"`
	IncludeProfile bool `parse:"auto:include_profile" description:"Include user profile data"`
	IncludePosts   bool `parse:"auto:include_posts" description:"Include user posts"`

	// Headers
	Authorization string `parse:"header:Authorization" validate:"required" description:"Bearer token"`
}

// Request schema with parse tag and json tag support
type CreateUserRequest struct {
	// Path parameter
	OrgID int `parse:"path:org_id" validate:"required" description:"Organization ID"`

	// Query parameters
	Role     string `parse:"query:role" validate:"required,oneof=admin user" description:"User role"`
	IsActive bool   `parse:"query:active" description:"User active status"`

	// Headers
	APIKey string `parse:"header:X-API-Key" validate:"required" description:"API key"`

	// Body fields with json tag aliasing
	Email    string `json:"user_email" parse:"body:email" validate:"required,email" description:"User email"`
	Password string `json:"user_password" parse:"body:password" validate:"required,min=6" description:"User password"`
	Name     string `json:"full_name" parse:"body:name" validate:"required" description:"User full name"`
}

// Request schema using only json tag (fallback parsing)
type SimpleUserRequest struct {
	// These will be parsed from JSON body using json tag names
	Email    string `json:"email" validate:"required,email" description:"User email"`
	Password string `json:"password" validate:"required,min=6" description:"User password"`
	Name     string `json:"name" validate:"required" description:"User full name"`
	Age      int    `json:"age" validate:"gte=18" description:"User age"`
	IsActive bool   `json:"is_active" description:"User active status"`
}

// UserResponse represents user data with validation
type UserResponse struct {
	ID        int       `json:"id" validate:"required" description:"User ID"`
	Email     string    `json:"email" validate:"required,email" description:"User email"`
	Name      string    `json:"name" validate:"required" description:"User name"`
	Role      string    `json:"role" validate:"required,oneof=admin user" description:"User role"`
	IsActive  bool      `json:"is_active" description:"User active status"`
	OrgID     int       `json:"org_id" validate:"required" description:"Organization ID"`
	CreatedAt time.Time `json:"created_at" validate:"required" description:"Account creation date"`
}

// ErrorResponse schema for error responses
type ErrorResponse struct {
	Error   string `json:"error" validate:"required" description:"Error message"`
	Details string `json:"details,omitempty" description:"Error details"`
	Code    int    `json:"code" validate:"required" description:"HTTP status code"`
}

// LoginResponse schema for login responses
type LoginResponse struct {
	Token     string       `json:"token" validate:"required" description:"JWT token"`
	User      UserResponse `json:"user" validate:"required" description:"User information"`
	ExpiresAt time.Time    `json:"expires_at" validate:"required" description:"Token expiration time"`
}

// Handler
type AuthHandler struct{}

// Handler with request parsing (no response validation)
func (h *AuthHandler) Login(c *fiber.Ctx, req *LoginRequest) error {
	// req is automatically parsed and validated
	return c.JSON(fiber.Map{
		"message": "Login successful",
		"email":   req.Email,
		"token":   "jwt_token_here",
	})
}

// Handler with request parsing and response validation
// This demonstrates the complete flow: parse request -> validate request -> execute handler -> validate response
func (h *AuthHandler) Register(c *fiber.Ctx, req *RegisterRequest) (interface{}, error) {
	// req is automatically parsed and validated
	// return data and error for automatic response formatting and validation
	return UserResponse{
		ID:        1,
		Email:     req.Email,
		Name:      req.Name,
		Role:      "user",
		IsActive:  true,
		OrgID:     1,
		CreatedAt: time.Now(),
	}, nil
}

// Handler with request parsing and response validation
func (h *AuthHandler) LoginWithValidation(c *fiber.Ctx, req *LoginRequest) (interface{}, error) {
	// req is automatically parsed and validated
	// return data and error for automatic response formatting and validation
	return LoginResponse{
		Token: "jwt_token_here",
		User: UserResponse{
			ID:        1,
			Email:     req.Email,
			Name:      "Example User",
			Role:      "user",
			IsActive:  true,
			OrgID:     1,
			CreatedAt: time.Now(),
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

// ListUsers demonstrates parsing from query parameters and headers
func (h *AuthHandler) ListUsers(c *fiber.Ctx, req *UserFilterRequest) (interface{}, error) {
	return fiber.Map{
		"users": []UserResponse{
			{
				ID:        1,
				Email:     "user1@example.com",
				Name:      "User 1",
				Role:      "user",
				IsActive:  true,
				OrgID:     1,
				CreatedAt: time.Now(),
			},
			{
				ID:        2,
				Email:     "user2@example.com",
				Name:      "User 2",
				Role:      "admin",
				IsActive:  true,
				OrgID:     1,
				CreatedAt: time.Now(),
			},
		},
		"pagination": fiber.Map{
			"page":  req.Page,
			"limit": req.Limit,
			"total": 2,
		},
		"filters": fiber.Map{
			"search":    req.Search,
			"sort_by":   req.SortBy,
			"sort_desc": req.SortDesc,
		},
		"auth": fiber.Map{
			"authorization": req.Authorization,
			"session_id":    req.SessionID,
		},
	}, nil
}

// GetUser demonstrates smart parsing (auto-detect source) with response validation
func (h *AuthHandler) GetUser(c *fiber.Ctx, req *GetUserRequest) (interface{}, error) {
	// Simulate user not found
	if req.UserID == 999 {
		return nil, fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	return UserResponse{
		ID:        req.UserID,
		Email:     "user@example.com",
		Name:      "Example User",
		Role:      "user",
		IsActive:  true,
		OrgID:     1,
		CreatedAt: time.Now(),
	}, nil
}

// CreateUser demonstrates parsing from path, query, headers, and body with response validation
func (h *AuthHandler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
	return UserResponse{
		ID:        1,
		Email:     req.Email,
		Name:      req.Name,
		Role:      req.Role,
		IsActive:  req.IsActive,
		OrgID:     req.OrgID,
		CreatedAt: time.Now(),
	}, nil
}

// Simple handler without request parsing
func (h *AuthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "timestamp": time.Now()})
}

type UserHandler struct{}

func (h *UserHandler) CreateSimpleUser(c *fiber.Ctx, req *SimpleUserRequest) (interface{}, error) {
	return UserResponse{
		ID:        2,
		Email:     req.Email,
		Name:      req.Name,
		Role:      "user", // Default role
		IsActive:  req.IsActive,
		OrgID:     1, // Default org
		CreatedAt: time.Now(),
	}, nil
}

func (h *UserHandler) CreateUserFromMap(c *fiber.Ctx) (interface{}, error) {
	// Example of parsing from map
	userData := map[string]interface{}{
		"email":     "john@example.com",
		"password":  "secret123",
		"name":      "John Doe",
		"age":       25,
		"is_active": true,
	}

	var req SimpleUserRequest
	if err := autofiber.ParseFromMap(userData, &req); err != nil {
		return nil, err
	}

	return UserResponse{
		ID:        3,
		Email:     req.Email,
		Name:      req.Name,
		Role:      "user",
		IsActive:  req.IsActive,
		OrgID:     1,
		CreatedAt: time.Now(),
	}, nil
}

// Custom validation function for strong password
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check if password contains at least one uppercase letter, one lowercase letter, and one number
	hasUpper := false
	hasLower := false
	hasNumber := false

	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= 'a' && char <= 'z' {
			hasLower = true
		} else if char >= '0' && char <= '9' {
			hasNumber = true
		}
	}

	return hasUpper && hasLower && hasNumber
}

func main() {
	// Create AutoFiber app with docs configuration
	app := autofiber.New(fiber.Config{
		EnablePrintRoutes: true,
	}).
		WithDocsInfo(autofiber.OpenAPIInfo{
			Title:       "AutoFiber Complete Flow Example",
			Description: "Demonstrating complete flow: parse request -> validate request -> execute handler -> validate response",
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

	// Add custom validator with custom validation rules
	validator := autofiber.GetValidator()
	validator.RegisterValidation("strong_password", validateStrongPassword)

	// Add Fiber logger middleware
	app.Use(logger.New())

	handler := &AuthHandler{}

	// Register routes with auto-parse and documentation
	app.Post("/login", handler.Login,
		autofiber.WithRequestSchema(LoginRequest{}),
		autofiber.WithDescription("Authenticate user and return JWT token (no response validation)"),
		autofiber.WithTags("auth", "authentication"),
	)

	// Route with complete flow: parse request -> validate request -> execute handler -> validate response
	app.Post("/login-with-validation", handler.LoginWithValidation,
		autofiber.WithRequestSchema(LoginRequest{}),
		autofiber.WithResponseSchema(LoginResponse{}),
		autofiber.WithDescription("Authenticate user with response validation (complete flow demonstration)"),
		autofiber.WithTags("auth", "authentication"),
	)

	app.Post("/register", handler.Register,
		autofiber.WithRequestSchema(RegisterRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Register a new user account (with response validation)"),
		autofiber.WithTags("auth", "user"),
	)

	// Route with query parameters and headers using parse tag
	app.Get("/users", handler.ListUsers,
		autofiber.WithRequestSchema(UserFilterRequest{}),
		autofiber.WithDescription("List users with filtering, pagination, and authentication"),
		autofiber.WithTags("user", "admin"),
	)

	// Route with smart parsing (auto-detect source) and response validation
	app.Get("/users/:user_id", handler.GetUser,
		autofiber.WithRequestSchema(GetUserRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Get user by ID with smart parsing and response validation"),
		autofiber.WithTags("user", "admin"),
	)

	// Route with path parameter, query parameters, headers, and body using parse tag with response validation
	app.Post("/organizations/:org_id/users", handler.CreateUser,
		autofiber.WithRequestSchema(CreateUserRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Create a new user in an organization (complete flow with response validation)"),
		autofiber.WithTags("user", "admin"),
	)

	app.Get("/health", handler.Health,
		autofiber.WithDescription("Health check endpoint"),
		autofiber.WithTags("system"),
	)

	// Serve API documentation
	app.ServeDocs("/docs")
	app.ServeSwaggerUI("/swagger", "/docs")

	// Start server with log
	log.Println("Server is running at http://localhost:3000")
	log.Println("API Documentation: http://localhost:3000/docs")
	log.Println("Swagger UI: http://localhost:3000/swagger")
	log.Println("")
	log.Println("Complete Flow Examples:")
	log.Println("- POST /register: Parse request -> Validate request -> Execute handler -> Validate response")
	log.Println("- POST /login-with-validation: Same complete flow with login response")
	log.Println("- GET /users/:user_id: Smart parsing with response validation")
	log.Println("- POST /organizations/:org_id/users: Multi-source parsing with response validation")
	app.Listen(":3000")
}
