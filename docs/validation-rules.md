# Validation Rules and Best Practices

This guide covers validation rules, custom validators, and best practices for AutoFiber applications.

## Table of Contents

- [Built-in Validation Rules](#built-in-validation-rules)
- [Custom Validators](#custom-validators)
- [Validation Best Practices](#validation-best-practices)
- [Common Validation Patterns](#common-validation-patterns)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Built-in Validation Rules

AutoFiber uses `go-playground/validator/v10` for validation. Here are the most commonly used rules:

### String Validation

```go
type StringValidation struct {
    // Required field
    Name string `validate:"required"`

    // Email validation
    Email string `validate:"required,email"`

    // Length constraints
    Username string `validate:"required,min=3,max=20"`
    Bio      string `validate:"omitempty,max=500"`

    // Pattern matching
    Phone    string `validate:"required,e164"`           // International phone format
    URL      string `validate:"omitempty,url"`           // Valid URL
    UUID     string `validate:"required,uuid"`           // UUID format
    Alpha    string `validate:"required,alpha"`          // Alphabetic characters only
    Alphanum string `validate:"required,alphanum"`       // Alphanumeric characters
    Numeric  string `validate:"required,numeric"`        // Numeric characters only

    // Case validation
    Lowercase string `validate:"required,lowercase"`     // Must be lowercase
    Uppercase string `validate:"required,uppercase"`     // Must be uppercase

    // Specific formats
    Date     string `validate:"required,datetime=2006-01-02"`           // Date format
    DateTime string `validate:"required,datetime=2006-01-02T15:04:05Z"` // DateTime format
    Time     string `validate:"required,datetime=15:04:05"`             // Time format
}
```

### Numeric Validation

```go
type NumericValidation struct {
    // Integer validation
    Age     int `validate:"required,min=18,max=120"`
    Score   int `validate:"required,gte=0,lte=100"`
    UserID  int `validate:"required,min=1"`

    // Float validation
    Price   float64 `validate:"required,min=0.01"`
    Rating  float64 `validate:"omitempty,gte=0,lte=5"`
    Weight  float64 `validate:"required,gt=0"`

    // Range validation
    Page    int `validate:"omitempty,gte=1"`
    Limit   int `validate:"omitempty,gte=1,lte=100"`
}
```

### Boolean Validation

```go
type BooleanValidation struct {
    // Boolean fields
    IsActive   bool `validate:"required"`
    IsVerified bool `validate:"omitempty"`
    IsPublic   bool `validate:"omitempty"`
}
```

### Array/Slice Validation

```go
type ArrayValidation struct {
    // Array with length constraints
    Tags      []string `validate:"required,min=1,max=10"`
    Categories []string `validate:"omitempty,dive,oneof=tech sports news"`

    // Array of objects
    Items     []Item   `validate:"required,min=1,dive"`

    // Array of numbers
    Scores    []int    `validate:"omitempty,dive,gte=0,lte=100"`
}
```

### Struct Validation

```go
type NestedValidation struct {
    // Nested struct validation
    Profile   *Profile `validate:"required"`
    Address   Address  `validate:"omitempty"`

    // Array of structs
    Orders    []Order  `validate:"omitempty,dive"`
}

type Profile struct {
    FirstName string `validate:"required,min=2"`
    LastName  string `validate:"required,min=2"`
    Bio       string `validate:"omitempty,max=500"`
}

type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    Country string `validate:"required,oneof=US CA UK"`
    ZipCode string `validate:"required"`
}
```

### Enum Validation

```go
type EnumValidation struct {
    // String enums
    Role      string `validate:"required,oneof=admin user guest"`
    Status    string `validate:"required,oneof=active inactive pending"`
    Language  string `validate:"omitempty,oneof=en es fr de"`

    // Integer enums
    Priority  int    `validate:"required,oneof=1 2 3 4 5"`
    Category  int    `validate:"omitempty,oneof=1 2 3"`
}
```

### Conditional Validation

```go
type ConditionalValidation struct {
    // Conditional required fields
    Email     string `validate:"required,email"`
    Password  string `validate:"required_if=Email,min=6"`

    // Conditional validation based on other fields
    Age       int    `validate:"required,min=18"`
    ParentID  int    `validate:"required_if=Age lt 18"`

    // Cross-field validation
    StartDate time.Time `validate:"required"`
    EndDate   time.Time `validate:"required,gtfield=StartDate"`
}
```

## Custom Validators

You can create custom validation functions for specific business rules.

### Registering Custom Validators

```go
func main() {
    app := autofiber.New()

    // Get the validator instance
    validator := autofiber.GetValidator()

    // Register custom validators
    validator.RegisterValidation("strong_password", validateStrongPassword)
    validator.RegisterValidation("unique_email", validateUniqueEmail)
    validator.RegisterValidation("valid_phone", validatePhoneNumber)
    validator.RegisterValidation("future_date", validateFutureDate)

    // ... rest of your app
}
```

### Custom Validator Examples

```go
// Strong password validation
func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()

    // Check minimum length
    if len(password) < 8 {
        return false
    }

    // Check for uppercase, lowercase, number, and special character
    hasUpper := false
    hasLower := false
    hasNumber := false
    hasSpecial := false

    for _, char := range password {
        switch {
        case char >= 'A' && char <= 'Z':
            hasUpper = true
        case char >= 'a' && char <= 'z':
            hasLower = true
        case char >= '0' && char <= '9':
            hasNumber = true
        case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
            hasSpecial = true
        }
    }

    return hasUpper && hasLower && hasNumber && hasSpecial
}

// Unique email validation (example with database check)
func validateUniqueEmail(fl validator.FieldLevel) bool {
    email := fl.Field().String()

    // In a real application, you would check against your database
    // This is just an example
    existingEmails := []string{"admin@example.com", "user@example.com"}

    for _, existing := range existingEmails {
        if email == existing {
            return false
        }
    }

    return true
}

// Phone number validation
func validatePhoneNumber(fl validator.FieldLevel) bool {
    phone := fl.Field().String()

    // Remove spaces, dashes, and parentheses
    cleaned := strings.ReplaceAll(phone, " ", "")
    cleaned = strings.ReplaceAll(cleaned, "-", "")
    cleaned = strings.ReplaceAll(cleaned, "(", "")
    cleaned = strings.ReplaceAll(cleaned, ")", "")

    // Check if it's a valid phone number (basic check)
    if len(cleaned) < 10 || len(cleaned) > 15 {
        return false
    }

    // Check if all characters are digits
    for _, char := range cleaned {
        if char < '0' || char > '9' {
            return false
        }
    }

    return true
}

// Future date validation
func validateFutureDate(fl validator.FieldLevel) bool {
    date := fl.Field().Interface().(time.Time)
    return date.After(time.Now())
}

// Custom validation with parameters
func validateMinAge(fl validator.FieldLevel) bool {
    minAge := 18

    // Get parameter from tag if provided
    if param := fl.Param(); param != "" {
        if age, err := strconv.Atoi(param); err == nil {
            minAge = age
        }
    }

    birthDate := fl.Field().Interface().(time.Time)
    age := time.Now().Year() - birthDate.Year()

    // Adjust age if birthday hasn't occurred this year
    if time.Now().YearDay() < birthDate.YearDay() {
        age--
    }

    return age >= minAge
}
```

### Using Custom Validators

```go
type UserRegistration struct {
    Email       string    `validate:"required,email,unique_email"`
    Password    string    `validate:"required,strong_password"`
    Phone       string    `validate:"required,valid_phone"`
    BirthDate   time.Time `validate:"required,min_age=18"`
    EventDate   time.Time `validate:"required,future_date"`
}
```

## Validation Best Practices

### 1. Use Appropriate Validation Rules

```go
// Good: Specific and meaningful validation
type UserRequest struct {
    Email     string `validate:"required,email"`
    Password  string `validate:"required,min=8,strong_password"`
    Age       int    `validate:"required,min=18,max=120"`
    Role      string `validate:"required,oneof=admin user guest"`
    Phone     string `validate:"required,valid_phone"`
}

// Avoid: Too generic or missing validation
type UserRequest struct {
    Email     string `validate:"required"`           // Missing email format
    Password  string `validate:"required"`           // Missing length and strength
    Age       int    `validate:"required"`           // Missing range
    Role      string `validate:"required"`           // Missing enum validation
    Phone     string `validate:"required"`           // Missing format validation
}
```

### 2. Separate Request and Response Validation

```go
// Request validation (what users send)
type CreateUserRequest struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
    Name     string `validate:"required,min=2,max=50"`
}

// Response validation (what you send back)
type UserResponse struct {
    ID        int       `validate:"required,min=1"`
    Email     string    `validate:"required,email"`
    Name      string    `validate:"required,min=2,max=50"`
    CreatedAt time.Time `validate:"required"`
    UpdatedAt time.Time `validate:"required"`
}
```

### 3. Use Conditional Validation

```go
type ConditionalRequest struct {
    // Basic fields
    Email     string `validate:"required,email"`
    Password  string `validate:"required,min=8"`

    // Conditional fields
    Phone     string `validate:"required_if=TwoFactorEnabled,valid_phone"`
    TwoFactorEnabled bool `validate:"omitempty"`

    // Age-based validation
    Age       int    `validate:"required,min=13"`
    ParentEmail string `validate:"required_if=Age lt 18,email"`
}
```

### 4. Validate Nested Objects

```go
type OrderRequest struct {
    Customer Customer `validate:"required"`
    Items    []Item   `validate:"required,min=1,dive"`
    Shipping Address `validate:"required"`
}

type Customer struct {
    Name  string `validate:"required,min=2"`
    Email string `validate:"required,email"`
    Phone string `validate:"required,valid_phone"`
}

type Item struct {
    ProductID int     `validate:"required,min=1"`
    Quantity  int     `validate:"required,min=1,max=100"`
    Price     float64 `validate:"required,min=0.01"`
}

type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    State   string `validate:"required"`
    ZipCode string `validate:"required"`
    Country string `validate:"required,oneof=US CA UK"`
}
```

### 5. Use Cross-Field Validation

```go
type DateRangeRequest struct {
    StartDate time.Time `validate:"required"`
    EndDate   time.Time `validate:"required,gtfield=StartDate"`

    // Business rule: end date must be within 30 days of start date
    // This would require a custom validator
}

type PasswordChangeRequest struct {
    CurrentPassword string `validate:"required"`
    NewPassword     string `validate:"required,min=8,strong_password"`
    ConfirmPassword string `validate:"required,eqfield=NewPassword"`
}
```

## Common Validation Patterns

### User Registration

```go
type UserRegistration struct {
    // Basic information
    Email       string `validate:"required,email,unique_email"`
    Password    string `validate:"required,min=8,strong_password"`
    ConfirmPassword string `validate:"required,eqfield=Password"`

    // Personal information
    FirstName   string `validate:"required,min=2,max=50,alpha"`
    LastName    string `validate:"required,min=2,max=50,alpha"`
    BirthDate   time.Time `validate:"required,min_age=18"`

    // Contact information
    Phone       string `validate:"required,valid_phone"`
    Address     Address `validate:"required"`

    // Terms and conditions
    AcceptTerms bool `validate:"required,eq=true"`

    // Optional fields
    Bio         string `validate:"omitempty,max=500"`
    Website     string `validate:"omitempty,url"`
}
```

### Product Management

```go
type CreateProductRequest struct {
    // Basic product info
    Name        string  `validate:"required,min=2,max=100"`
    Description string  `validate:"required,min=10,max=1000"`
    SKU         string  `validate:"required,alphanum,unique_sku"`

    // Pricing
    Price       float64 `validate:"required,min=0.01"`
    SalePrice   float64 `validate:"omitempty,min=0.01,ltefield=Price"`

    // Inventory
    Stock       int     `validate:"required,min=0"`
    MinStock    int     `validate:"omitempty,min=0,ltefield=Stock"`

    // Categories
    CategoryID  int     `validate:"required,min=1"`
    Tags        []string `validate:"omitempty,dive,min=2,max=20"`

    // Status
    IsActive    bool    `validate:"required"`
    IsFeatured  bool    `validate:"omitempty"`

    // Images
    Images      []string `validate:"omitempty,dive,url"`
}
```

### API Pagination

```go
type PaginationRequest struct {
    Page     int    `validate:"omitempty,gte=1"`
    Limit    int    `validate:"omitempty,gte=1,lte=100"`
    SortBy   string `validate:"omitempty,oneof=id name created_at updated_at"`
    SortDesc bool   `validate:"omitempty"`
    Search   string `validate:"omitempty,min=2,max=100"`
}
```

### Search and Filtering

```go
type SearchRequest struct {
    // Search query
    Query     string   `validate:"omitempty,min=2,max=200"`

    // Filters
    Categories []string `validate:"omitempty,dive,oneof=tech sports news entertainment"`
    DateFrom   string   `validate:"omitempty,datetime=2006-01-02"`
    DateTo     string   `validate:"omitempty,datetime=2006-01-02"`
    PriceMin   float64  `validate:"omitempty,min=0"`
    PriceMax   float64  `validate:"omitempty,min=0,gtefield=PriceMin"`

    // Sorting
    SortBy    string `validate:"omitempty,oneof=relevance date price rating"`
    SortDesc  bool   `validate:"omitempty"`

    // Pagination
    Page      int    `validate:"omitempty,gte=1"`
    Limit     int    `validate:"omitempty,gte=1,lte=50"`
}
```

## Error Handling

### Custom Error Messages

```go
// Register custom error messages
func registerCustomMessages(validator *validator.Validate) {
    validator.RegisterTranslation("required", trans, func(ut ut.Translator) error {
        return ut.Add("required", "{0} is required", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("required", fe.Field())
        return t
    })

    validator.RegisterTranslation("email", trans, func(ut ut.Translator) error {
        return ut.Add("email", "{0} must be a valid email address", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("email", fe.Field())
        return t
    })

    validator.RegisterTranslation("strong_password", trans, func(ut ut.Translator) error {
        return ut.Add("strong_password", "{0} must contain uppercase, lowercase, number, and special character", true)
    }, func(ut ut.Translator, fe validator.FieldError) string {
        t, _ := ut.T("strong_password", fe.Field())
        return t
    })
}
```

### Validation Error Response

```go
type ValidationError struct {
    Field   string `json:"field"`
    Tag     string `json:"tag"`
    Value   string `json:"value"`
    Message string `json:"message"`
}

type ValidationErrorResponse struct {
    Error   string            `json:"error"`
    Details []ValidationError `json:"details"`
}

// Convert validator errors to structured response
func formatValidationErrors(err error) ValidationErrorResponse {
    var errors []ValidationError

    for _, err := range err.(validator.ValidationErrors) {
        errors = append(errors, ValidationError{
            Field:   err.Field(),
            Tag:     err.Tag(),
            Value:   err.Param(),
            Message: err.Error(),
        })
    }

    return ValidationErrorResponse{
        Error:   "Validation failed",
        Details: errors,
    }
}
```

## Examples

### Complete User Management with Validation

```go
// Request schemas with comprehensive validation
type CreateUserRequest struct {
    Email       string    `validate:"required,email,unique_email"`
    Password    string    `validate:"required,min=8,strong_password"`
    FirstName   string    `validate:"required,min=2,max=50,alpha"`
    LastName    string    `validate:"required,min=2,max=50,alpha"`
    BirthDate   time.Time `validate:"required,min_age=18"`
    Phone       string    `validate:"required,valid_phone"`
    Role        string    `validate:"required,oneof=admin user guest"`
    IsActive    bool      `validate:"required"`
}

type UpdateUserRequest struct {
    UserID      int       `validate:"required,min=1"`
    Email       string    `validate:"omitempty,email,unique_email"`
    FirstName   string    `validate:"omitempty,min=2,max=50,alpha"`
    LastName    string    `validate:"omitempty,min=2,max=50,alpha"`
    Phone       string    `validate:"omitempty,valid_phone"`
    Role        string    `validate:"omitempty,oneof=admin user guest"`
    IsActive    bool      `validate:"omitempty"`
}

// Response schemas with validation
type UserResponse struct {
    ID          int       `validate:"required,min=1"`
    Email       string    `validate:"required,email"`
    FirstName   string    `validate:"required,min=2,max=50"`
    LastName    string    `validate:"required,min=2,max=50"`
    FullName    string    `validate:"required"`
    Role        string    `validate:"required,oneof=admin user guest"`
    IsActive    bool      `validate:"required"`
    CreatedAt   time.Time `validate:"required"`
    UpdatedAt   time.Time `validate:"required"`
}

// Handler with validation
func (h *UserHandler) CreateUser(c *fiber.Ctx, req *CreateUserRequest) (interface{}, error) {
    // Business logic here
    user := UserResponse{
        ID:        1,
        Email:     req.Email,
        FirstName: req.FirstName,
        LastName:  req.LastName,
        FullName:  req.FirstName + " " + req.LastName,
        Role:      req.Role,
        IsActive:  req.IsActive,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    return user, nil
}

// Route registration with validation
func main() {
    app := autofiber.New()

    // Register custom validators
    validator := autofiber.GetValidator()
    validator.RegisterValidation("strong_password", validateStrongPassword)
    validator.RegisterValidation("unique_email", validateUniqueEmail)
    validator.RegisterValidation("valid_phone", validatePhoneNumber)
    validator.RegisterValidation("min_age", validateMinAge)

    handler := &UserHandler{}

    app.Post("/users", handler.CreateUser,
        autofiber.WithRequestSchema(CreateUserRequest{}),
        autofiber.WithResponseSchema(UserResponse{}),
        autofiber.WithDescription("Create a new user with comprehensive validation"),
        autofiber.WithTags("users", "admin"),
    )

    app.Listen(":3000")
}
```

This comprehensive validation guide ensures that your AutoFiber applications have robust data validation and provide clear, helpful error messages to users.
