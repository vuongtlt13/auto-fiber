# AutoFiber - FastAPI-like Wrapper for Fiber

AutoFiber is a wrapper around Fiber that provides automatic request/response parsing and validation, similar to FastAPI's automatic request handling. It also includes automatic OpenAPI/Swagger documentation generation.

## 🚀 Features

- ✅ **Auto Request Parsing**: Automatically parse JSON request bodies into Go structs
- ✅ **Auto Validation**: Validate requests using struct tags
- ✅ **Type Safety**: Full type safety with Go generics
- ✅ **Route Options**: Declarative route configuration with options
- ✅ **Group Support**: Auto-parse support in route groups
- ✅ **Middleware Integration**: Seamless middleware integration
- ✅ **Auto API Documentation**: Generate OpenAPI/Swagger docs automatically
- ✅ **Swagger UI**: Built-in Swagger UI for interactive API documentation
- ✅ **Clean Architecture**: Well-organized, maintainable code structure

## 📁 Project Structure

```
autofiber/
├── autofiber.go      # Core functionality and initialization
├── types.go          # Type definitions and structs
├── options.go        # Route option functions
├── handlers.go       # Handler creation and processing logic
├── routes.go         # HTTP method route handlers
├── docs_config.go    # Documentation configuration methods
├── docs.go           # OpenAPI specification generation
├── group.go          # Route group functionality
├── go.mod            # Go module dependencies
├── README.md         # Project documentation
└── example/
    └── main.go       # Example application
```

## 🛠️ Installation

```bash
go get github.com/yourusername/autofiber
```

## ⚡ Quick Start

### 1. Define your request schemas

```go
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
```

### 2. Create your handlers

```go
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

// Simple handler without request parsing
func (h *AuthHandler) Health(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "ok", "timestamp": time.Now()})
}
```

### 3. Register routes with AutoFiber

```go
func main() {
    // Create AutoFiber app with docs configuration
    app := autofiber.New().
        WithDocsInfo(autofiber.OpenAPIInfo{
            Title:       "My API",
            Description: "A sample API demonstrating AutoFiber capabilities",
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
```

## 📚 Auto Documentation

AutoFiber automatically generates OpenAPI 3.0 documentation from your route definitions and struct schemas.

### Features

- **Automatic Schema Generation**: Converts Go structs to OpenAPI schemas
- **Request/Response Documentation**: Documents request bodies and responses
- **Validation Rules**: Includes validation rules in documentation
- **Path Parameters**: Automatically detects and documents path parameters
- **Tags and Descriptions**: Organize endpoints with tags and descriptions
- **Swagger UI**: Interactive API documentation interface

### Struct Tags for Documentation

```go
type UserRequest struct {
    Email     string    `json:"email" validate:"required,email" description:"User email address" example:"user@example.com"`
    Password  string    `json:"password" validate:"required,min=6" description:"User password"`
    Age       int       `json:"age" validate:"gte=0,lte=130" description:"User age"`
    Website   string    `json:"website" validate:"url" description:"User website"`
    BirthDate time.Time `json:"birth_date" description:"User birth date"`
}
```

### Accessing Documentation

- **OpenAPI JSON**: `GET /docs` - Raw OpenAPI specification
- **Swagger UI**: `GET /swagger` - Interactive documentation interface

## 🔧 Route Options

### `WithRequestSchema(schema interface{})`

Sets the request schema for auto-parsing and validation.

```go
app.Post("/users", handler.CreateUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
)
```

### `WithResponseSchema(schema interface{})`

Sets the response schema for documentation purposes.

```go
app.Get("/users/:id", handler.GetUser,
    autofiber.WithRequestSchema(GetUserRequest{}),
    autofiber.WithResponseSchema(UserResponse{}),
)
```

### `WithMiddleware(middleware ...fiber.Handler)`

Adds middleware to the route.

```go
app.Get("/protected", handler.Protected,
    autofiber.WithMiddleware(authMiddleware, rateLimitMiddleware),
)
```

### `WithDescription(description string)`

Sets the route description for documentation.

```go
app.Post("/login", handler.Login,
    autofiber.WithDescription("Authenticate user and return JWT token"),
)
```

### `WithTags(tags ...string)`

Sets the route tags for documentation.

```go
app.Post("/users", handler.CreateUser,
    autofiber.WithTags("users", "admin"),
)
```

## 📝 Handler Signatures

AutoFiber supports different handler signatures:

### 1. Simple Handler

```go
func (h *Handler) Simple(c *fiber.Ctx) error
```

### 2. Handler with Request Parsing

