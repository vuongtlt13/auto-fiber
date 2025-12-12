// Package autofiber provides core types and configuration for the AutoFiber web framework.
package autofiber

import (
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// RouteOption is a function that configures route options for an endpoint.
type RouteOption func(*RouteOptions)

// RouteOptions contains configuration for a route, such as schemas, middleware, and metadata.
type RouteOptions struct {
	RequestSchema  interface{}     // Struct for request parsing and validation
	ResponseSchema interface{}     // Struct for response validation and documentation
	Middleware     []fiber.Handler // Middleware handlers for the route
	Description    string          // Description for API documentation
	Tags           []string        // Tags for API documentation
	RequireJWTAuth bool            // Require HTTP Bearer (JWT) auth for this route (OpenAPI security)
}

// ParseSource defines where a field should be parsed from (e.g., body, query, path, header, etc.).
type ParseSource string

const (
	// Body indicates the field should be parsed from the request body.
	Body ParseSource = "body"
	// Query indicates the field should be parsed from the query string.
	Query ParseSource = "query"
	// Path indicates the field should be parsed from the URL path parameters.
	Path ParseSource = "path"
	// Header indicates the field should be parsed from the request headers.
	Header ParseSource = "header"
	// Cookie indicates the field should be parsed from cookies.
	Cookie ParseSource = "cookie"
	// Form indicates the field should be parsed from form data.
	Form ParseSource = "form"
	// Auto enables smart parsing based on HTTP method and struct tags.
	Auto ParseSource = "auto"
)

// FieldInfo contains parsing information for a struct field.
type FieldInfo struct {
	Source      ParseSource // Source to parse the field from
	Key         string      // Custom key name (e.g., "user_id" for "UserId")
	Required    bool        // Whether the field is required
	Default     interface{} // Default value if not provided
	Description string      // Description for documentation
}

// ParseError represents a parsing error for a specific field and source.
type ParseError struct {
	Field   string // Name of the field
	Source  string // Source of the field (e.g., body, query)
	Message string // Error message
}

// Error returns the error message for a ParseError.
func (e *ParseError) Error() string {
	return e.Field + " (" + e.Source + "): " + e.Message
}

// HandlerFunc is a Fiber handler function.
type HandlerFunc func(*fiber.Ctx) error

// HandlerWithRequest is a generic handler function with request parsing and response.
// T is the type of the parsed request struct.
type HandlerWithRequest[T any] func(*fiber.Ctx, *T) (interface{}, error)

// validate is the global validator instance used for request and response validation.
var validate = validator.New()

// GetValidator returns the global validator instance for validation.
func GetValidator() *validator.Validate {
	return validate
}

// ValidateStruct is a helper that validates any struct using the global validator.
// It is especially convenient to use from your domain models or services:
//
//   type User struct {
//       Email string `json:"email" validate:"required,email"`
//       Age   int    `json:"age" validate:"gte=18"`
//   }
//
//   u := &User{Email: "test@example.com", Age: 20}
//   if err := autofiber.ValidateStruct(u); err != nil {
//       // handle validation error
//   }
//
func ValidateStruct[T any](m *T) error {
	return GetValidator().Struct(m)
}

// applyOptions applies a list of RouteOption to RouteOptions and returns the configured RouteOptions.
func applyOptions(options []RouteOption) *RouteOptions {
	opts := &RouteOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// Infer JWT auth requirement from request schema if not explicitly set.
	// If a schema has a required Authorization header (parse:"header:Authorization" with required),
	// mark the route as requiring JWT auth for consistent behavior/docs.
	if !opts.RequireJWTAuth && opts.RequestSchema != nil {
		if schemaRequiresAuthHeader(opts.RequestSchema) {
			opts.RequireJWTAuth = true
		}
	}

	return opts
}

// schemaRequiresAuthHeader returns true if the schema declares a required Authorization header.
func schemaRequiresAuthHeader(schema interface{}) bool {
	t := reflect.TypeOf(schema)
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	return structHasAuthHeader(t)
}

func structHasAuthHeader(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Recurse into embedded structs (excluding time.Time)
		if f.Anonymous {
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct && ft != reflect.TypeOf(time.Time{}) {
				if structHasAuthHeader(ft) {
					return true
				}
			}
		}

		parseTag := f.Tag.Get("parse")
		validateTag := f.Tag.Get("validate")

		var source, key string
		if parseTag != "" {
			parts := strings.Split(parseTag, ",")
			sourcePart := parts[0]
			sourceKey := strings.SplitN(sourcePart, ":", 2)
			source = sourceKey[0]
			if len(sourceKey) == 2 {
				key = sourceKey[1]
			} else {
				key = f.Name
			}
		}

		required := strings.Contains(parseTag, "required") || strings.Contains(validateTag, "required")

		if strings.EqualFold(source, "header") && strings.EqualFold(key, "authorization") && required {
			return true
		}
	}
	return false
}
