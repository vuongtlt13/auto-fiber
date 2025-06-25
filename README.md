# AutoFiber

A FastAPI-like wrapper for the Fiber web framework in Go, providing automatic request parsing, validation, and OpenAPI/Swagger documentation generation.

## Features

- **🔄 Auto Request Parsing**: Automatically parse requests from multiple sources (body, query, path, headers, cookies, form)
- **🧠 Smart Parsing**: Auto-detect the best source based on HTTP method (GET: path→query→body, POST: body→path→query)
- **✅ Request Validation**: Built-in validation using struct tags with `go-playground/validator`
- **✅ Response Validation**: Validate response data before sending to client
- **📚 Auto Documentation**: Generate OpenAPI 3.0 specification and Swagger UI
- **🔒 Type Safety**: Full type safety with Go generics
- **⚙️ Route Options**: Flexible route configuration with options pattern
- **🔌 Middleware Integration**: Seamless integration with Fiber middleware
- **🎯 Clean Architecture**: Modular design with separate concerns

## Installation

```bash
go get github.com/vuongtlt13/auto-fiber
```

## Quick Start

```go
package main

import (
    "time"
    "autofiber"
    "github.com/gofiber/fiber/v2"
)

// Request schema with parse tag
type CreateUserRequest struct {
    // Path parameter
    OrgID int `parse:"path:org_id" validate:"required" description:"Organization ID"`

    // Query parameters
    Role     string `parse:"query:role" validate:"required,oneof=admin user" description:"User role"`
    IsActive bool   `parse:"query:active" description:"User active status"`

    // Headers
    APIKey string `parse:"header:X-API-Key" validate:"required" description:"API key"`

    // Body fields
    Email    string `parse:"body:email" validate:"required,email" description:"User email"`
    Password string `parse:"body:password" validate:"required,min=6" description:"User password"`
    Name     string `parse:"body:name" validate:"required" description:"User full name"`
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

func (h *UserHandler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
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

func main() {
    app := autofiber.New().
        WithDocsInfo(autofiber.OpenAPIInfo{
            Title:       "AutoFiber API",
            Description: "A sample API with parse tag and validation",
            Version:     "1.0.0",
        })

    handler := &UserHandler{}

    // Register route with auto-parsing and response validation
    app.Post("/organizations/:org_id/users", handler.CreateUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
        autofiber.WithDescription("Create a new user in an organization"),
        autofiber.WithTags("users", "admin"),
    )

    // Serve documentation
    app.ServeDocs("/docs")
    app.ServeSwaggerUI("/swagger", "/docs")

    app.Listen(":3000")
}
```

## Parse Tag

AutoFiber uses the `parse` tag to specify where each field should be parsed from:

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

### Smart Parsing with Auto

Use `parse:"auto:key"` for automatic source detection:

```go
type GetUserRequest struct {
    // Auto-detect based on HTTP method
    UserID int `parse:"auto:user_id" validate:"required"`
    Page   int `parse:"auto:page" validate:"gte=1"`
    Active bool `parse:"auto:active"`
}
```

**HTTP Method Priority:**

- **GET**: `path` → `query` → `body`
- **POST/PUT/PATCH**: `body` → `path` → `query`
- **DELETE**: `path` → `query`

### Parse Tag Options

```go
type Request struct {
    // Required field
    UserID int `parse:"path:user_id,required"`

    // With default value
    Page   int  `parse:"query:page,default:1"`
    Limit  int  `parse:"query:limit,default:10"`
    Active bool `parse:"query:active,default:true"`

    // Multiple options
    Role string `parse:"query:role,required,default:user"`
}
```

## Smart Parsing

AutoFiber automatically detects the best source for parsing fields based on HTTP method:

### HTTP Method Priority

- **GET**: `path` → `query` → `body`
- **POST/PUT/PATCH**: `body` → `path` → `query`
- **DELETE**: `path` → `query`

### Field Source Tags

