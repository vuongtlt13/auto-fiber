# Routing

AutoFiber wraps Fiber's router and adds automatic parsing, validation, and OpenAPI doc generation on top.

## Registering Routes

```go
app := autofiber.New()

app.Get("/users",       listUsers)
app.Post("/users",      createUser, autofiber.WithRequestSchema(CreateUserRequest{}))
app.Put("/users/:id",   updateUser, autofiber.WithRequestSchema(UpdateUserRequest{}))
app.Delete("/users/:id", deleteUser)
app.Patch("/users/:id", patchUser,  autofiber.WithRequestSchema(PatchUserRequest{}))
```

Every method accepts optional `RouteOption` variadic arguments for schema, tags, description, auth, and middleware.

## Route Options

| Option | Purpose |
|---|---|
| `WithRequestSchema(v)` | Schema used for parsing and validation |
| `WithResponseSchema(v)` | Schema used for response validation and OpenAPI |
| `WithJwtAuth()` | Require Bearer token; adds `bearerAuth` to OpenAPI |
| `WithTags(tags...)` | OpenAPI operation tags |
| `WithDescription(s)` | OpenAPI operation description |
| `WithMiddleware(h...)` | Fiber handlers prepended before the route handler |

## Route Groups

Groups share a URL prefix and can have group-level middleware or auth.

```go
api := app.Group("/api/v1")

api.Get("/health", healthHandler)
api.Post("/login",  loginHandler, autofiber.WithRequestSchema(LoginRequest{}))
```

### Group-Level Middleware

`WithMiddleware` runs the provided handlers before every route in the group.

```go
api := app.Group("/api/v1").WithMiddleware(loggingMiddleware, rateLimitMiddleware)

api.Get("/users", listUsers)    // loggingMiddleware + rateLimitMiddleware run first
api.Post("/users", createUser)  // same
```

`WithMiddleware` is chainable and additive — multiple calls append to the list.

### Group-Level JWT Auth

`WithJwtAuth` marks every route in the group as requiring an `Authorization` header.
A missing header returns `401 Missing Authorization header` before the handler runs.
All routes in the group also get `security: bearerAuth` in the OpenAPI spec.

```go
protected := app.Group("/admin").WithJwtAuth()

protected.Get("/dashboard",  dashboardHandler)
protected.Delete("/users/:id", deleteUserHandler)
```

### Combining Both

```go
protected := app.Group("/admin").
    WithJwtAuth().
    WithMiddleware(auditLogMiddleware)
```

Group-level middleware runs before any per-route middleware.

### Per-Route Override Inside a Group

You can still pass route-specific options that extend (not replace) group settings.

```go
admin := app.Group("/admin").WithJwtAuth()

// This route also enforces request schema validation on top of JWT.
admin.Post("/users", createUser,
    autofiber.WithRequestSchema(CreateUserRequest{}),
    autofiber.WithMiddleware(adminOnlyMiddleware),
)
```

## Handler Signatures

AutoFiber supports two signatures:

```go
// With request parsing (requires WithRequestSchema)
func handler(c *fiber.Ctx, req *RequestSchema) (interface{}, error)

// Without request parsing
func handler(c *fiber.Ctx) (interface{}, error)
```

Any other signature causes a panic at registration time.

## Accessing Raw Fiber App

The underlying `*fiber.App` is available as `app.App` for anything AutoFiber does not wrap directly (e.g. `app.App.Static(...)`).

## OpenAPI / Swagger

```go
app.ServeDocs("/docs")           // serves OpenAPI JSON
app.ServeSwaggerUI("/swagger", "/docs")  // serves Swagger UI
```

See [structs-and-tags.md](structs-and-tags.md) for how schemas map to OpenAPI components.
