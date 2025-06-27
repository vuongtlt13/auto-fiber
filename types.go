// Package autofiber provides core types and configuration for the AutoFiber web framework.
package autofiber

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AutoFiberGroup represents a group of routes with a common prefix and shared middleware.
type AutoFiberGroup struct {
	Group *fiber.Group // Underlying Fiber group
	app   *AutoFiber   // Reference to the parent AutoFiber app
}

// RouteOption is a function that configures route options for an endpoint.
type RouteOption func(*RouteOptions)

// RouteOptions contains configuration for a route, such as schemas, middleware, and metadata.
type RouteOptions struct {
	RequestSchema  interface{}     // Struct for request parsing and validation
	ResponseSchema interface{}     // Struct for response validation and documentation
	Middleware     []fiber.Handler // Middleware handlers for the route
	Description    string          // Description for API documentation
	Tags           []string        // Tags for API documentation
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

// applyOptions applies a list of RouteOption to RouteOptions and returns the configured RouteOptions.
func applyOptions(options []RouteOption) *RouteOptions {
	opts := &RouteOptions{}
	for _, opt := range options {
		opt(opts)
	}
	return opts
}
