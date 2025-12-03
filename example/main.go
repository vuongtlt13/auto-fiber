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
	BirthDate time.Time `json:"birthDate" description:"User birth date"`
}

// UserFilterRequest demonstrates parsing from multiple sources using parse tag
type UserFilterRequest struct {
	// Query parameters
	Page     int    `parse:"query:page" validate:"gte=1" description:"Page number" example:"1"`
	Limit    int    `parse:"query:limit" validate:"gte=1,lte=100" description:"Items per page" example:"10"`
	Search   string `parse:"query:search" description:"Search term"`
	SortBy   string `parse:"query:sortBy" description:"Sort field" example:"name"`
	SortDesc bool   `parse:"query:sortDesc" description:"Sort descending"`

	// Headers
	Authorization string `parse:"header:Authorization" validate:"required" description:"Bearer token"`
	Accept        string `parse:"header:Accept" description:"Accept header"`

	// Cookies
	SessionID string `parse:"cookie:sessionId" description:"Session ID from cookie"`
}

// GetUserRequest demonstrates smart parsing (auto-detect source)
type GetUserRequest struct {
	// These will be auto-detected based on HTTP method
	UserID         int  `parse:"auto:userId" validate:"required" description:"User ID (auto-detected from path/query/body)"`
	IncludeProfile bool `parse:"auto:includeProfile" description:"Include user profile data"`
	IncludePosts   bool `parse:"auto:includePosts" description:"Include user posts"`

	// Headers
	Authorization string `parse:"header:Authorization" validate:"required" description:"Bearer token"`
}

// Request schema with parse tag and json tag support
type CreateUserRequest struct {
	// Path parameter
	OrgID int `parse:"path:orgId" validate:"required" description:"Organization ID"`

	// Query parameters
	Role     string `parse:"query:role" validate:"required,oneof=admin user" description:"User role"`
	IsActive bool   `parse:"query:active" description:"User active status"`

	// Headers
	APIKey string `parse:"header:X-API-Key" validate:"required" description:"API key"`

	// Body fields with json tag aliasing
	Email    string `json:"userEmail" parse:"body:email" validate:"required,email" description:"User email"`
	Password string `json:"userPassword" parse:"body:password" validate:"required,min=6" description:"User password"`
	Name     string `json:"fullName" parse:"body:name" validate:"required" description:"User full name"`
}

// Request schema using only json tag (fallback parsing)
type SimpleUserRequest struct {
	// These will be parsed from JSON body using json tag names
	Email    string `json:"email" validate:"required,email" description:"User email"`
	Password string `json:"password" validate:"required,min=6" description:"User password"`
	Name     string `json:"name" validate:"required" description:"User full name"`
	Age      int    `json:"age" validate:"gte=18" description:"User age"`
	IsActive bool   `json:"isActive" description:"User active status"`
}

// UserResponse represents user data with validation
type UserResponse struct {
	ID        int       `json:"id" validate:"required" description:"User ID"`
	Email     string    `json:"email" validate:"required,email" description:"User email"`
	Name      string    `json:"name" validate:"required" description:"User name"`
	Role      string    `json:"role" validate:"required,oneof=admin user" description:"User role"`
	IsActive  bool      `json:"isActive" description:"User active status"`
	OrgID     int       `json:"orgId" validate:"required" description:"Organization ID"`
	CreatedAt time.Time `json:"createdAt" validate:"required" description:"Account creation date"`
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
	ExpiresAt time.Time    `json:"expiresAt" validate:"required" description:"Token expiration time"`
}

// --- Generic APIResponse example ---
type APIResponse[T any] struct {
	Code    int    `json:"code" validate:"required"`
	Message string `json:"message" validate:"required"`
	Data    T      `json:"data"`
}

