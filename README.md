# AutoFiber

A FastAPI-like wrapper for [Fiber](https://github.com/gofiber/fiber) that adds automatic request parsing, validation, and OpenAPI/Swagger documentation — from struct tags alone.

## Features

- **Multi-source parsing** — body, query, path, header, cookie, form via a single `parse` tag
- **Automatic validation** — `go-playground/validator/v10` applied on every request
- **OpenAPI 3.0 docs** — generated from struct tags; served as Swagger UI
- **JWT auth** — declare once per route or group; runtime 401 on missing header
- **Group middleware** — attach middleware or JWT auth to an entire route group
- **Custom error handler** — control how validation errors are formatted
- **Custom validators** — register per-instance validation tags
- **File downloads** — return CSV/PDF responses, bypassing JSON serialization
- **Response validation** — validate outgoing data against a schema before sending

## Installation

```sh
go get github.com/vuongtlt13/auto-fiber
```

## Quick Start

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    autofiber "github.com/vuongtlt13/auto-fiber"
)

type CreateUserRequest struct {
    OrgID int    `parse:"path:org_id"  validate:"required"`
    Role  string `parse:"query:role"   validate:"required,oneof=admin user"`
    Email string `json:"email"         validate:"required,email"`
    Name  string `json:"name"          validate:"required"`
}

type UserResponse struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
    Role  string `json:"role"`
}

func createUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    return UserResponse{ID: 1, Email: req.Email, Name: req.Name, Role: req.Role}, nil
}

func main() {
    app := autofiber.New(fiber.Config{},
        autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
            Title:   "My API",
            Version: "1.0.0",
        }),
    )

    app.Post("/orgs/:org_id/users", createUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
        autofiber.WithTags("users"),
    )

    app.ServeDocs("/docs")
    app.ServeSwaggerUI("/swagger", "/docs")
    app.Listen(":3000")
}
```

## What You Get

### Successful request

```
POST /orgs/42/users?role=admin
Content-Type: application/json

{"email": "jane@example.com", "name": "Jane Doe"}
```

```json
{
  "id": 1,
  "email": "jane@example.com",
  "name": "Jane Doe",
  "role": "admin"
}
```

### Automatic validation error

Send an invalid request — missing `name`, invalid `email`, unknown `role`:

```
POST /orgs/42/users?role=superadmin
Content-Type: application/json

{"email": "not-an-email"}
```

```
HTTP 422 Unprocessable Entity
```

```json
{
  "message": "Validation failed",
  "details": [
    {
      "field": "CreateUserRequest.Role",
      "message": "Key: 'CreateUserRequest.Role' Error:Field validation for 'Role' failed on the 'oneof' tag",
      "tag": "oneof"
    },
    {
      "field": "CreateUserRequest.Email",
      "message": "Key: 'CreateUserRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag",
      "tag": "email"
    },
    {
      "field": "CreateUserRequest.Name",
      "message": "Key: 'CreateUserRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag",
      "tag": "required"
    }
  ]
}
```

No error-handling code needed in your handler — AutoFiber generates this from the `validate` tags.

### Auto-generated OpenAPI spec

`GET /docs` returns a complete OpenAPI 3.0 document. The `POST /orgs/{org_id}/users` operation looks like:

```json
{
  "paths": {
    "/orgs/{org_id}/users": {
      "post": {
        "tags": ["users"],
        "parameters": [
          { "name": "org_id", "in": "path",  "required": true, "schema": { "type": "integer" } },
          { "name": "role",   "in": "query", "required": true, "schema": { "type": "string", "enum": ["admin", "user"] } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/CreateUserRequest" }
            }
          }
        },
        "responses": {
          "200": {
            "content": {
              "application/json": {
                "schema": { "$ref": "#/components/schemas/UserResponse" }
              }
            }
          }
        }
      }
    }
  }
}
```

Visit `/swagger` for the interactive Swagger UI where you can try every endpoint directly in the browser.

## Documentation

| Topic | Description |
|---|---|
| [Routing](docs/routing.md) | Route registration, groups, group middleware |
| [Request Parsing](docs/structs-and-tags.md) | `parse` tags, sources, embedded structs, defaults |
| [Validation](docs/validation-rules.md) | Built-in rules, patterns |
| [Custom Validators](docs/custom-validators.md) | Per-instance custom validation tags |
| [Authentication](docs/authentication.md) | JWT auth, `WithJwtAuth`, schema-inferred auth |
| [Error Handling](docs/error-handling.md) | Custom error format, error types |
| [Complete Flow](docs/complete-flow.md) | End-to-end request/response lifecycle |
| [Migration Guide](docs/migration-guide.md) | Migrate from older handler signatures |

## License

MIT
