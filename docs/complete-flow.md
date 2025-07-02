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
    OrgID    int    `parse:"path:org_id" validate:"required"`
    Role     string `parse:"query:role" validate:"required,oneof=admin user"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Name     string `json:"name" validate:"required"`
}
```

**What happens**:

- `OrgID` is extracted from URL path `/organizations/:org_id/users`
- `Role` is extracted from query string `?role=admin`
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
// Recommended handler signature for AutoFiber:
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    user := UserResponse{
        ID:        1,
        Email:     req.Email,
        Name:      req.Name,
        Role:      req.Role,
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
  "created_at": "2024-01-15T10:30:00Z"
}
```

## Handler Signatures

AutoFiber requires specific handler signatures for proper functionality:

### Required: Complete Flow (Request Parsing + Response Validation)

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

### Required: Simple Handler (No Request Parsing)

```go
func (h *Handler) SimpleHandler(c *fiber.Ctx) (interface{}, error) {
    result := ResponseSchema{
        ID:   1,
        Name: "Default User",
        // ... other fields
    }
    return result, nil
}
```

### For Health Check or Custom Response Only

```go
func (h *Handler) Health(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok", "timestamp": time.Now()})
}
```

### ⚠️ NOT Supported (Will Cause Panic)

```go
// Do NOT use this pattern - AutoFiber requires (interface{}, error) return
func (h *Handler) BadHandler(c *fiber.Ctx, req *RequestSchema) error {
    return c.JSON(...)
}

