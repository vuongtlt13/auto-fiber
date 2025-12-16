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
  handlers.go       // Handler creation, signature validation, Authorization checks, response validation
  parser.go         // Request parsing from multiple sources (body, query, path, ...)
  validator.go      // Response validation logic
  map_parser.go     // Parse struct from map/interface
  docs.go           // OpenAPI/Swagger documentation generation (bearerAuth, security)
  options.go        // Route option functions (WithRequestSchema, WithResponseSchema, WithJwtAuth, ...)
  types.go          // Core types, RouteOptions, ParseSource, RequireJWTAuth inference
  example/          // Example usage and demo app
  docs/             // Documentation and guides
```

- **app.go**: Initialize app, register routes, groups, listen.
- **group.go**: Support for route groups, group middleware.
- **handlers.go**: Create handlers with correct signature, signature validation, Authorization enforcement (401 on missing header when JWT is required), response validation.
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

## Using Struct Validation Like Pydantic

AutoFiber uses [`go-playground/validator`](https://pkg.go.dev/github.com/go-playground/validator/v10) under the hood.  
You can declare structs with `validate:"..."` tags and manually trigger validation, similar to Pydantic,
either via the `ValidateStruct` helper or directly from `GetValidator()`:

```go
type UserInput struct {
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=18"`
}

func main() {
    input := &UserInput{
        Email: "invalid-email",
        Age:   16,
    }

    // Option 1: use AutoFiber's generic helper
    if err := autofiber.ValidateStruct(input); err != nil {
        fmt.Printf("validation error (ValidateStruct): %+v\n", err)
    }

    // Option 2: use the global validator directly (if you need more control)
    v := autofiber.GetValidator()
    if err := v.Struct(input); err != nil {
        fmt.Printf("validation error (GetValidator): %+v\n", err)
    }
}
```

In a request handler, if you use:

```go
autofiber.WithRequestSchema(MyRequest{})
```

AutoFiber will:

1. Parse data into `*MyRequest` (body, query, path, header, cookie, form) based on tags.
2. Call `ValidateStruct(req)` (or `GetValidator().Struct(req)`) to validate.
3. Only execute your handler when the data is valid.

### Notes about `required`, zero values, and nullable fields

AutoFiber follows `go-playground/validator`'s semantics for `required`:

- For **value types** (`int`, `float`, `string`, `bool`, structs):
  - `required` means the field must be **non-zero** for its Go type:
    - `0` for integers is considered **invalid** for `required`
    - `0.0` for floats is invalid
    - `""` (empty string) is invalid
    - `false` for bool is invalid
- For **pointer and reference types** (`*T`, `[]T`, `map[...]T`, etc.):
  - `required` means the value must be **non-nil**.

Common patterns:

```go
type User struct {
    // Must be present and non-empty
    Name string `json:"name" validate:"required"`

    // 0 is allowed, but you still want a lower bound
    Age int `json:"age" validate:"gte=0"`

    // "Required but nullable" for JSON:
    // - JSON must contain "nickname"
    // - value can be null or a string
    Nickname *string `json:"nickname" validate:"required"`
}
```

With the above:

- `{"name": "A", "age": 0, "nickname": "B"}` → valid
- `{"name": "A", "age": 0, "nickname": null}` → valid (field present but null)
- `{"name": "A", "age": 0}` → invalid (missing required `nickname`)
- `{"age": 0, "nickname": "B"}` → invalid (missing required `name`)

If you want `0` to be a valid value **and** enforce presence, prefer pointer types plus `required` or value types with range checks (e.g. `gte=0`) instead of `required` alone.

## Handler Signatures

**Supported Signatures for AutoFiber:**

```go
// Standard handler with request parsing: return data and error
// You can use interface{} or the concrete response schema type
func (h *Handler) CompleteHandler(c *fiber.Ctx, req *RequestSchema) (interface{}, error) {
    return ResponseSchema{...}, nil
}

// When using WithResponseSchema, prefer returning the concrete schema type for better type safety
func (h *Handler) CompleteHandlerTyped(c *fiber.Ctx, req *RequestSchema) (*ResponseSchema, error) {
    return &ResponseSchema{...}, nil
}

// Handler without request parsing: return data and error
func (h *Handler) SimpleHandler(c *fiber.Ctx) (interface{}, error) {
    return ResponseSchema{...}, nil
}

