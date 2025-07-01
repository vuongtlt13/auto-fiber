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
- **OpenAPI Schema Naming & Generic Response**:
  - **Schema Naming:** AutoFiber generates OpenAPI schema names that are RFC3986-compliant. For generic structs, the schema name will be in the form `APIResponse_User` (for `APIResponse[User]`). For non-generic structs, the schema name is simply the type name (e.g., `LoginResponse`).
  - **Generic Response Support:** You can use generic response wrappers for consistent API responses. Example:
    ```go
    type APIResponse[T any] struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
        Data    T      `json:"data"`
    }
    // Usage in route:
    app.Get("/user", handler.GetUser, autofiber.WithResponseSchema(APIResponse[User]{}))
    ```
  - **Request Body Rules:** Only POST, PUT, and PATCH methods generate a `requestBody` in the OpenAPI spec. GET, DELETE, HEAD, and OPTIONS never have a request body, even if a request schema is provided.

## Installation

```sh
go get github.com/vuongtlt13/auto-fiber
```

## Project Structure

```
auto-fiber/
  app.go            // App core: AutoFiber struct, route registration, group, listen, etc.
  group.go          // Route grouping logic
  handlers.go       // Handler creation, signature validation, error handling
  parser.go         // Request parsing from multiple sources (body, query, path, ...)
  validator.go      // Response validation logic
  map_parser.go     // Parse struct from map/interface
  docs.go           // OpenAPI/Swagger documentation generation
  options.go        // Route option functions (WithRequestSchema, WithResponseSchema, ...)
  types.go          // Core types, RouteOptions, ParseSource, etc.
  example/          // Example usage and demo app
  docs/             // Documentation and guides
```

- **app.go**: Initialize app, register routes, groups, listen.
- **group.go**: Support for route groups, group middleware.
- **handlers.go**: Create handlers with correct signature, signature validation.
- **parser.go**: Automatically parse requests from multiple sources (body, query, path, header, cookie).
- **validator.go**: Validate response before returning to client.
- **map_parser.go**: Support parsing struct from map/interface (for test, mock, ...).
- **docs.go**: Generate OpenAPI spec, serve Swagger UI/docs.
- **options.go**: Option functions for routes (schema, tags, description, ...).
- **types.go**: Define core types, RouteOptions, ParseSource, ...

## Quick Start

```go
package main

import (
    "time"
    "github.com/gofiber/fiber/v2"
    autofiber "github.com/vuongtlt13/auto-fiber"
)

// Request schema with parse tag
// (parse from path, query, header, body)
type CreateUserRequest struct {
    OrgID    int    `parse:"path:org_id" validate:"required"`
    Role     string `parse:"query:role" validate:"required,oneof=admin user"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required"`
}

type UserResponse struct {
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Name      string    `json:"name" validate:"required"`
    Role      string    `json:"role" validate:"required,oneof=admin user"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
}

type APIResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}

type UserHandler struct{}

// Handler signature for AutoFiber:
func (h *UserHandler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    user := UserResponse{
        ID:        1,
        Email:     req.Email,
        Name:      req.Name,
        Role:      req.Role,
        CreatedAt: time.Now(),
    }
    return user, nil
}

// Handler returning generic response
func (h *UserHandler) GetUser(c *fiber.Ctx) (interface{}, error) {
    user := UserResponse{
        ID:        1,
        Email:     "user@example.com",
        Name:      "John Doe",
        Role:      "user",
        CreatedAt: time.Now(),
    }
    return APIResponse[UserResponse]{Code: 0, Message: "success", Data: user}, nil
}

func main() {
    app := autofiber.NewWithOptions(
        fiber.Config{EnablePrintRoutes: true},
        autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
            Title:       "AutoFiber API",
            Description: "A sample API with complete request/response flow",
            Version:     "0.3.1",
        }),
    )

    handler := &UserHandler{}

    app.Post("/organizations/:org_id/users", handler.CreateUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
        autofiber.WithDescription("Create a new user in an organization"),
        autofiber.WithTags("users", "admin"),
    )

    app.Get("/user", handler.GetUser,
        autofiber.WithResponseSchema(APIResponse[UserResponse]{}),
        autofiber.WithDescription("Get a user with generic response"),
        autofiber.WithTags("users"),
    )

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

## Handler Signatures

**Recommended:**

```go
// Standard handler: return data and error, AutoFiber will marshal JSON automatically
func (h *Handler) CompleteHandler(c *fiber.Ctx, req *RequestSchema) (interface{}, error) {
    return ResponseSchema{...}, nil
}
```

**Use only for health check or custom response:**

```go
func (h *Handler) Health(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok"})
}
```

**NOT recommended:**

```go
// Do not call c.JSON manually if you already have a request schema
func (h *Handler) BadHandler(c *fiber.Ctx, req *RequestSchema) error {
    return c.JSON(...)
}
```

## Documentation

- [docs/README.md](docs/README.md) - Documentation index & guides
- [docs/structs-and-tags.md](docs/structs-and-tags.md) - Struct/tag/validation best practices
- [docs/complete-flow.md](docs/complete-flow.md) - Full request/response flow
- [docs/validation-rules.md](docs/validation-rules.md) - Validation rules & custom validators
- [example/](example/) - Example app

## Contributing

If you find any issues or want to improve the documentation:

1. Check the existing documentation first
2. Create an issue or pull request
3. Follow the same format and style as existing docs
4. Include practical examples and use cases

## License

MIT