// Do NOT use this pattern either
func (h *Handler) AnotherBadHandler(c *fiber.Ctx, req *RequestSchema) {
    // No return statement
}
```

> **Important:** AutoFiber requires handlers to return `(interface{}, error)` for automatic JSON marshaling and response validation. The old signature `func(c *fiber.Ctx, req *T) error` is no longer supported and will cause a panic.

## Response Validation

When you specify `WithResponseSchema()`, AutoFiber automatically validates the response:

```go
type UserResponse struct {
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Name      string    `json:"name" validate:"required"`
    Role      string    `json:"role" validate:"required,oneof=admin user"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
}

app.Post("/users", handler.CreateUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
    autofiber.WithResponseSchema(UserResponse{}),
    autofiber.WithDescription("Create a new user"),
)

// You can also use generic response schemas for consistent API responses:
type APIResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}

app.Get("/user", handler.GetUser, autofiber.WithResponseSchema(APIResponse[UserResponse]{}))

func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    user := UserResponse{
        ID:        1,
        Email:     req.Email,
        Name:      req.Name,
        Role:      "user",
        CreatedAt: time.Now(),
    }
    // Response will be automatically validated against UserResponse schema
    return user, nil
}
```

If response validation fails, AutoFiber returns a 500 error with validation details.

## Parse Tag

AutoFiber uses a unified `parse` tag to specify where each field should be parsed from:

### Parse Tag Format

```go
type Request struct {
    // Basic format: parse:"source:key"
    UserID int `parse:"path:user_id" validate:"required"`
    Page   int `parse:"query:page" validate:"gte=1"`
    Token  string `parse:"header:Authorization" validate:"required"`
    Email  string `parse:"body:email" validate:"required,email"`

    // With options: parse:"source:key,required,default:value"
    Limit   int  `parse:"query:limit,required,default:10"`
    Active  bool `parse:"query:active,default:true"`
    Role    string `parse:"query:role,required,default:user"`
}
```

### Supported Sources

- `path` - URL path parameters (`/users/:id`)
- `query` - Query string parameters (`?page=1&limit=10`)
- `header` - HTTP headers (`Authorization: Bearer token`)
- `cookie` - Cookies (`session_id=abc123`)
- `form` - Form data (`multipart/form-data`)
- `body` - JSON body (for POST/PUT/PATCH requests)
- `auto` - Smart detection based on HTTP method

### Parse Tag Options

- `required` - Field is required (returns 422 if missing)
- `default:value` - Default value if field is empty

## Auto Parsing

When no `parse` tag is specified, AutoFiber automatically detects the best source:

### HTTP Method Priority

- **GET**: `path` → `query` (no body parsing)
- **POST/PUT/PATCH**: `path` → `query` → `body`
- **DELETE**: `path` → `query` (no body parsing)

### Auto Parsing Example

```go
type AutoRequest struct {
    // These will be auto-detected based on HTTP method
    UserID int    `json:"user_id" validate:"required" description:"User ID"`
    Page   int    `json:"page" validate:"gte=1" description:"Page number"`
    Email  string `json:"email" validate:"required,email" description:"User email"`
    Name   string `json:"name" validate:"required" description:"User name"`
}

// For GET /users/:user_id?page=1
// - user_id: parsed from path
// - page: parsed from query
// - email, name: not parsed (GET doesn't parse body)

// For POST /users/:user_id?page=1
// - user_id: parsed from path
// - page: parsed from query
// - email, name: parsed from JSON body
```

## JSON Tag Support

AutoFiber supports `json` tag for field aliasing:

```go
type Request struct {
    // Use json tag to alias field names
    Email    string `json:"user_email" parse:"body:email" validate:"required,email"`
    Password string `json:"user_password" parse:"body:password" validate:"required,min=6"`
    Name     string `json:"full_name" parse:"body:name" validate:"required"`
}
```

## Map and Interface Parsing

AutoFiber provides utilities to parse structs from maps and interfaces:

### Parse From Map

```go
// Parse from map[string]interface{}
userData := map[string]interface{}{
    "email":     "john@example.com",
    "password":  "secret123",
    "name":      "John Doe",
    "age":       25,
    "is_active": true,
}

var req SimpleUserRequest
if err := autofiber.ParseFromMap(userData, &req); err != nil {
    return err
}
```

### Parse From Interface

```go
// Parse from any interface{} (map, struct, etc.)
data := map[string]string{
    "email": "john@example.com",
    "name":  "John Doe",
}

var req SimpleUserRequest
if err := autofiber.ParseFromInterface(data, &req); err != nil {
    return err
}
```

## Request Validation

Use struct tags for validation:

```go
type UserRequest struct {
    Email     string `parse:"body:email,required" description:"User email"`
    Password  string `parse:"body:password,required" description:"User password"`
    Age       int    `parse:"body:age" validate:"gte=18,lte=100" description:"User age"`
    Role      string `parse:"body:role,required" validate:"oneof=admin user guest" description:"User role"`
    IsActive  bool   `parse:"body:is_active" description:"User active status"`
}
```

## Route Options

Configure routes with flexible options:

```go
    app.Post("/users", handler.CreateUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
    autofiber.WithDescription("Create a new user account"),
    autofiber.WithTags("users", "admin"),
    autofiber.WithMiddleware(authMiddleware, loggingMiddleware),
)
```

## API Documentation

AutoFiber automatically generates OpenAPI 3.0 specification and serves Swagger UI:

```go
app := autofiber.NewWithOptions(
    fiber.Config{},
    autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
        Title:       "My API",
        Description: "API description",
        Version:     "1.0.0",
        Contact: &autofiber.OpenAPIContact{
            Name:  "API Team",
            Email: "team@example.com",
        },
    }),
)

// Serve documentation
app.ServeDocs("/docs")           // OpenAPI JSON at /docs
app.ServeSwaggerUI("/swagger", "/docs") // Swagger UI at /swagger
```

### Smart Documentation Generation

AutoFiber automatically generates proper OpenAPI documentation from your parse tags and auto parsing:

```go
type CreateUserRequest struct {
    // Path parameter - appears in "Parameters" section
    OrgID int `parse:"path:org_id,required" description:"Organization ID"`

    // Query parameters - appear in "Parameters" section
    Role     string `parse:"query:role,required" description:"User role"`
    IsActive bool   `parse:"query:active,default:true" description:"User active status"`

    // Header parameters - appear in "Parameters" section
    APIKey string `parse:"header:X-API-Key,required" description:"API key"`

    // Body fields - appear in "Request Body" section
    Email    string `parse:"body:email,required" description:"User email"`
    Password string `parse:"body:password,required" description:"User password"`
    Name     string `parse:"body:name,required" description:"User full name"`
}

type AutoRequest struct {
    // Auto-detected fields - appear in appropriate sections
    UserID int    `json:"user_id" validate:"required" description:"User ID"`
    Page   int    `json:"page" validate:"gte=1" description:"Page number"`
    Email  string `json:"email" validate:"required,email" description:"User email"`
    Name   string `json:"name" validate:"required" description:"User name"`
}
```

**Generated OpenAPI Documentation:**

- **Parameters**: Path, query, header, cookie parameters
- **Request Body**: Body fields and auto-detected body fields
- **Response Schema**: Response validation schema
- **Validation**: Required fields, format validation, enum values
- **Descriptions**: Field descriptions from struct tags
- **Default Values**: From `default` option in parse tag

### Documentation Features

- **Automatic Parameter Detection**: Path, query, header, cookie parameters
- **Request Body Schema**: Body fields and auto-detected body fields
- **Response Schema**: Response validation schema
- **JSON Tag Support**: Field aliasing in documentation
- **Validation Rules**: Required fields, format validation, enum values
- **Field Descriptions**: From `description` struct tags
- **Default Values**: From `default` option in parse tag
- **Type Safety**: Proper OpenAPI types (string, integer, boolean, etc.)
- **Interactive Testing**: Swagger UI for testing APIs directly
- **Security Schemes**: Bearer token authentication support

## Error Handling

AutoFiber provides clear error responses:

- **400 Bad Request**: Parse errors (invalid JSON, type conversion)
- **422 Unprocessable Entity**: Validation errors (missing required fields, format validation)
- **500 Internal Server Error**: Response validation errors

## Examples

See the `example/` directory for complete working examples:

- **Complete Flow**: Parse request → Validate request → Execute handler → Validate response
- **Request Parsing**: Parse from multiple sources (body, query, path, headers, cookies)
- **Response Validation**: Validate response data before sending
- **Auto Parsing**: Automatic field source detection
- **JSON Tag**: Field aliasing
- **Map Parsing**: Parsing structs from maps and interfaces
- **API Documentation**: OpenAPI and Swagger UI
- **Custom Validation**: Custom validation functions

## Run Example

```bash
go run example/main.go
```

Then visit:

- API: http://localhost:3000
- Swagger UI: http://localhost:3000/swagger
- OpenAPI JSON: http://localhost:3000/docs

### Example Endpoints

The example includes endpoints demonstrating the complete flow:

- `POST /register` - Parse request → Validate request → Execute handler → Validate response
- `POST /login-with-validation` - Complete flow with login response validation
- `GET /users/:user_id` - Smart parsing with response validation
- `POST /organizations/:org_id/users` - Multi-source parsing with response validation

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

- Use generic response wrappers (e.g., `APIResponse[T]`) for consistent API responses and OpenAPI documentation. The OpenAPI spec will reference the correct schema name (e.g., `APIResponse_User`).
- Only POST, PUT, PATCH methods generate a request body in OpenAPI. GET, DELETE, HEAD, OPTIONS do not, even if a request schema is provided.
