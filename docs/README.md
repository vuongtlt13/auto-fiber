# AutoFiber Documentation

Welcome to the AutoFiber documentation! This directory contains comprehensive guides and examples to help you build robust APIs with AutoFiber.

## 📚 Documentation Index

### Core Guides

- **[Structs and Tags Guide](structs-and-tags.md)** - Complete guide to creating request/response structs with parsing tags, validation tags, and best practices
- **[Complete Request/Response Flow](complete-flow.md)** - Detailed explanation of the complete flow: Parse Request → Validate Request → Execute Handler → Validate Response → Return JSON
- **[Validation Rules and Best Practices](validation-rules.md)** - Comprehensive guide to validation rules, custom validators, and validation patterns
- **[Migration Guide](migration-guide.md)** - Guide to migrate from old handler signatures to new required signatures

### Quick Navigation

| Topic               | Description                                   | File                                       |
| ------------------- | --------------------------------------------- | ------------------------------------------ |
| **Getting Started** | Basic setup and first API                     | [Main README](../README.md)                |
| **Structs & Tags**  | How to create request/response structs        | [structs-and-tags.md](structs-and-tags.md) |
| **Complete Flow**   | Understanding the full request/response cycle | [complete-flow.md](complete-flow.md)       |
| **Validation**      | Validation rules and custom validators        | [validation-rules.md](validation-rules.md) |
| **Migration**       | Migrate from old handler signatures           | [migration-guide.md](migration-guide.md)   |
| **Examples**        | Working examples and patterns                 | [../example/](../example/)                 |

## 🚀 Quick Start

If you're new to AutoFiber, start here:

1. **Read the [Main README](../README.md)** for installation and basic usage
2. **Check [structs-and-tags.md](structs-and-tags.md)** to understand how to create your request/response structs
3. **Review [complete-flow.md](complete-flow.md)** to understand the full request/response flow
4. **Explore [validation-rules.md](validation-rules.md)** for advanced validation techniques

## 📖 What You'll Learn

### From Structs and Tags Guide

- How to create request and response structs
- Understanding parse tags for different data sources (path, query, header, body, etc.)
- Using validation tags effectively
- JSON tag best practices
- Special cases and edge scenarios
- Common patterns for CRUD operations, authentication, and search
- **How OpenAPI schema names are generated for generic and non-generic types**
- **How to use generic response wrappers and reference them in OpenAPI**
- **Request body generation rules for each HTTP method**

### From Complete Flow Guide

- The complete request/response flow in AutoFiber
- **Required handler signature:**
  ```go
  func (h *Handler) MyEndpoint(c *fiber.Ctx, req *RequestSchema) (interface{}, error) {
      // Business logic
      return ResponseSchema{...}, nil
  }
  ```
- Response validation and error handling
- Real-world examples with authentication and user management
- Best practices for building robust APIs

### From Validation Rules Guide

- Built-in validation rules for strings, numbers, arrays, and structs
- Creating custom validators for business logic
- Conditional validation and cross-field validation
- Common validation patterns for user registration, product management, and pagination
- Error handling and custom error messages

## 🎯 Common Use Cases

### User Management API

```go
// Request schema with multi-source parsing
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

// Handler signature CHUẨN cho AutoFiber:
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

// Handler không có request parsing:
func (h *Handler) GetUser(c *fiber.Ctx) (interface{}, error) {
    user := UserResponse{
        ID:        1,
        Email:     "user@example.com",
        Name:      "John Doe",
        Role:      "user",
        CreatedAt: time.Now(),
    }
    return user, nil
}
```

### Authentication Flow

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
    Token     string       `json:"token" validate:"required"`
    User      UserResponse `json:"user" validate:"required"`
    ExpiresAt time.Time    `json:"expires_at" validate:"required"`
}
```

### Search and Pagination

```go
type SearchRequest struct {
    Query      string   `parse:"query:q" validate:"omitempty,min=2"`
    Categories []string `parse:"query:categories" validate:"omitempty,dive,oneof=tech sports news"`
    Page       int      `parse:"query:page" validate:"omitempty,gte=1"`
    Limit      int      `parse:"query:limit" validate:"omitempty,gte=1,lte=100"`
}
```

## 🔧 Advanced Topics

### Custom Validators

```go
// Register custom validator
validator := autofiber.GetValidator()
validator.RegisterValidation("strong_password", validateStrongPassword)

// Use in struct
type UserRequest struct {
    Password string `validate:"required,strong_password"`
}
```

### Response Validation

```go
app.Post("/users", handler.CreateUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
    autofiber.WithResponseSchema(UserResponse{}), // Enables response validation
)
```

### Multi-Source Parsing

```go
type ComplexRequest struct {
    UserID int    `parse:"path:user_id" validate:"required"`
    Token  string `parse:"header:Authorization" validate:"required"`
    Email  string `json:"email" validate:"required,email"`
    Name   string `json:"name" validate:"required"`
}
```

### OpenAPI Schema Naming for Generics

- Generic structs (e.g., `APIResponse[User]`) are named as `APIResponse_User` in OpenAPI. Non-generic structs use their type name.

### Generic Response Support

- Use generic wrappers for consistent API responses and OpenAPI documentation.

### Request Body Rules

- Only POST, PUT, PATCH methods generate a request body in OpenAPI. GET, DELETE, HEAD, OPTIONS do not, even if a request schema is provided.

## ⚠️ Important Notes

### Handler Signatures

AutoFiber requires specific handler signatures:

```go
// ✅ Required: With request parsing
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    return ResponseSchema{...}, nil
}

// ✅ Required: Without request parsing
func (h *Handler) GetUser(c *fiber.Ctx) (interface{}, error) {
    return ResponseSchema{...}, nil
}

// ❌ NOT supported (will cause panic)
func (h *Handler) BadHandler(c *fiber.Ctx, req *RequestSchema) error {
    return c.JSON(...)
}
```

> **Important:** AutoFiber requires handlers to return `(interface{}, error)` for automatic JSON marshaling and response validation. The old signature `func(c *fiber.Ctx, req *T) error` is no longer supported.

## 📝 Contributing

If you find any issues or want to improve the documentation:

1. Check the existing documentation first
2. Create an issue or pull request
3. Follow the same format and style as existing docs
4. Include practical examples and use cases

## 🆘 Need Help?

- **GitHub Issues**: [Create an issue](https://github.com/vuongtlt13/auto-fiber/issues)
- **Examples**: Check the [example directory](../example/) for working code
- **Main README**: [Back to main documentation](../README.md)

---

Happy coding with AutoFiber! 🚀
