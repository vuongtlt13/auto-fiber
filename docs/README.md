# AutoFiber Documentation

## Guides

| File | What it covers |
|---|---|
| [routing.md](routing.md) | Route registration, HTTP methods, groups, group middleware, group JWT auth |
| [structs-and-tags.md](structs-and-tags.md) | `parse` tags, sources (body/query/path/header/cookie/form), embedded structs, defaults |
| [validation-rules.md](validation-rules.md) | Built-in validator rules, patterns for strings/numbers/arrays |
| [custom-validators.md](custom-validators.md) | Per-instance custom validation tags, cross-field validation |
| [authentication.md](authentication.md) | JWT/Bearer auth, `WithJwtAuth`, schema-inferred auth, custom auth middleware |
| [error-handling.md](error-handling.md) | `WithErrorHandler`, error types, custom response format |
| [complete-flow.md](complete-flow.md) | End-to-end request/response lifecycle with real examples |
| [migration-guide.md](migration-guide.md) | Migrate from older handler signatures |

## Where to Start

1. [Main README](../README.md) — installation and quick start
2. [routing.md](routing.md) — register your first routes and groups
3. [structs-and-tags.md](structs-and-tags.md) — define request/response schemas
4. [validation-rules.md](validation-rules.md) — add validation constraints
5. [authentication.md](authentication.md) — protect routes with JWT
6. [error-handling.md](error-handling.md) — customize error responses