```go
func (h *Handler) WithRequest(c *fiber.Ctx, req *RequestSchema) error
```

### 3. Handler with Request Parsing and Response Formatting

```go
func (h *Handler) WithResponse(c *fiber.Ctx, req *RequestSchema) (interface{}, error)
```

## 🏗️ Route Groups

AutoFiber supports route groups with auto-parse capabilities:

```go
// Create a group
api := app.Group("/api/v1")

// Add routes to the group
api.Post("/login", handler.Login,
    autofiber.WithRequestSchema(LoginRequest{}),
)

api.Get("/users", handler.ListUsers,
    autofiber.WithRequestSchema(UserFilterRequest{}),
)

// Groups can also have middleware
admin := app.Group("/admin", authMiddleware)
admin.Get("/dashboard", handler.Dashboard)
```

## 🌐 HTTP Methods

AutoFiber supports all HTTP methods with options:

```go
app.Get("/users", handler.ListUsers, options...)
app.Post("/users", handler.CreateUser, options...)
app.Put("/users/:id", handler.UpdateUser, options...)
app.Delete("/users/:id", handler.DeleteUser, options...)
app.Patch("/users/:id", handler.PartialUpdate, options...)
app.Head("/health", handler.Health, options...)
app.Options("/users", handler.Options, options...)
app.All("/catch-all", handler.CatchAll, options...)
```

## ✅ Validation

AutoFiber uses `go-playground/validator` for validation. Common tags:

```go
type UserRequest struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=6,max=50"`
    Age       int    `json:"age" validate:"gte=0,lte=130"`
    Website   string `json:"website" validate:"url"`
    Username  string `json:"username" validate:"required,alphanum"`
}
```

## ❌ Error Handling

When validation fails, AutoFiber returns structured error responses:

```json
{
  "error": "Validation failed",
  "details": {
    "email": "email must be a valid email",
    "password": "password must be at least 6 characters"
  }
}
```

## 🔄 Comparison with Traditional Fiber

### Before (Traditional Fiber)

```go
func (h *AuthHandler) Login(c *fiber.Ctx) error {
    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }

    if err := validate.Struct(req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Validation failed"})
    }

    return c.JSON(fiber.Map{"email": req.Email})
}

app.Post("/login", handler.Login)
```

### After (With AutoFiber)

```go
func (h *AuthHandler) Login(c *fiber.Ctx, req *LoginRequest) error {
    // req is automatically parsed and validated
    return c.JSON(fiber.Map{"email": req.Email})
}

app.Post("/login", handler.Login,
    autofiber.WithRequestSchema(LoginRequest{}),
    autofiber.WithDescription("Authenticate user and return JWT token"),
    autofiber.WithTags("auth"),
)
```

## 🎯 Benefits

1. **Less Boilerplate**: No need to manually parse and validate requests
2. **Type Safety**: Full type safety with Go generics
3. **Declarative**: Route configuration is declarative and self-documenting
4. **FastAPI-like**: Similar developer experience to FastAPI
5. **Maintainable**: Cleaner, more readable handler code
6. **Auto Documentation**: Built-in OpenAPI/Swagger documentation generation
7. **Interactive Docs**: Swagger UI for testing APIs directly
8. **Clean Architecture**: Well-organized, maintainable code structure

## 🚀 Migration Guide

To migrate from traditional Fiber to AutoFiber:

1. **Replace Fiber.New()** with `autofiber.New()`
2. **Update handler signatures** to accept typed request parameters
3. **Add route options** for request/response schemas
4. **Remove manual validation** calls from handlers
5. **Add documentation options** for auto-generated API docs

## 🧪 Running the Example

```bash
# Clone the repository
git clone https://github.com/yourusername/autofiber.git
cd autofiber

# Install dependencies
go mod tidy

# Run the example
go run example/main.go
```

Then visit:

- **API**: http://localhost:3000
- **OpenAPI JSON**: http://localhost:3000/docs
- **Swagger UI**: http://localhost:3000/swagger

## 📦 Dependencies

- `github.com/gofiber/fiber/v2` - Web framework
- `github.com/go-playground/validator/v10` - Validation
- `github.com/stretchr/testify` - Testing (optional)

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Fast web framework for Go
- [FastAPI](https://fastapi.tiangolo.com/) - Inspiration for the developer experience
- [OpenAPI](https://swagger.io/specification/) - API documentation standard

---

AutoFiber provides a much cleaner and more maintainable approach to building HTTP APIs in Go, similar to the experience with FastAPI in Python, with the added benefit of automatic API documentation generation and a clean, modular architecture.
