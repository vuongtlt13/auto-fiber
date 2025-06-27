# Complete Request/Response Flow

This guide explains the complete flow in AutoFiber: Parse Request → Validate Request → Execute Handler → Validate Response → Return JSON.

## Table of Contents

- [Overview](#overview)
- [Flow Details](#flow-details)
- [Handler Signatures](#handler-signatures)
- [Response Validation](#response-validation)
- [Error Handling](#error-handling)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Overview

AutoFiber provides a complete request/response flow similar to FastAPI:

```
Parse Request → Validate Request → Execute Handler → Validate Response → Return JSON
```

This flow ensures that:

1. Incoming data is properly parsed from multiple sources
2. Data is validated against your schema
3. Your business logic is executed
4. Response data is validated before sending
5. Clean JSON is returned to the client

## Flow Details

### 1. Parse Request

AutoFiber automatically parses data from multiple sources based on your struct tags:

```go
type CreateUserRequest struct {
    // Path parameter
    OrgID int `parse:"path:org_id" validate:"required"`

    // Query parameters
    Role     string `parse:"query:role" validate:"required,oneof=admin user"`
    IsActive bool   `parse:"query:active" validate:"omitempty"`

    // Headers
    APIKey string `parse:"header:X-API-Key" validate:"required"`

    // Body fields
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Name     string `json:"name" validate:"required"`
}
```

**What happens**:

- `OrgID` is extracted from URL path `/organizations/:org_id/users`
- `Role` and `IsActive` are extracted from query string `?role=admin&active=true`
- `APIKey` is extracted from header `X-API-Key: your-api-key`
- `Email`, `Password`, `Name` are extracted from JSON body

### 2. Validate Request

Parsed data is validated against your struct tags:

```go
// Validation rules are checked
Email:    "user@example.com" // ✓ Valid email
Password: "password123"      // ✓ At least 6 characters
Name:     "John Doe"         // ✓ Required field present
Role:     "admin"            // ✓ One of allowed values
```

**What happens**:

- Each field is validated according to its `validate` tag
- If validation fails, a 422 Unprocessable Entity error is returned
- Validation errors include detailed field-specific messages

### 3. Execute Handler

Your business logic is executed with the parsed and validated data:

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    // Business logic here
    user := UserResponse{
        ID:        1,
        Email:     req.Email,
        Name:      req.Name,
        Role:      req.Role,
        IsActive:  req.IsActive,
        OrgID:     req.OrgID,
        CreatedAt: time.Now(),
    }

    return user, nil
}
```

**What happens**:

- Your handler receives the parsed and validated request
- You can focus on business logic without worrying about parsing/validation
- Return data and error for automatic response handling

### 4. Validate Response

Response data is validated against your response schema:

```go
type UserResponse struct {
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Name      string    `json:"name" validate:"required"`
    Role      string    `json:"role" validate:"required,oneof=admin user"`
    IsActive  bool      `json:"is_active"`
    OrgID     int       `json:"org_id" validate:"required"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
}
```

**What happens**:

- Response data is validated against the response schema
- If validation fails, a 500 Internal Server Error is returned
- This ensures data integrity and API consistency

### 5. Return JSON

Validated response is automatically serialized to JSON:

```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "John Doe",
  "role": "admin",
  "is_active": true,
  "org_id": 123,
  "created_at": "2024-01-15T10:30:00Z"
}
```

## Handler Signatures

AutoFiber supports different handler signatures depending on your needs:

### Simple Handler (No Request Parsing)

```go
func (h *Handler) SimpleHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "message": "Hello World",
        "timestamp": time.Now(),
    })
}
```

**When to use**:

- For simple endpoints that don't need request parsing
- For health checks, status endpoints
- When you want full control over the response

### Handler with Request Parsing (No Response Validation)

```go
func (h *Handler) RequestHandler(c *fiber.Ctx, req *RequestSchema) error {
    // Process the parsed request
    result := processData(req)

    return c.JSON(fiber.Map{
        "data": result,
        "processed_at": time.Now(),
    })
}
```

**When to use**:

- When you need request parsing but don't want response validation
- For endpoints that return dynamic data structures
- When you want to handle response formatting manually

### Handler with Complete Flow (Request Parsing + Response Validation)

```go
func (h *Handler) CompleteHandler(c *fiber.Ctx, req *RequestSchema) (interface{}, error) {
    // Business logic here
    result := ResponseSchema{
        ID:   1,
        Name: req.Name,
        // ... other fields
    }

    // Return data and error - response will be automatically validated
    return result, nil
}
```

**When to use**:

- For most API endpoints
- When you want automatic response validation
- For consistent API responses
- When following the complete flow pattern

## Response Validation

Response validation ensures that your API returns consistent and valid data.

### Enabling Response Validation

```go
app.Post("/users", handler.CreateUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
    autofiber.WithResponseSchema(UserResponse{}), // This enables response validation
    autofiber.WithDescription("Create a new user"),
)
```

### Response Schema Definition

```go
type UserResponse struct {
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Name      string    `json:"name" validate:"required"`
    Role      string    `json:"role" validate:"required,oneof=admin user"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
}
```

### Validation Rules for Responses

```go
type ComprehensiveResponse struct {
    // Required fields
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`

    // Optional fields with validation
    Bio       string    `json:"bio,omitempty" validate:"omitempty,max=500"`
    Avatar    string    `json:"avatar,omitempty" validate:"omitempty,url"`

    // Nested objects
    Profile   *Profile  `json:"profile,omitempty" validate:"omitempty"`

    // Arrays
    Tags      []string  `json:"tags,omitempty" validate:"omitempty,dive,min=1"`

    // Timestamps
    CreatedAt time.Time `json:"created_at" validate:"required"`
    UpdatedAt time.Time `json:"updated_at" validate:"required"`
}
```

### Response Validation Errors

If response validation fails, AutoFiber returns a 500 error:

```json
{
  "error": "Response validation failed",
  "details": "Field 'email' failed validation: 'invalid-email' is not a valid email"
}
```

**Common causes**:

- Missing required fields
- Invalid data types
- Validation rule violations
- Nested object validation failures

## Error Handling

AutoFiber provides clear error responses for different scenarios:

### Parse Errors (400 Bad Request)

```json
{
  "error": "Invalid request",
  "details": "user_id (path): invalid integer value 'abc'"
}
```

**Causes**:

- Invalid JSON in request body
- Type conversion failures (string to int, etc.)
- Missing required path parameters

### Validation Errors (422 Unprocessable Entity)

```json
{
  "error": "Validation failed",
  "details": "Field 'email' failed validation: 'invalid-email' is not a valid email"
}
```

**Causes**:

- Missing required fields
- Invalid email formats
- Value out of range
- Enum value violations

### Response Validation Errors (500 Internal Server Error)

```json
{
  "error": "Response validation failed",
  "details": "Field 'id' failed validation: '0' is not greater than or equal to 1"
}
```

**Causes**:

- Handler returned invalid data
- Missing required response fields
- Response validation rule violations

### Handler Errors

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    // Return custom error
    if req.Email == "admin@example.com" {
        return nil, fiber.NewError(fiber.StatusConflict, "Email already exists")
    }

    // Return business logic error
    if err := validateBusinessRules(req); err != nil {
        return nil, err
    }

    // Success case
    return UserResponse{...}, nil
}
```

## Examples

### Complete User Management API

```go
// Request schemas
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Name     string `json:"name" validate:"required"`
    Role     string `json:"role" validate:"required,oneof=admin user"`
}

type UpdateUserRequest struct {
    UserID int    `parse:"path:user_id" validate:"required,min=1"`
    Email  string `json:"email,omitempty" validate:"omitempty,email"`
    Name   string `json:"name,omitempty" validate:"omitempty,min=2"`
    Role   string `json:"role,omitempty" validate:"omitempty,oneof=admin user"`
}

type GetUsersRequest struct {
    Page     int    `parse:"query:page" validate:"omitempty,gte=1"`
    Limit    int    `parse:"query:limit" validate:"omitempty,gte=1,lte=100"`
    Search   string `parse:"query:search" validate:"omitempty,min=2"`
    Role     string `parse:"query:role" validate:"omitempty,oneof=admin user"`
}

// Response schemas
type UserResponse struct {
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Name      string    `json:"name" validate:"required"`
    Role      string    `json:"role" validate:"required,oneof=admin user"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
    UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type UsersListResponse struct {
    Users []UserResponse `json:"users" validate:"required"`
    Page  int            `json:"page" validate:"required,gte=1"`
    Limit int            `json:"limit" validate:"required,gte=1"`
    Total int            `json:"total" validate:"required,gte=0"`
}

// Handlers
type UserHandler struct{}

func (h *UserHandler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    // Business logic
    user := UserResponse{
        ID:        1,
        Email:     req.Email,
        Name:      req.Name,
        Role:      req.Role,
        IsActive:  true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    return user, nil
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx, req *UpdateUserRequest) (interface{}, error) {
    // Business logic
    user := UserResponse{
        ID:        req.UserID,
        Email:     req.Email,
        Name:      req.Name,
        Role:      req.Role,
        IsActive:  true,
        CreatedAt: time.Now().Add(-24 * time.Hour), // Simulate existing user
        UpdatedAt: time.Now(),
    }

    return user, nil
}

func (h *UserHandler) GetUsers(c *fiber.Ctx, req *GetUsersRequest) (interface{}, error) {
    // Business logic
    users := []UserResponse{
        {
            ID:        1,
            Email:     "user1@example.com",
            Name:      "User 1",
            Role:      "user",
            IsActive:  true,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
        {
            ID:        2,
            Email:     "admin@example.com",
            Name:      "Admin User",
            Role:      "admin",
            IsActive:  true,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }

    return UsersListResponse{
        Users: users,
        Page:  req.Page,
        Limit: req.Limit,
        Total: len(users),
    }, nil
}

// Route registration
func main() {
    app := autofiber.New()
    handler := &UserHandler{}

    // Create user with complete flow
    app.Post("/users", handler.CreateUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
        autofiber.WithDescription("Create a new user"),
        autofiber.WithTags("users"),
    )

    // Update user with complete flow
    app.Put("/users/:user_id", handler.UpdateUser,
        autofiber.WithRequestSchema(UpdateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
        autofiber.WithDescription("Update an existing user"),
        autofiber.WithTags("users"),
    )

    // Get users with complete flow
    app.Get("/users", handler.GetUsers,
        autofiber.WithRequestSchema(GetUsersRequest{}),
        autofiber.WithResponseSchema(UsersListResponse{}),
        autofiber.WithDescription("List users with pagination and filtering"),
        autofiber.WithTags("users"),
    )

    app.Listen(":3000")
}
```

### Authentication Flow

```go
// Request schemas
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type AuthenticatedRequest struct {
    Authorization string `parse:"header:Authorization" validate:"required"`
    UserID        int    `parse:"path:user_id" validate:"required,min=1"`
}

// Response schemas
type LoginResponse struct {
    Token     string       `json:"token" validate:"required"`
    User      UserResponse `json:"user" validate:"required"`
    ExpiresAt time.Time    `json:"expires_at" validate:"required"`
}

// Handlers
type AuthHandler struct{}

func (h *AuthHandler) Login(c *fiber.Ctx, req *LoginRequest) (interface{}, error) {
    // Authentication logic
    token := "jwt_token_here"
    user := UserResponse{
        ID:        1,
        Email:     req.Email,
        Name:      "John Doe",
        Role:      "user",
        IsActive:  true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    return LoginResponse{
        Token:     token,
        User:      user,
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }, nil
}

func (h *AuthHandler) GetProfile(c *fiber.Ctx, req *AuthenticatedRequest) (interface{}, error) {
    // Get user profile logic
    user := UserResponse{
        ID:        req.UserID,
        Email:     "user@example.com",
        Name:      "John Doe",
        Role:      "user",
        IsActive:  true,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    return user, nil
}

// Route registration
func main() {
    app := autofiber.New()
    handler := &AuthHandler{}

    app.Post("/login", handler.Login,
        autofiber.WithRequestSchema(LoginRequest{}),
        autofiber.WithResponseSchema(LoginResponse{}),
        autofiber.WithDescription("Authenticate user and return JWT token"),
        autofiber.WithTags("auth"),
    )

    app.Get("/users/:user_id/profile", handler.GetProfile,
        autofiber.WithRequestSchema(AuthenticatedRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
        autofiber.WithDescription("Get user profile (authenticated)"),
        autofiber.WithTags("users", "auth"),
    )

    app.Listen(":3000")
}
```

## Best Practices

### 1. Use Consistent Handler Signatures

```go
// For most API endpoints, use the complete flow signature
func (h *Handler) CreateResource(c *fiber.Ctx, req *CreateRequest) (interface{}, error) {
    // Business logic
    return ResponseSchema{...}, nil
}

// For simple endpoints, use simple signature
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok"})
}
```

### 2. Define Clear Request/Response Schemas

```go
// Good: Clear, focused schemas
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Name     string `json:"name" validate:"required"`
}

type UserResponse struct {
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Name      string    `json:"name" validate:"required"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
}

// Avoid: Mixed concerns in one schema
type UserRequest struct {
    // Request fields
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`

    // Response fields (don't mix!)
    ID        int       `json:"id" validate:"required"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
}
```

### 3. Handle Errors Gracefully

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    // Check business rules
    if err := h.validateBusinessRules(req); err != nil {
        return nil, fiber.NewError(fiber.StatusBadRequest, err.Error())
    }

    // Check for conflicts
    if h.userExists(req.Email) {
        return nil, fiber.NewError(fiber.StatusConflict, "User already exists")
    }

    // Success case
    user := UserResponse{...}
    return user, nil
}
```

### 4. Use Response Validation for Data Integrity

```go
// Always validate responses for critical endpoints
app.Post("/users", handler.CreateUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
    autofiber.WithResponseSchema(UserResponse{}), // Ensures data integrity
)
```

### 5. Document Your Endpoints

```go
app.Post("/users", handler.CreateUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
    autofiber.WithResponseSchema(UserResponse{}),
    autofiber.WithDescription("Create a new user account"),
    autofiber.WithTags("users", "admin"),
)
```

### 6. Test Your Complete Flow

```go
func TestCreateUserFlow(t *testing.T) {
    app := autofiber.New()
    handler := &UserHandler{}

    app.Post("/users", handler.CreateUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
    )

    // Test valid request
    req := httptest.NewRequest("POST", "/users", strings.NewReader(`{
        "email": "user@example.com",
        "password": "password123",
        "name": "John Doe"
    }`))
    req.Header.Set("Content-Type", "application/json")

    resp, err := app.Test(req)
    if err != nil {
        t.Fatal(err)
    }

    if resp.StatusCode != fiber.StatusOK {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }

    var response UserResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        t.Fatal(err)
    }

    if response.Email != "user@example.com" {
        t.Errorf("Expected email 'user@example.com', got '%s'", response.Email)
    }
}
```

This complete flow ensures that your AutoFiber applications are robust, consistent, and maintainable.
