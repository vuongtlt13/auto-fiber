# Error Handling

AutoFiber produces structured errors for parse failures and validation failures. You can customize how those errors are serialized to the HTTP response.

## Error Types

### `ParseError`

Returned when a required field is missing or a value cannot be converted.

```go
type ParseError struct {
    Field   string // struct field name or "body"
    Source  string // "query", "path", "header", "cookie", "form", "body"
    Message string
}
```

### `ValidationRequestError`

Returned when `go-playground/validator` rejects the parsed request.

```go
type ValidationRequestError struct {
    Message string
    Details []FieldErrorDetail
}

type FieldErrorDetail struct {
    Field   string // validator namespace, e.g. "CreateUserRequest.Email"
    Message string
    Tag     string // validator tag that failed, e.g. "required", "email"
}
```

### `ValidationResponseError`

Returned when response validation fails (only relevant when `WithResponseSchema` is used).

```go
type ValidationResponseError struct {
    Message string
    Details []FieldErrorDetail
}
```

## Default Behavior

Without a custom error handler, AutoFiber returns validation errors to Fiber's default error handler unchanged. The response format depends on how you configure `fiber.Config.ErrorHandler`.

## Custom Error Handler

Use `WithErrorHandler` to control the JSON shape of parse/validation errors.

```go
app := autofiber.NewWithOptions(
    fiber.Config{},
    autofiber.WithErrorHandler(func(c *fiber.Ctx, err error) error {
        switch e := err.(type) {
        case *autofiber.ValidationRequestError:
            return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
                "success": false,
                "message": e.Message,
                "errors":  e.Details,
            })
        case *autofiber.ValidationResponseError:
            // Log internally; don't expose response schema details to clients.
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "success": false,
                "message": "internal error",
            })
        default:
            return err // let Fiber handle everything else
        }
    }),
)
```

The custom handler is called for:
- `*ParseError` wrapped in a `ValidationRequestError`
- `validator.ValidationErrors` wrapped in a `ValidationRequestError`
- `*ValidationResponseError`

Errors not originating from AutoFiber (e.g. `fiber.NewError(...)` returned by your handler) are passed through unchanged.

## Fiber-Level Error Handler

For errors returned by handlers directly (business logic errors, `fiber.NewError`, etc.) configure `fiber.Config.ErrorHandler` as normal:

```go
app := autofiber.NewWithOptions(
    fiber.Config{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            code := fiber.StatusInternalServerError
            if e, ok := err.(*fiber.Error); ok {
                code = e.Code
            }
            return c.Status(code).JSON(fiber.Map{
                "success": false,
                "message": err.Error(),
            })
        },
    },
    autofiber.WithErrorHandler(...), // for parse/validation errors
)
```

The two handlers complement each other: AutoFiber's handler covers parse/validation errors; Fiber's covers everything else.

## Panic Behavior

AutoFiber panics at **registration time** (not request time) for programmer errors:

- Invalid handler signature
- Unknown parse source (e.g. `parse:"patha:id"` — typo)

These panics are intentional: they surface configuration bugs immediately on startup rather than silently failing during a live request.
