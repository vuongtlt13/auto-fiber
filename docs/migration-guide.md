# Migration Guide: Handler Signatures Update

This guide helps you migrate from the old handler signatures to the new required signatures in AutoFiber.

## What Changed?

AutoFiber now requires handlers to return `(interface{}, error)` instead of just `error`. This change enables:

- Automatic JSON marshaling
- Response validation
- Better error handling
- Consistent API responses

## Migration Steps

### 1. Update Handler Signatures

**Before (Old - Will Cause Panic):**

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) error {
    user := UserResponse{
        ID:   1,
        Name: req.Name,
    }
    return c.JSON(user)
}
```

**After (New - Required):**

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    user := UserResponse{
        ID:   1,
        Name: req.Name,
    }
    return user, nil
}
```

### 2. Update Handlers Without Request Parsing

**Before:**

```go
func (h *Handler) GetUser(c *fiber.Ctx) error {
    user := UserResponse{
        ID:   1,
        Name: "John Doe",
    }
    return c.JSON(user)
}
```

**After:**

```go
func (h *Handler) GetUser(c *fiber.Ctx) (interface{}, error) {
    user := UserResponse{
        ID:   1,
        Name: "John Doe",
    }
    return user, nil
}
```

### 3. Keep Custom Response Handlers (Optional)

For health checks or custom responses, you can still use the old signature:

```go
func (h *Handler) Health(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status": "ok",
        "timestamp": time.Now(),
    })
}
```

## Common Migration Patterns

### Pattern 1: Simple Return

**Before:**

```go
func (h *Handler) GetUsers(c *fiber.Ctx) error {
    users := []UserResponse{...}
    return c.JSON(users)
}
```

**After:**

```go
func (h *Handler) GetUsers(c *fiber.Ctx) (interface{}, error) {
    users := []UserResponse{...}
    return users, nil
}
```

### Pattern 2: With Error Handling

**Before:**

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) error {
    if req.Email == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Email required"})
    }

    user := UserResponse{...}
    return c.JSON(user)
}
```

**After:**

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    if req.Email == "" {
        return nil, errors.New("Email required")
    }

    user := UserResponse{...}
    return user, nil
}
```

### Pattern 3: With Database Operations

**Before:**

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) error {
    user, err := db.CreateUser(req)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(user)
}
```

**After:**

```go
func (h *Handler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    user, err := db.CreateUser(req)
    if err != nil {
        return nil, err
    }

    return user, nil
}
```

## Error Handling

### Automatic Error Responses

AutoFiber automatically handles errors and returns appropriate HTTP status codes:

```go
func (h *Handler) GetUser(c *fiber.Ctx, req *GetUserRequest) (interface{}, error) {
    user, err := db.GetUser(req.ID)
    if err != nil {
        return nil, err // AutoFiber will return 500 Internal Server Error
    }

    if user == nil {
        return nil, errors.New("user not found") // AutoFiber will return 404 Not Found
    }

    return user, nil
}
```

### Custom Error Types

You can create custom error types for specific status codes:

```go
type NotFoundError struct {
    message string
}

func (e NotFoundError) Error() string {
    return e.message
}

func (h *Handler) GetUser(c *fiber.Ctx, req *GetUserRequest) (interface{}, error) {
    user, err := db.GetUser(req.ID)
    if err != nil {
        return nil, err
    }

    if user == nil {
        return nil, NotFoundError{"user not found"}
    }

    return user, nil
}
```

## Testing Your Migration

### 1. Run Your Tests

```bash
go test ./...
```

### 2. Check for Panic Errors

Look for errors like:

```
panic: handler signature not supported
```

### 3. Update Test Handlers

If you have test handlers, update them too:

**Before:**

```go
func TestHandler(t *testing.T) {
    handler := &Handler{}
    req := &CreateUserRequest{...}

    // This will panic now
    err := handler.CreateUser(nil, req)
}
```

**After:**

```go
func TestHandler(t *testing.T) {
    handler := &Handler{}
    req := &CreateUserRequest{...}

    result, err := handler.CreateUser(nil, req)
    // Handle result and error
}
```

## Benefits of the New Signature

1. **Automatic JSON Marshaling**: No need to call `c.JSON()` manually
2. **Response Validation**: Automatic validation of response data
3. **Consistent Error Handling**: Standardized error responses
4. **Better Testing**: Easier to test return values
5. **Type Safety**: Better type checking at compile time

## Need Help?

If you encounter issues during migration:

1. Check the [Complete Flow Guide](complete-flow.md) for detailed examples
2. Look at the [example directory](../example/) for working code
3. Create an issue on GitHub with your specific error

## Summary

The migration is straightforward:

1. Change `error` return to `(interface{}, error)`
2. Remove `c.JSON()` calls
3. Return data and error directly
4. Update tests accordingly

This change makes AutoFiber more robust and easier to use while maintaining backward compatibility for custom response handlers.