type User struct {
	ID   int    `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type UserList struct {
	Users []User `json:"users" validate:"required"`
}

// Handler
type AuthHandler struct{}

// Handler with request parsing (no response validation)
func (h *AuthHandler) Login(c *fiber.Ctx, req *LoginRequest) (interface{}, error) {
	// req is automatically parsed and validated
	return fiber.Map{
		"message": "Login successful",
		"email":   req.Email,
		"token":   "jwt_token_here",
	}, nil
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

// Handler for single user with generic response
func (h *AuthHandler) GetUserGeneric(c *fiber.Ctx) (interface{}, error) {
	return APIResponse[User]{
		Code:    200,
		Message: "success",
		Data:    User{ID: 1, Name: "Alice"},
	}, nil
}

// Handler for user list with generic response
func (h *AuthHandler) ListUsersGeneric(c *fiber.Ctx) (interface{}, error) {
	return APIResponse[UserList]{
		Code:    200,
		Message: "success",
		Data:    UserList{Users: []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}},
	}, nil
}

// Handler for single user with pointer generic response
func (h *AuthHandler) GetUserPointerGeneric(c *fiber.Ctx) (APIResponse[*User], error) {
	return APIResponse[*User]{
		Code:    200,
		Message: "success",
		Data:    &User{ID: 10, Name: "Pointer Alice"},
	}, nil
}

// Handler for single user with pointer to generic response
func (h *AuthHandler) GetUserPointerToGeneric(c *fiber.Ctx) (*APIResponse[*User], error) {
	return &APIResponse[*User]{
		Code:    200,
		Message: "success",
		Data:    &User{ID: 11, Name: "Pointer Bob"},
	}, nil
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

// Example structs to demonstrate ConvertRequestToOpenAPISchema and ConvertResponseToOpenAPISchema
type ExampleRequest struct {
	UserID   int    `parse:"body:userId" json:"userId" validate:"required"`
	UserName string `parse:"body:userName" json:"userName" validate:"required"`
	Token    string `parse:"header:authorization" json:"token"`
	Page     int    `parse:"query:page" json:"page"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	SkipMe   string `json:"" validate:"required"`
	SkipMe2  string `json:"," validate:"required"`
	NoTags   string
}