```go
type Request struct {
    // Using parse tag (recommended)
    UserID int `parse:"path:user_id" validate:"required" description:"User ID from path"`
    Page   int `parse:"query:page" validate:"gte=1" description:"Page number"`
    Token  string `parse:"header:Authorization" validate:"required" description:"Bearer token"`
    Email  string `parse:"body:email" validate:"required,email" description:"User email"`

    // Smart parsing with auto
    ID     int  `parse:"auto:id" validate:"required" description:"Auto-detected source"`
    Active bool `parse:"auto:active" description:"Auto-detected source"`
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

## Request Validation

Use struct tags for validation:

```go
type UserRequest struct {
    Email     string `parse:"body:email" validate:"required,email" description:"User email"`
    Password  string `parse:"body:password" validate:"required,min=6,max=50" description:"User password"`
    Age       int    `parse:"body:age" validate:"gte=18,lte=100" description:"User age"`
    Role      string `parse:"body:role" validate:"required,oneof=admin user guest" description:"User role"`
    IsActive  bool   `parse:"body:is_active" description:"User active status"`
}
```

## Response Validation

AutoFiber can validate response data before sending to clients:

```go
// Response schema with validation
type UserResponse struct {
    ID        int       `json:"id" validate:"required" description:"User ID"`
    Email     string    `json:"email" validate:"required,email" description:"User email"`
    Name      string    `json:"name" validate:"required" description:"User name"`
    Role      string    `json:"role" validate:"required,oneof=admin user" description:"User role"`
    CreatedAt time.Time `json:"created_at" validate:"required" description:"Account creation date"`
}

// Register route with response validation
app.Post("/users", handler.CreateUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
    autofiber.WithResponseSchema(UserResponse{}), // This enables response validation
    autofiber.WithDescription("Create a new user"),
)
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

AutoFiber automatically generates OpenAPI 3.0 specification and serves Swagger UI based on your parse tags:

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

AutoFiber automatically generates proper OpenAPI documentation from your parse tags:

```go
type CreateUserRequest struct {
    // Path parameter - appears in "Parameters" section
    OrgID int `parse:"path:org_id" validate:"required" description:"Organization ID"`

    // Query parameters - appear in "Parameters" section
    Role     string `parse:"query:role" validate:"required,oneof=admin user" description:"User role"`
    IsActive bool   `parse:"query:active" description:"User active status"`

    // Header parameters - appear in "Parameters" section
    APIKey string `parse:"header:X-API-Key" validate:"required" description:"API key"`

    // Body fields - appear in "Request Body" section
    Email    string `parse:"body:email" validate:"required,email" description:"User email"`
    Password string `parse:"body:password" validate:"required,min=6" description:"User password"`
    Name     string `parse:"body:name" validate:"required" description:"User full name"`
}
```

**Generated OpenAPI Documentation:**

- **Parameters**: `org_id` (path), `role` (query), `active` (query), `X-API-Key` (header)
- **Request Body**: `email`, `password`, `name` (JSON schema)
- **Validation**: Required fields, email format, min length, enum values
- **Descriptions**: Field descriptions from struct tags

### Documentation Features

- **Automatic Parameter Detection**: Path, query, header, cookie parameters
- **Request Body Schema**: Only body fields appear in request body
- **Validation Rules**: Required fields, format validation, enum values
- **Field Descriptions**: From `description` struct tags
- **Examples**: From `example` struct tags
- **Type Safety**: Proper OpenAPI types (string, integer, boolean, etc.)
- **Interactive Testing**: Swagger UI for testing APIs directly

## Project Structure

```
auto-fiber/
├── autofiber.go          # Core initialization and group methods
├── types.go              # Type definitions and structs
├── options.go            # Route options and configuration
├── handlers.go           # Handler creation and middleware logic
├── routes.go             # HTTP method route handlers
├── docs_config.go        # Documentation configuration
├── docs.go               # OpenAPI specification generation
├── pkg/
│   └── middleware/
│       └── middleware.go  # Request/response parsing and validation
└── example/
    └── main.go           # Complete usage example
```

## Examples

See the `example/` directory for complete working examples:

- **Basic Usage**: Simple request/response handling
- **Parse Tag**: Using parse tag for field source specification
- **Smart Parsing**: Auto-detection of field sources with `auto`
- **Response Validation**: Validating response data
- **API Documentation**: OpenAPI and Swagger UI

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
