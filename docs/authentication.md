# Authentication

AutoFiber has built-in support for Bearer (JWT) authentication that wires into both the runtime enforcement and the OpenAPI spec.

## Two Ways to Declare JWT Auth

### 1. Route Option: `WithJwtAuth()`

Use this when you want auth enforced but don't need to parse the token value in the handler.

```go
app.Get("/profile", profileHandler,
    autofiber.WithJwtAuth(),
)
```

- Adds `security: [{bearerAuth: []}]` to the OpenAPI operation.
- Returns `401 Missing Authorization header` at runtime if the header is absent.
- The token value is available via `c.Get("Authorization")` in the handler.

### 2. Request Schema: Required `Authorization` Header

Use this when your handler needs to read and validate the token value.

```go
type ProfileRequest struct {
    Authorization string `parse:"header:Authorization" validate:"required" description:"Bearer <token>"`
    UserID        int    `parse:"path:user_id"         validate:"required"`
}

app.Get("/users/:user_id/profile", profileHandler,
    autofiber.WithRequestSchema(ProfileRequest{}),
)
```

AutoFiber detects the required `Authorization` field and automatically sets `RequireJWTAuth = true` — no need to also call `WithJwtAuth()`.

## Group-Level Auth

Apply `WithJwtAuth()` to an entire group so you don't repeat it per route.

```go
protected := app.Group("/admin").WithJwtAuth()

protected.Get("/dashboard",   dashboardHandler)
protected.Get("/users",       listUsersHandler)
protected.Delete("/users/:id", deleteUserHandler)
```

See [routing.md](routing.md) for how to combine group auth with group middleware.

## Runtime Behavior

| Condition | Result |
|---|---|
| `RequireJWTAuth = true`, header present | Request proceeds |
| `RequireJWTAuth = true`, header absent | `401 Missing Authorization header` |
| `RequireJWTAuth = false` | No auth check |

`RequireJWTAuth` is set to `true` by either `WithJwtAuth()` or auto-inference from a required `Authorization` header field in the request schema.

## OpenAPI Output

Any route with JWT auth gets:

```yaml
security:
  - bearerAuth: []
```

And the following is added once to `components.securitySchemes`:

```yaml
securitySchemes:
  bearerAuth:
    type: http
    scheme: bearer
    bearerFormat: JWT
```

Users can click **Authorize** in Swagger UI and enter a Bearer token once; it applies to all secured routes.

## Custom Auth Middleware

For more complex auth (API keys, session cookies, role checks), use `WithMiddleware`:

```go
func requireRole(role string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := c.Get("Authorization")
        if !validateToken(token, role) {
            return fiber.NewError(fiber.StatusForbidden, "Insufficient permissions")
        }
        return c.Next()
    }
}

app.Delete("/admin/users/:id", deleteUser,
    autofiber.WithJwtAuth(),
    autofiber.WithMiddleware(requireRole("admin")),
)
```
