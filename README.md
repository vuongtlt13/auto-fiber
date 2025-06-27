# AutoFiber

A FastAPI-like wrapper for the Fiber web framework in Go, providing automatic request parsing, validation, and OpenAPI/Swagger documentation generation.

## Features

- **🔄 Complete Request/Response Flow**: Parse request → Validate request → Execute handler → Validate response → Return JSON
- **🧠 Smart Parsing**: Auto-detect the best source based on HTTP method (GET: path→query, POST: body→path→query)
- **🏷️ Unified Parse Tag**: Single `parse` tag with options like `required` and `default`
- **🗺️ Map/Interface Parsing**: Parse structs from maps, interfaces, and other data structures
- **✅ Request Validation**: Built-in validation using struct tags with `go-playground/validator`
- **✅ Response Validation**: Validate response data before sending to client
- **📚 Auto Documentation**: Generate OpenAPI 3.0 specification and Swagger UI
- **🔒 Type Safety**: Full type safety with Go generics
- **⚙️ Route Options**: Flexible route configuration with options pattern
- **🔌 Middleware Integration**: Seamless integration with Fiber middleware
- **🎯 Clean Architecture**: Modular design with separate concerns

## Installation

```sh
go get github.com/vuongtlt13/auto-fiber
```

## Quick Start

```go
package main

import (
    "time"
    "github.com/vuongtlt13/auto-fiber"
    "github.com/gofiber/fiber/v2"
)

// Request schema with parse tag
type CreateUserRequest struct {
    // Path parameter
    OrgID int `parse:"path:org_id,required" description:"Organization ID"`

    // Query parameters
    Role     string `parse:"query:role,required" description:"User role"`
    IsActive bool   `parse:"query:active,default:true" description:"User active status"`

    // Headers
    APIKey string `parse:"header:X-API-Key,required" description:"API key"`

    // Body fields
    Email    string `parse:"body:email,required" description:"User email"`
    Password string `parse:"body:password,required" description:"User password"`
    Name     string `parse:"body:name,required" description:"User full name"`
}

// Simple request using auto parsing (no parse tag needed)
type SimpleUserRequest struct {
    Email    string `json:"email" validate:"required,email" description:"User email"`
    Password string `json:"password" validate:"required,min=6" description:"User password"`
    Name     string `json:"name" validate:"required" description:"User full name"`
    Age      int    `json:"age" validate:"gte=18" description:"User age"`
    IsActive bool   `json:"is_active" description:"User active status"`
}

// Response schema with validation
type UserResponse struct {
    ID        int       `json:"id" validate:"required" description:"User ID"`
    Email     string    `json:"email" validate:"required,email" description:"User email"`
    Name      string    `json:"name" validate:"required" description:"User name"`
    Role      string    `json:"role" validate:"required,oneof=admin user" description:"User role"`
    IsActive  bool      `json:"is_active" description:"User active status"`
    OrgID     int       `json:"org_id" validate:"required" description:"Organization ID"`
    CreatedAt time.Time `json:"created_at" validate:"required" description:"Account creation date"`
}

type UserHandler struct{}

// Handler with complete flow: parse request → validate request → execute handler → validate response
func (h *UserHandler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    // 1. Request is automatically parsed and validated
    // 2. Business logic here
    user := UserResponse{
        ID:        1,
        Email:     req.Email,
        Name:      req.Name,
        Role:      req.Role,
        IsActive:  req.IsActive,
        OrgID:     req.OrgID,
        CreatedAt: time.Now(),
    }
    // 3. Response will be automatically validated before returning JSON
    return user, nil
}

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

func main() {
    app := autofiber.New().
        WithDocsInfo(autofiber.OpenAPIInfo{
            Title:       "AutoFiber API",
            Description: "A sample API with complete request/response flow",
            Version:     "1.0.0",
        })

    handler := &UserHandler{}

    // Register route with complete flow
    app.Post("/organizations/:org_id/users", handler.CreateUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}), // Enables response validation
        autofiber.WithDescription("Create a new user in an organization"),
        autofiber.WithTags("users", "admin"),
    )

    // Register route with auto parsing and response validation
    app.Post("/users/simple", handler.CreateSimpleUser,
        autofiber.WithRequestSchema(SimpleUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}), // Enables response validation
        autofiber.WithDescription("Create a simple user using auto parsing"),
        autofiber.WithTags("users"),
    )

    // Serve documentation
    app.ServeDocs("/docs")
    app.ServeSwaggerUI("/swagger", "/docs")

    app.Listen(":3000")
}
```

## Complete Request/Response Flow

AutoFiber provides a complete flow similar to FastAPI:

```
Parse Request → Validate Request → Execute Handler → Validate Response → Return JSON
```

### Flow Details

1. **Parse Request**: Automatically parse from multiple sources (body, query, path, headers, cookies)
2. **Validate Request**: Validate parsed data against struct tags
3. **Execute Handler**: Run your business logic
4. **Validate Response**: Validate response data before sending
5. **Return JSON**: Send validated response to client

### Handler Signatures

```go
// Simple handler (no request parsing)
func (h *Handler) SimpleHandler(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"message": "Hello"})
}

// Handler with request parsing (no response validation)
func (h *Handler) RequestHandler(c *fiber.Ctx, req *RequestSchema) error {
    return c.JSON(fiber.Map{"data": req})
}

// Handler with complete flow (request parsing + response validation)
func (h *Handler) CompleteHandler(c *fiber.Ctx, req *RequestSchema) (interface{}, error) {
    // Return data and error - response will be automatically validated
    return ResponseSchema{...}, nil
}
```

### Response Validation

When you specify `WithResponseSchema()`, AutoFiber automatically validates the response:

```go
// Response schema with validation
type UserResponse struct {
    ID        int       `json:"id" validate:"required" description:"User ID"`
    Email     string    `json:"email" validate:"required,email" description:"User email"`
    Name      string    `json:"name" validate:"required" description:"User name"`
    Role      string    `json:"role" validate:"required,oneof=admin user" description:"User role"`
    CreatedAt time.Time `json:"created_at" validate:"required" description:"Account creation date"`
}

// Route with response validation
app.Post("/users", handler.CreateUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
    autofiber.WithResponseSchema(UserResponse{}), // This enables response validation
    autofiber.WithDescription("Create a new user"),
)

// Handler that returns validated response
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
app := autofiber.New().
    WithDocsInfo(autofiber.OpenAPIInfo{
        Title:       "My API",
        Description: "API description",
        Version:     "1.0.0",
        Contact: &autofiber.OpenAPIContact{
            Name:  "API Team",
            Email: "team@example.com",
        },
    }).
    WithDocsServer(autofiber.OpenAPIServer{
        URL:         "http://localhost:3000",
        Description: "Development server",
    })

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

## Project Structure

```
auto-fiber/
├── app.go                # Core initialization and group methods
├── types.go              # Type definitions and structs
├── options.go            # Route options and configuration
├── handlers.go           # Handler creation and complete flow logic
├── routes.go             # HTTP method route handlers
├── docs_config.go        # Documentation configuration
├── docs.go               # OpenAPI specification generation
├── validator.go          # Validation logic
├── parser.go             # Parsing logic
├── map_parser.go         # Map parsing logic
├── middleware.go         # Middleware and response validation
├── group.go              # Route group functionality
└── example/
    └── main.go           # Complete usage example with full flow
```

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
