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

- **app.go**: Khởi tạo app, đăng ký route, group, listen.
- **group.go**: Hỗ trợ group route, middleware cho group.
- **handlers.go**: Tạo handler đúng signature, kiểm tra lỗi signature.
- **parser.go**: Tự động parse request từ nhiều nguồn (body, query, path, header, cookie).
- **validator.go**: Validate response trước khi trả về client.
- **map_parser.go**: Hỗ trợ parse struct từ map/interface (phục vụ test, mock, ...).
- **docs.go**: Sinh OpenAPI spec, serve Swagger UI/docs.
- **options.go**: Các hàm option cho route (schema, tags, description, ...).
- **types.go**: Định nghĩa các type core, RouteOptions, ParseSource, ...

## Quick Start

```go
package main

import (
    "time"
    "github.com/vuongtlt13/auto-fiber"
    "github.com/gofiber/fiber/v2"
)

// Request schema with parse tag
// (parse từ path, query, header, body)
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

func main() {
    app := autofiber.NewWithOptions(
        fiber.Config{EnablePrintRoutes: true},
        autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
            Title:       "AutoFiber API",
            Description: "A sample API with complete request/response flow",
            Version:     "1.0.0",
        }),
    )

    handler := &UserHandler{}

    app.Post("/organizations/:org_id/users", handler.CreateUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
        autofiber.WithDescription("Create a new user in an organization"),
        autofiber.WithTags("users", "admin"),
    )

    app.ServeDocs("/docs")
    app.ServeSwaggerUI("/swagger", "/docs")
    app.Listen(":3000")
}
```

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

## Documentation

- [docs/README.md](docs/README.md) - Documentation index & guides
- [docs/structs-and-tags.md](docs/structs-and-tags.md) - Struct/tag/validation best practices
- [docs/complete-flow.md](docs/complete-flow.md) - Full request/response flow
- [docs/validation-rules.md](docs/validation-rules.md) - Validation rules & custom validators
- [example/](example/) - Example app

## License

MIT