type ExampleResponse struct {
	ID         int       `json:"id" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	Email      string    `json:"email" validate:"required,email"`
	CreatedAt  time.Time `json:"createdAt" validate:"required"`
	IsActive   bool      `json:"isActive"`
	UserType   string    `json:"userType"`
	APIKey     string    `json:"apiKey"`
	HTTPStatus string    `json:"httpStatus"`
	SkipMe     string    `json:"-"`
}

// Handler demonstrating the new convert functions
func (h *UserHandler) ExampleConvertFunctions(c *fiber.Ctx) (interface{}, error) {
	dg := autofiber.NewDocsGenerator()

	req := ExampleRequest{
		UserID:   123,
		UserName: "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	resp := ExampleResponse{
		ID:         123,
		Name:       "John Doe",
		Email:      "john@example.com",
		CreatedAt:  time.Now(),
		IsActive:   true,
		UserType:   "admin",
		APIKey:     "abc123",
		HTTPStatus: "200",
		SkipMe:     "this will be skipped",
	}

	requestSchema := dg.ConvertRequestToOpenAPISchema(req)
	responseSchema := dg.ConvertResponseToOpenAPISchema(resp)

	return fiber.Map{
		"message": "Example of ConvertRequestToOpenAPISchema and ConvertResponseToOpenAPISchema",
		"request_schema": fiber.Map{
			"type":       requestSchema.Type,
			"properties": requestSchema.Properties,
			"required":   requestSchema.Required,
		},
		"response_schema": fiber.Map{
			"type":       responseSchema.Type,
			"properties": responseSchema.Properties,
			"required":   responseSchema.Required,
		},
		"explanation": fiber.Map{
			"request_fields_included":  []string{"userId", "userName", "email", "password"},
			"request_fields_skipped":   []string{"token", "page", "SkipMe", "SkipMe2", "NoTags"},
			"response_fields_included": []string{"id", "name", "email", "createdAt", "isActive", "userType", "apiKey", "httpStatus"},
			"response_fields_skipped":  []string{"SkipMe"},
		},
	}, nil
}

// --- Embedded Struct Example ---
type UserBase struct {
	ID    int    `json:"id" validate:"required" description:"User ID"`
	Name  string `json:"name" validate:"required" description:"User name"`
	Email string `json:"email" validate:"required,email" description:"User email"`
}

type Address struct {
	Street  string `json:"street" validate:"required" description:"Street address"`
	City    string `json:"city" validate:"required" description:"City name"`
	Country string `json:"country" validate:"required" description:"Country name"`
}

type CreateUserWithEmbeddedRequest struct {
	UserBase           // Embedded user fields
	Address            // Embedded address fields
	PhoneNumber string `json:"phoneNumber" validate:"required" description:"Phone number"`
	IsActive    bool   `json:"isActive" description:"User active status"`
}

type EmbeddedUserResponse struct {
	UserBase              // Embedded user fields
	Address               // Embedded address fields
	CreatedAt   time.Time `json:"created_at" description:"Creation timestamp"`
	PhoneNumber string    `json:"phoneNumber" description:"Phone number"`
}

func (h *UserHandler) CreateEmbeddedUser(c *fiber.Ctx, req *CreateUserWithEmbeddedRequest) (interface{}, error) {
	return EmbeddedUserResponse{
		UserBase:    req.UserBase,
		Address:     req.Address,
		CreatedAt:   time.Now(),
		PhoneNumber: req.PhoneNumber,
	}, nil
}

func (h *UserHandler) CreateEmbeddedUserGeneric(c *fiber.Ctx, req *CreateUserWithEmbeddedRequest) (*APIResponse[EmbeddedUserResponse], error) {
	// Create response with embedded structs
	response := &EmbeddedUserResponse{
		UserBase: UserBase{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		},
		Address: Address{
			Street:  req.Street,
			City:    req.City,
			Country: req.Country,
		},
		CreatedAt:   time.Now(),
		PhoneNumber: req.PhoneNumber,
	}

	return &APIResponse[EmbeddedUserResponse]{
		Code:    200,
		Message: "User created successfully with embedded structs",
		Data:    *response,
	}, nil
}

func main() {
	// Create AutoFiber app with docs configuration
	app := autofiber.New(
		fiber.Config{
			EnablePrintRoutes: true,
		},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:       "AutoFiber Complete Flow Example",
			Description: "Demonstrating complete flow: parse request -> validate request -> execute handler -> validate response",
			Version:     "1.0.0",
			Contact: &autofiber.OpenAPIContact{
				Name:  "AutoFiber Team",
				Email: "team@autofiber.com",
			},
		}),
	)

	// Add custom validator with custom validation rules
	validator := autofiber.GetValidator()
	validator.RegisterValidation("strong_password", validateStrongPassword)

	// Add Fiber logger middleware
	app.Use(logger.New())

	handler := &AuthHandler{}
	userHandler := &UserHandler{}

	// Create group for auth
	authGroup := app.Group("/auth")
	authGroup.Post("/login",
		handler.Login,
		autofiber.WithRequestSchema(LoginRequest{}),
		autofiber.WithDescription("Authenticate user and return JWT token (no response validation)"),
		autofiber.WithTags("auth", "authentication"),
	)
	authGroup.Post("/login-with-validation", handler.LoginWithValidation,
		autofiber.WithRequestSchema(LoginRequest{}),
		autofiber.WithResponseSchema(LoginResponse{}),
		autofiber.WithDescription("Authenticate user with response validation (complete flow demonstration)"),
		autofiber.WithTags("auth", "authentication"),
	)
	authGroup.Post("/register", handler.Register,
		autofiber.WithRequestSchema(RegisterRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Register a new user account (with response validation)"),
		autofiber.WithTags("auth", "user"),
	)
	authGroup.Get("/user-generic", handler.GetUserGeneric,
		autofiber.WithResponseSchema(APIResponse[User]{}),
		autofiber.WithDescription("Get a single user (generic response)"),
		autofiber.WithTags("user", "generic"),
	)
	authGroup.Get("/users-generic", handler.ListUsersGeneric,
		autofiber.WithResponseSchema(APIResponse[UserList]{}),
		autofiber.WithDescription("Get a list of users (generic response)"),
		autofiber.WithTags("user", "generic"),
	)
	authGroup.Get("/user-pointer-generic", handler.GetUserPointerGeneric,
		autofiber.WithResponseSchema(APIResponse[*User]{}),
		autofiber.WithDescription("Get a single user (pointer generic response)"),
		autofiber.WithTags("user", "generic", "pointer"),
	)
	authGroup.Get("/user-pointer-to-generic", handler.GetUserPointerToGeneric,
		autofiber.WithResponseSchema(&APIResponse[*User]{}),
		autofiber.WithDescription("Get a single user (pointer to generic response)"),
		autofiber.WithTags("user", "generic", "pointer"),
	)

	// Create group for user
	userGroup := app.Group("/users")
	userGroup.Get("/", handler.ListUsers,
		autofiber.WithRequestSchema(UserFilterRequest{}),
		autofiber.WithDescription("List users with filtering, pagination, and authentication"),
		autofiber.WithTags("user", "admin"),
	)
	userGroup.Get("/:user_id", handler.GetUser,
		autofiber.WithRequestSchema(GetUserRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Get user by ID with smart parsing and response validation"),
		autofiber.WithTags("user", "admin"),
	)
	userGroup.Post("/", userHandler.CreateSimpleUser,
		autofiber.WithRequestSchema(SimpleUserRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Create simple user (body only, json tag)"),
		autofiber.WithTags("user"),
	)
	userGroup.Post("/from-map", userHandler.CreateUserFromMap,
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Create user from map (manual parse)"),
		autofiber.WithTags("user", "example"),
	)

	// The route for creating a user in an organization remains outside the group due to its special path
	app.Post("/organizations/:org_id/users", handler.CreateUser,
		autofiber.WithRequestSchema(CreateUserRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Create a new user in an organization (complete flow with response validation)"),
		autofiber.WithTags("user", "admin"),
	)

	app.Get("/health",
		func(c *fiber.Ctx) (interface{}, error) {
			err := handler.Health(c)
			return nil, err
		},
		autofiber.WithDescription("Health check endpoint"),
		autofiber.WithTags("system"),
	)

	// Add at root
	app.Get("/schema-convert-example", userHandler.ExampleConvertFunctions,
		autofiber.WithDescription("Example demonstrating ConvertRequestToOpenAPISchema and ConvertResponseToOpenAPISchema functions"),
		autofiber.WithTags("example", "convert"),
	)

	// Add new route for embedded struct example
	app.Post("/embedded-users", userHandler.CreateEmbeddedUser,
		autofiber.WithRequestSchema(CreateUserWithEmbeddedRequest{}),
		autofiber.WithResponseSchema(EmbeddedUserResponse{}),
	)

	// Add new route for embedded struct + generic response example
	app.Post("/embedded-users-generic", userHandler.CreateEmbeddedUserGeneric,
		autofiber.WithRequestSchema(CreateUserWithEmbeddedRequest{}),
		autofiber.WithResponseSchema(&APIResponse[EmbeddedUserResponse]{}),
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
	log.Println("- POST /auth/register: Parse request -> Validate request -> Execute handler -> Validate response")
	log.Println("- POST /auth/login-with-validation: Same complete flow with login response")
	log.Println("- GET /users/:user_id: Smart parsing with response validation")
	log.Println("- POST /organizations/:org_id/users: Multi-source parsing with response validation")
	log.Println("Generic response endpoints: /auth/user-generic, /auth/users-generic")
	log.Fatal(app.Listen(":3000"))
}