// When using WithResponseSchema, prefer returning the concrete schema type
func (h *Handler) SimpleHandlerTyped(c *fiber.Ctx) (*ResponseSchema, error) {
    return &ResponseSchema{...}, nil
}
```

**Use only for health check or custom response:**

```go
func (h *Handler) Health(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok"})
}
```

**NOT supported (will cause panic):**

```go
// Do not use this signature - AutoFiber requires (interface{}, error) or (*Schema, error) return
func (h *Handler) BadHandler(c *fiber.Ctx, req *RequestSchema) error {
    return c.JSON(...)
}
```

> **Note:** 
> - AutoFiber supports both `(interface{}, error)` and `(*ResponseSchema, error)` return types.
> - **When using `WithResponseSchema`, prefer returning the concrete schema type** (e.g., `*UserResponse`) instead of `interface{}` for better type safety and clarity.
> - The old signature `func(c *fiber.Ctx, req *T) error` is no longer supported.

## JWT Auth: Two Ways to Declare and Enforce Authorization

AutoFiber supports Bearer (JWT) in OpenAPI/Swagger and runtime enforcement. You can declare JWT in two ways:

1) **Route option:** `WithJwtAuth()`
   - Adds `bearerAuth` security to the operation (OpenAPI) and ensures runtime checks for `Authorization`.

2) **Request schema:** required `Authorization` header
   - Example field: ``Authorization string `parse:"header:Authorization" validate:"required"``  
   - `applyOptions` will auto-set `RequireJWTAuth` when it detects a required Authorization header in your schema (including embedded structs).

### Runtime behavior
- If `RequireJWTAuth` is true (either via `WithJwtAuth` or auto-inferred from schema), and the `Authorization` header is missing, the request returns **401 Missing Authorization header**.
- This check happens for both handlers with and without a request schema.

### Example: route option (no header parsing in schema)
```go
app.Get("/profile",
    handler.Profile,
    autofiber.WithJwtAuth(), // declares Bearer auth in docs + runtime 401 on missing Authorization
)
```

### Example: request schema (explicit header parsing)
```go
type ProfileRequest struct {
    Authorization string `parse:"header:Authorization" validate:"required" description:"Bearer <token>"`
}

app.Get("/profile",
    handler.ProfileWithHeaderParse,
    autofiber.WithRequestSchema(ProfileRequest{}), // auto-infers RequireJWTAuth from schema
)
```

### What Swagger UI shows
- Any route with JWT (either method) gets `security: [{"bearerAuth": []}]` and the `bearerAuth` scheme is added to `components.securitySchemes`.
- Users can click **Authorize** and enter a Bearer token once; it applies to all secured routes.

## HTTP Methods and Request Bodies (DELETE behavior)

AutoFiber aligns request body handling with common HTTP API practices:

- **GET, DELETE, HEAD, OPTIONS**:
  - By default, **no request body** is generated in OpenAPI (no `requestBody`), even if your request schema is a struct.
  - Fields without `parse` tags are treated as **path/query parameters only**, not body.
  - If you want a body for these methods (e.g., a bulk DELETE), you **must explicitly use** `parse:"body:..."` on the fields you want in the body.

- **POST, PUT, PATCH**:
  - If the request schema is a struct and you don't specify `parse:"body:..."`, AutoFiber will:
    - Treat struct fields as coming from the body by default (unless a `parse` tag says otherwise).
    - Generate a `requestBody` in OpenAPI pointing to the struct schema.

This means:
- `DELETE /resource/:id` is typically modeled with **path + query** only (no body).
- Advanced patterns like `DELETE /resources` with a JSON body for bulk operations are supported, but require **explicit** `parse:"body:..."` tags on the relevant fields.

## Documentation

- [docs/README.md](docs/README.md) - Documentation index & guides
- [docs/structs-and-tags.md](docs/structs-and-tags.md) - Struct/tag/validation best practices
- [docs/complete-flow.md](docs/complete-flow.md) - Full request/response flow
- [docs/validation-rules.md](docs/validation-rules.md) - Validation rules & custom validators
- [docs/migration-guide.md](docs/migration-guide.md) - Migrate from old handler signatures
- [example/](example/) - Example app

## Contributing

If you find any issues or want to improve the documentation:

1. Check the existing documentation first
2. Create an issue or pull request
3. Follow the same format and style as existing docs
4. Include practical examples and use cases

## License

MIT
