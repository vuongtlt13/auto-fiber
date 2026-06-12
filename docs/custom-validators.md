# Custom Validators

AutoFiber uses `go-playground/validator/v10` for all request validation. You can register custom validation tags at both the global and per-instance level.

## Global Validator

The global validator is shared across all `AutoFiber` instances that do not configure their own.

```go
v := autofiber.GetValidator()
v.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
    p := fl.Field().String()
    return len(p) >= 8 && strings.ContainsAny(p, "0123456789")
})

// Now usable in any struct validated by AutoFiber:
type RegisterRequest struct {
    Password string `json:"password" validate:"required,strong_password"`
}
```

## Per-Instance Validator (Recommended)

Register validators on a specific `AutoFiber` instance so they don't pollute the global state. Two options:

### Option 1: `WithValidatorSetup` at construction

```go
app := autofiber.NewWithOptions(
    fiber.Config{},
    autofiber.WithValidatorSetup(func(v *validator.Validate) {
        v.RegisterValidation("strong_password", validateStrongPassword)
        v.RegisterValidation("username",        validateUsername)
    }),
)
```

### Option 2: `RegisterValidator` after construction

```go
app := autofiber.New()

if err := app.RegisterValidator("strong_password", validateStrongPassword); err != nil {
    log.Fatal(err)
}
```

Both methods register on the same underlying `*validator.Validate` instance owned by `app`.

## Writing a Validator Function

```go
func validateStrongPassword(fl validator.FieldLevel) bool {
    p := fl.Field().String()
    if len(p) < 8 {
        return false
    }
    hasDigit := strings.ContainsAny(p, "0123456789")
    hasUpper := strings.ContainsAny(p, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
    return hasDigit && hasUpper
}
```

## Using the Tag in a Struct

```go
type RegisterRequest struct {
    Email    string `json:"email"    validate:"required,email"`
    Password string `json:"password" validate:"required,strong_password"`
    Username string `json:"username" validate:"required,username"`
}

app.Post("/register", registerHandler,
    autofiber.WithRequestSchema(RegisterRequest{}),
)
```

## Cross-Field Validation

Use `validator.StructLevel` for rules that span multiple fields:

```go
app := autofiber.NewWithOptions(
    fiber.Config{},
    autofiber.WithValidatorSetup(func(v *validator.Validate) {
        v.RegisterStructValidation(func(sl validator.StructLevel) {
            req := sl.Current().Interface().(ChangePasswordRequest)
            if req.NewPassword == req.OldPassword {
                sl.ReportError(req.NewPassword, "new_password", "NewPassword", "different_from_old", "")
            }
        }, ChangePasswordRequest{})
    }),
)
```

## Validation Error Shape

When a custom validator fails, `ValidationRequestError.Details` contains:

```json
{
  "Field": "ChangePasswordRequest.NewPassword",
  "Message": "Key: 'ChangePasswordRequest.NewPassword' Error: ...",
  "Tag": "different_from_old"
}
```

See [error-handling.md](error-handling.md) for how to customize the response format.
