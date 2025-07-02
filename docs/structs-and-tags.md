# Structs and Tags Guide

This guide covers how to create request and response structs for AutoFiber, including parsing tags, validation tags, and best practices.

## Table of Contents

- [Overview](#overview)
- [Request Structs](#request-structs)
- [Response Structs](#response-structs)
- [Parse Tags](#parse-tags)
- [Validation Tags](#validation-tags)
- [JSON Tags](#json-tags)
- [Convert Functions](#convert-functions)
- [Special Cases](#special-cases)
- [Best Practices](#best-practices)

## Overview

AutoFiber uses struct tags to define:

- **Where to parse data from** (parse tags)
- **How to validate data** (validation tags)
- **How to serialize data** (json tags)
- **Documentation metadata** (description, example tags)

## Request Structs

Request structs define the structure of incoming data and how it should be parsed and validated.

### Basic Request Struct

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Name     string `json:"name" validate:"required"`
    Age      int    `json:"age" validate:"gte=18"`
}
```

### Multi-Source Request Struct

```go
type UpdateUserRequest struct {
    // Path parameter
    UserID int `parse:"path:user_id" validate:"required,min=1"`

    // Query parameters
    IncludeProfile bool `parse:"query:include_profile" validate:"omitempty"`
    Version        int  `parse:"query:version" validate:"omitempty,gte=1"`

    // Headers
    Authorization string `parse:"header:Authorization" validate:"required"`
    ContentType   string `parse:"header:Content-Type" validate:"omitempty"`

    // Body fields
    Email    string `json:"email" validate:"omitempty,email"`
    Name     string `json:"name" validate:"omitempty,min=2"`
    IsActive bool   `json:"is_active" validate:"omitempty"`
}
```

## Response Structs

Response structs define the structure of outgoing data and how it should be validated before sending.

### Basic Response Struct

```go
type UserResponse struct {
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Name      string    `json:"name" validate:"required"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
}
```

### Complex Response Struct

```go
type UserDetailResponse struct {
    ID        int       `json:"id" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    Name      string    `json:"name" validate:"required"`
    Profile   *Profile  `json:"profile,omitempty" validate:"omitempty"`
    Posts     []Post    `json:"posts,omitempty" validate:"omitempty,dive"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
    UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type Profile struct {
    Bio       string `json:"bio" validate:"omitempty,max=500"`
    Avatar    string `json:"avatar" validate:"omitempty,url"`
    Location  string `json:"location" validate:"omitempty"`
}

type Post struct {
    ID      int    `json:"id" validate:"required"`
    Title   string `json:"title" validate:"required,min=1,max=200"`
    Content string `json:"content" validate:"required,min=1"`
}
```

## OpenAPI Schema Naming

- For non-generic structs, the schema name is the type name (e.g., `LoginResponse`).
- For generic structs, the schema name is the base name plus the type parameter (e.g., `APIResponse_User` for `APIResponse[User]`).
- All schema names are sanitized to contain only alphanumeric characters and underscores, ensuring compatibility with Swagger UI and code generators.

### Example: Generic Response

```go
type APIResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

// Usage in route registration:
app.Get("/user", handler.GetUser, autofiber.WithResponseSchema(APIResponse[User]{}))
```

## Parse Tags

Parse tags specify where AutoFiber should extract data from the HTTP request.

### Parse Tag Format

```
parse:"source:key,required,default:value"
```

### Supported Sources

#### 1. Path Parameters

Extract data from URL path parameters.

```go
type GetUserRequest struct {
    UserID int `parse:"path:user_id" validate:"required,min=1"`
    OrgID  int `parse:"path:org_id" validate:"required,min=1"`
}
```

**Usage**: `GET /organizations/:org_id/users/:user_id`

**When to use**:

- For resource identifiers in RESTful APIs
- When the parameter is part of the URL structure
- For required parameters that identify the resource

**Special cases**:

- Path parameters are always strings, AutoFiber converts them to the target type
- If conversion fails, returns 400 Bad Request
- Path parameters are always required by nature

#### 2. Query Parameters

Extract data from URL query string.

```go
type ListUsersRequest struct {
    Page     int    `parse:"query:page" validate:"omitempty,gte=1"`
    Limit    int    `parse:"query:limit" validate:"omitempty,gte=1,lte=100"`
    Search   string `parse:"query:search" validate:"omitempty,min=2"`
    SortBy   string `parse:"query:sort_by" validate:"omitempty,oneof=name email created_at"`
    SortDesc bool   `parse:"query:sort_desc" validate:"omitempty"`
}
```

**Usage**: `GET /users?page=1&limit=10&search=john&sort_by=name&sort_desc=true`

**When to use**:

- For optional filtering, pagination, and sorting parameters
- For parameters that don't identify the resource
- For parameters that can have default values

**Special cases**:

- Query parameters are optional by default
- Use `required` option to make them mandatory
- Boolean values are parsed as `true` for "true", "1", "yes", "on"; `false` for others
- Arrays can be specified as `?tags=tag1&tags=tag2` or `?tags=tag1,tag2`

#### 3. Headers

Extract data from HTTP headers.

```go
type AuthenticatedRequest struct {
    Authorization string `parse:"header:Authorization" validate:"required"`
    APIKey        string `parse:"header:X-API-Key" validate:"required"`
    Accept        string `parse:"header:Accept" validate:"omitempty,oneof=application/json application/xml"`
    UserAgent     string `parse:"header:User-Agent" validate:"omitempty"`
}
```

**Usage**:

```
Authorization: Bearer token123
X-API-Key: api_key_456
Accept: application/json
```

**When to use**:

- For authentication tokens
- For API keys and credentials
- For content negotiation
- For metadata about the request

**Special cases**:

- Header names are case-insensitive
- AutoFiber automatically handles common header names
- Use `X-` prefix for custom headers

#### 4. Cookies

Extract data from HTTP cookies.

```go
type SessionRequest struct {
    SessionID string `parse:"cookie:session_id" validate:"required"`
    Theme     string `parse:"cookie:theme" validate:"omitempty,oneof=light dark"`
    Language  string `parse:"cookie:lang" validate:"omitempty,len=2"`
}
```

**Usage**:

```
Cookie: session_id=abc123; theme=dark; lang=en
```

**When to use**:

- For session management
- For user preferences
- For client-side state

**Special cases**:

- Cookies are optional by default
- Cookie values are always strings
- Use validation tags for type conversion and validation

#### 5. Form Data

Extract data from form-encoded data.

```go
type UploadRequest struct {
    Title       string `parse:"form:title" validate:"required,min=1"`
    Description string `parse:"form:description" validate:"omitempty"`
    Category    string `parse:"form:category" validate:"required,oneof=image video document"`
    Public      bool   `parse:"form:public" validate:"omitempty"`
}
```

**Usage**: `multipart/form-data` or `application/x-www-form-urlencoded`

**When to use**:

- For file uploads
- For form submissions
- When working with traditional web forms

**Special cases**:

- Form data is typically used with POST requests
- File uploads require `multipart/form-data`
- Form fields are always strings, use validation for conversion

#### 6. Body

Extract data from JSON request body.

```go
type CreatePostRequest struct {
    Title   string   `parse:"body:title" validate:"required,min=1,max=200"`
    Content string   `parse:"body:content" validate:"required,min=1"`
    Tags    []string `parse:"body:tags" validate:"omitempty,dive,min=1"`
    Draft   bool     `parse:"body:draft" validate:"omitempty"`
}
```

**Usage**: JSON body in POST/PUT/PATCH requests

**When to use**:

- For complex data structures
- For data that doesn't fit in query parameters
- For sensitive data (not visible in URL)

**Special cases**:

- Body parsing only works with POST, PUT, PATCH methods
- GET and DELETE requests don't parse body by default
- Use `json` tag for field aliasing

#### 7. Auto Detection

Let AutoFiber automatically detect the best source based on HTTP method.

```go
type SmartRequest struct {
    UserID int    `parse:"auto:user_id" validate:"required"`
    Page   int    `parse:"auto:page" validate:"omitempty,gte=1"`
    Email  string `parse:"auto:email" validate:"omitempty,email"`
    Name   string `parse:"auto:name" validate:"omitempty"`
}
```

**Usage**:

- `GET /users/:user_id?page=1` - user_id from path, page from query
- `POST /users/:user_id?page=1` - user_id from path, page from query, email/name from body

**When to use**:

- When you want flexible parsing based on HTTP method
- For reusable structs across different endpoints
- When the same field might come from different sources

**Special cases**:

- GET requests: path → query (no body)
- POST/PUT/PATCH: path → query → body
- DELETE: path → query (no body)

### Parse Tag Options

#### Required Option

```go
type RequiredRequest struct {
    UserID int    `parse:"path:user_id,required" validate:"min=1"`
    Token  string `parse:"header:Authorization,required"`
    Email  string `parse:"body:email,required" validate:"email"`
}
```

**When to use**:

- For critical parameters that must be present
- For authentication tokens
- For resource identifiers

#### Default Option

```go
type DefaultRequest struct {
    Page     int    `parse:"query:page,default:1" validate:"gte=1"`
    Limit    int    `parse:"query:limit,default:10" validate:"gte=1,lte=100"`
    SortBy   string `parse:"query:sort_by,default:created_at" validate:"oneof=name email created_at"`
    SortDesc bool   `parse:"query:sort_desc,default:true"`
}
```

**When to use**:

- For optional parameters with sensible defaults
- For pagination parameters
- For user preferences

**Special cases**:

- Default values are applied before validation
- Default values must be valid according to validation rules
- Use string representation for default values

## Validation Tags

Validation tags use the `go-playground/validator` library to validate data.

### Common Validation Rules

#### String Validations

```go
type StringValidations struct {
    Required    string `validate:"required"`
    Email       string `validate:"email"`
    URL         string `validate:"url"`
    MinLength   string `validate:"min=3"`
    MaxLength   string `validate:"max=100"`
    Length      string `validate:"len=10"`
    Alpha       string `validate:"alpha"`
    Alphanum    string `validate:"alphanum"`
    Numeric     string `validate:"numeric"`
    Lowercase   string `validate:"lowercase"`
    Uppercase   string `validate:"uppercase"`
    OneOf       string `validate:"oneof=admin user guest"`
    Regex       string `validate:"regex=^[a-zA-Z0-9]+$"`
}
```

#### Numeric Validations

```go
type NumericValidations struct {
    Required int     `validate:"required"`
    Min      int     `validate:"min=18"`
    Max      int     `validate:"max=120"`
    Gte      int     `validate:"gte=0"`
    Lte      int     `validate:"lte=100"`
    Range    int     `validate:"gte=1,lte=10"`
    Float    float64 `validate:"gte=0.0,lte=1.0"`
}
```

#### Boolean Validations

```go
type BooleanValidations struct {
    Required bool `validate:"required"`
    // Boolean fields are typically optional
    IsActive bool `validate:"omitempty"`
}
```

#### Slice Validations

```go
type SliceValidations struct {
    Required []string `validate:"required"`
    MinItems []int    `validate:"min=1"`
    MaxItems []int    `validate:"max=10"`
    Dive     []string `validate:"dive,email"` // Validate each element
    Unique   []int    `validate:"unique"`     // Ensure no duplicates
}
```

#### Struct Validations

```go
type NestedValidations struct {
    Profile Profile `validate:"required"`
    Posts   []Post  `validate:"omitempty,dive"` // Validate each Post
}

type Profile struct {
    Bio      string `validate:"omitempty,max=500"`
    Avatar   string `validate:"omitempty,url"`
    Location string `validate:"omitempty"`
}
```

### Conditional Validation

```go
type ConditionalRequest struct {
    Type      string `validate:"required,oneof=email phone"`
    Email     string `validate:"omitempty,email,required_if=Type email"`
    Phone     string `validate:"omitempty,len=10,required_if=Type phone"`
    Password  string `validate:"required_with=Email"`
    ConfirmPW string `validate:"eqfield=Password"`
}
```

**Special cases**:

- `required_if`: Required if another field equals a value
- `required_with`: Required if another field is present
- `eqfield`: Must equal another field's value
- `nefield`: Must not equal another field's value

### Custom Validation

```go
// Register custom validation
validator := autofiber.GetValidator()
validator.RegisterValidation("strong_password", validateStrongPassword)

type CustomValidation struct {
    Password string `validate:"required,strong_password"`
}

func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()

    hasUpper := false
    hasLower := false
    hasNumber := false

    for _, char := range password {
        if char >= 'A' && char <= 'Z' {
            hasUpper = true
        } else if char >= 'a' && char <= 'z' {
            hasLower = true
        } else if char >= '0' && char <= '9' {
            hasNumber = true
        }
    }

    return hasUpper && hasLower && hasNumber
}
```

## JSON Tags

JSON tags control how fields are serialized and deserialized.

### Basic JSON Tags

```go
type JSONExample struct {
    ID        int       `json:"id"`                    // Use field name as JSON key
    Email     string    `json:"email"`                 // Use "email" as JSON key
    UserName  string    `json:"user_name"`             // Snake case
    IsActive  bool      `json:"is_active"`             // Boolean with underscore
    CreatedAt time.Time `json:"created_at,omitempty"`  // Omit if empty
    UpdatedAt time.Time `json:"updated_at,omitempty"`  // Omit if empty
    Password  string    `json:"-"`                     // Never include in JSON
}
```

### JSON Tag Options

#### Omitempty

```go
type OmitEmptyExample struct {
    ID        int       `json:"id"`                    // Always included
    Email     string    `json:"email,omitempty"`       // Omitted if empty string
    Age       int       `json:"age,omitempty"`         // Omitted if 0
    IsActive  bool      `json:"is_active,omitempty"`   // Omitted if false
    Profile   *Profile  `json:"profile,omitempty"`     // Omitted if nil
    Tags      []string  `json:"tags,omitempty"`        // Omitted if empty slice
}
```

**When to use**:

- For optional fields in responses
- For fields that might be empty
- For cleaner JSON output

#### String Option

```go
type StringOptionExample struct {
    ID        int       `json:"id,string"`             // Serialize as string
    Timestamp int64     `json:"timestamp,string"`      // Unix timestamp as string
    Price     float64   `json:"price,string"`          // Price as string
}
```

**When to use**:

- When JavaScript needs exact numeric precision
- For IDs that might exceed JavaScript's number limits
- For compatibility with frontend frameworks

## Special Cases

### Nested Structs

```go
type ComplexRequest struct {
    User UserInfo `json:"user" validate:"required"`
    Address Address `json:"address" validate:"omitempty"`
}

type UserInfo struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required"`
    Age      int    `json:"age" validate:"gte=18"`
}

type Address struct {
    Street  string `json:"street" validate:"required"`
    City    string `json:"city" validate:"required"`
    Country string `json:"country" validate:"required"`
}
```

### Polymorphic Requests

```go
type PolymorphicRequest struct {
    Type      string      `json:"type" validate:"required,oneof=user admin guest"`
    UserData  *UserData   `json:"user_data,omitempty" validate:"omitempty,required_if=Type user"`
    AdminData *AdminData  `json:"admin_data,omitempty" validate:"omitempty,required_if=Type admin"`
    GuestData *GuestData  `json:"guest_data,omitempty" validate:"omitempty,required_if=Type guest"`
}

type UserData struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

type AdminData struct {
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    Role      string `json:"role" validate:"required,oneof=admin super_admin"`
}

type GuestData struct {
    Name string `json:"name" validate:"required"`
}
```

### File Uploads

```go
type FileUploadRequest struct {
    Title       string `parse:"form:title" validate:"required,min=1"`
    Description string `parse:"form:description" validate:"omitempty"`
    Category    string `parse:"form:category" validate:"required,oneof=image video document"`
    Public      bool   `parse:"form:public" validate:"omitempty"`
    // File handling is done separately in the handler
}
```

### Pagination

```go
type PaginationRequest struct {
    Page     int    `parse:"query:page" validate:"omitempty,gte=1"`
    Limit    int    `parse:"query:limit" validate:"omitempty,gte=1,lte=100"`
    SortBy   string `parse:"query:sort_by" validate:"omitempty,oneof=created_at updated_at name email"`
    SortDesc bool   `parse:"query:sort_desc" validate:"omitempty"`
}

type PaginationResponse struct {
    Data       []interface{} `json:"data" validate:"required"`
    Page       int           `json:"page" validate:"required,gte=1"`
    Limit      int           `json:"limit" validate:"required,gte=1"`
    Total      int           `json:"total" validate:"required,gte=0"`
    TotalPages int           `json:"total_pages" validate:"required,gte=0"`
    HasNext    bool          `json:"has_next" validate:"required"`
    HasPrev    bool          `json:"has_prev" validate:"required"`
}
```

## Best Practices

### 1. Use Descriptive Field Names

```go
// Good
type CreateUserRequest struct {
    EmailAddress string `json:"email" validate:"required,email"`
    UserPassword string `json:"password" validate:"required,min=6"`
    FullName     string `json:"name" validate:"required"`
}

// Avoid
type CreateUserRequest struct {
    Email string `json:"e" validate:"required,email"`
    Pass  string `json:"p" validate:"required,min=6"`
    Name  string `json:"n" validate:"required"`
}
```

### 2. Group Related Fields

```go
type UserRequest struct {
    // Basic info
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Name     string `json:"name" validate:"required"`

    // Profile info
    Age      int    `json:"age" validate:"omitempty,gte=18"`
    Bio      string `json:"bio" validate:"omitempty,max=500"`
    Location string `json:"location" validate:"omitempty"`

    // Settings
    IsActive bool `json:"is_active" validate:"omitempty"`
    Theme    string `json:"theme" validate:"omitempty,oneof=light dark"`
}
```

### 3. Use Consistent Validation Rules

```go
// Define common validation patterns
const (
    EmailValidation    = "required,email"
    PasswordValidation = "required,min=6"
    NameValidation     = "required,min=2,max=100"
    AgeValidation      = "omitempty,gte=18,lte=120"
)

type ConsistentRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Age      int    `json:"age" validate:"omitempty,gte=18,lte=120"`
}
```

### 4. Handle Optional Fields Properly

```go
type OptionalFields struct {
    // Required fields
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`

    // Optional fields with defaults
    IsActive bool   `json:"is_active,omitempty" validate:"omitempty"`
    Role     string `json:"role,omitempty" validate:"omitempty,oneof=user admin"`

    // Optional fields that should be omitted when empty
    Bio      string `json:"bio,omitempty" validate:"omitempty,max=500"`
    Avatar   string `json:"avatar,omitempty" validate:"omitempty,url"`
}
```

### 5. Use Pointers for Optional Complex Types

```go
type ComplexOptional struct {
    // Required simple fields
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name" validate:"required"`

    // Optional complex fields (use pointers)
    Profile *Profile `json:"profile,omitempty" validate:"omitempty"`
    Address *Address `json:"address,omitempty" validate:"omitempty"`

    // Optional slices
    Tags []string `json:"tags,omitempty" validate:"omitempty,dive,min=1"`
}
```

### 6. Document Your Structs

```go
type WellDocumentedRequest struct {
    // User's email address for account creation
    Email string `json:"email" validate:"required,email" description:"User email address" example:"user@example.com"`

    // User's password (minimum 6 characters)
    Password string `json:"password" validate:"required,min=6" description:"User password" example:"password123"`

    // User's full name
    Name string `json:"name" validate:"required,min=2,max=100" description:"User full name" example:"John Doe"`

    // User's age (optional, must be 18 or older)
    Age int `json:"age,omitempty" validate:"omitempty,gte=18" description:"User age" example:"25"`
}
```

### 7. Test Your Structs

```go
func TestCreateUserRequest(t *testing.T) {
    tests := []struct {
        name    string
        request CreateUserRequest
        wantErr bool
    }{
        {
            name: "valid request",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Password: "password123",
                Name:     "John Doe",
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            request: CreateUserRequest{
                Email:    "invalid-email",
                Password: "password123",
                Name:     "John Doe",
            },
            wantErr: true,
        },
        {
            name: "password too short",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Password: "123",
                Name:     "John Doe",
            },
            wantErr: true,
        },
    }

    validator := autofiber.GetValidator()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Struct(tt.request)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateUserRequest validation error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Convert Functions

AutoFiber provides two specialized functions for converting Go structs to OpenAPI schemas with different behaviors for request and response scenarios.

### ConvertRequestToOpenAPISchema

This function converts a Go struct to OpenAPI schema specifically for **request parsing**. It follows these rules:

1. **Parse tags with body source**: Only fields with `parse:"body:..."` tags are included
2. **JSON tags as fallback**: If no parse tag, only fields with valid `json` tags (not empty, not "-") are included
3. **Skip fields**: Fields without parse tags and without valid json tags are skipped

#### Example

```go
type ExampleRequest struct {
    // ✅ Included: parse tag with body source
    UserID   int    `parse:"body:user_id" json:"id" validate:"required"`
    UserName string `parse:"body:user_name" json:"name" validate:"required"`

    // ❌ Skipped: parse tag but not body source
    Token    string `parse:"header:Authorization" json:"token"`
    Page     int    `parse:"query:page" json:"page"`

    // ✅ Included: valid json tag (no parse tag)
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`

    // ❌ Skipped: empty json tag
    SkipMe   string `json:"" validate:"required"`
    SkipMe2  string `json:"," validate:"required"`

    // ❌ Skipped: no tags
    NoTags   string
}

// Usage
dg := autofiber.NewDocsGenerator()
schema := dg.ConvertRequestToOpenAPISchema(ExampleRequest{})
// Result: only user_id, user_name, email, password fields are included
```

### ConvertResponseToOpenAPISchema

This function converts a Go struct to OpenAPI schema specifically for **response serialization**. It follows these rules:

1. **JSON tags priority**: Fields with `json` tags use the tag name
2. **CamelCase fallback**: Fields without json tags use camelCase field names
3. **Skip fields**: Fields with `json:"-"` are skipped

#### Example

```go
type ExampleResponse struct {
    // ✅ Included: uses json tag name
    ID        int       `json:"id" validate:"required"`
    Name      string    `json:"name" validate:"required"`
    Email     string    `json:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
    IsActive  bool      `json:"is_active"`

    // ✅ Included: uses camelCase field name
    UserType  string // becomes "userType"
    APIKey    string // becomes "apiKey"
    HTTPStatus string // becomes "httpStatus"

    // ❌ Skipped: json:"-" tag
    SkipMe    string `json:"-"`
}

// Usage
dg := autofiber.NewDocsGenerator()
schema := dg.ConvertResponseToOpenAPISchema(ExampleResponse{})
// Result: id, name, email, created_at, is_active, userType, apiKey, httpStatus fields are included
```

### When to Use Each Function

- **ConvertRequestToOpenAPISchema**: Use when generating OpenAPI schemas for request bodies, especially when you want to exclude non-body fields (headers, query params, path params)
- **ConvertResponseToOpenAPISchema**: Use when generating OpenAPI schemas for response bodies, ensuring all serializable fields are included with proper naming

### Integration with AutoFiber

These functions are used internally by AutoFiber when generating OpenAPI documentation. You can also use them directly in your code for custom schema generation:

```go
func (h *MyHandler) CustomSchemaExample(c *fiber.Ctx) (interface{}, error) {
    dg := autofiber.NewDocsGenerator()

    // Generate request schema
    requestSchema := dg.ConvertRequestToOpenAPISchema(MyRequest{})

    // Generate response schema
    responseSchema := dg.ConvertResponseToOpenAPISchema(MyResponse{})

    return fiber.Map{
        "request_schema": requestSchema,
        "response_schema": responseSchema,
    }, nil
}
```

## Common Patterns

### CRUD Operations

```go
// Create
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
    Name     string `json:"name" validate:"required"`
}

// Read (with filters)
type GetUsersRequest struct {
    Page     int    `parse:"query:page" validate:"omitempty,gte=1"`
    Limit    int    `parse:"query:limit" validate:"omitempty,gte=1,lte=100"`
    Search   string `parse:"query:search" validate:"omitempty,min=2"`
    Role     string `parse:"query:role" validate:"omitempty,oneof=user admin"`
}

// Update
type UpdateUserRequest struct {
    UserID int    `parse:"path:user_id" validate:"required,min=1"`
    Email  string `json:"email,omitempty" validate:"omitempty,email"`
    Name   string `json:"name,omitempty" validate:"omitempty,min=2"`
    Age    int    `json:"age,omitempty" validate:"omitempty,gte=18"`
}

// Delete
type DeleteUserRequest struct {
    UserID int `parse:"path:user_id" validate:"required,min=1"`
}
```

### Authentication

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type AuthenticatedRequest struct {
    Authorization string `parse:"header:Authorization" validate:"required"`
    UserID        int    `parse:"path:user_id" validate:"required,min=1"`
}

type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}
```

### Search and Filtering

```go
type SearchRequest struct {
    Query     string   `parse:"query:q" validate:"omitempty,min=2"`
    Categories []string `parse:"query:categories" validate:"omitempty,dive,oneof=tech sports news"`
    DateFrom  string   `parse:"query:date_from" validate:"omitempty,datetime=2006-01-02"`
    DateTo    string   `parse:"query:date_to" validate:"omitempty,datetime=2006-01-02"`
    SortBy    string   `parse:"query:sort_by" validate:"omitempty,oneof=relevance date title"`
    SortDesc  bool     `parse:"query:sort_desc" validate:"omitempty"`
}
```

This comprehensive guide should help you create robust and well-structured request/response structs for AutoFiber applications.
