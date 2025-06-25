package autofiber

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AutoFiber represents the main AutoFiber application
type AutoFiber struct {
	App           *fiber.App
	docsGenerator *DocsGenerator
	validator     *validator.Validate
	docsInfo      *OpenAPIInfo
	docsServers   []OpenAPIServer
}

// AutoFiberGroup represents an AutoFiber route group
type AutoFiberGroup struct {
	Group *fiber.Group
	app   *AutoFiber
}

// RouteOption is a function that configures route options
type RouteOption func(*RouteOptions)

// RouteOptions contains configuration for a route
type RouteOptions struct {
	RequestSchema  interface{}
	ResponseSchema interface{}
	Middleware     []fiber.Handler
	Description    string
	Tags           []string
}

// ParseSource defines where a field should be parsed from
type ParseSource string

const (
	Body   ParseSource = "body"
	Query  ParseSource = "query"
	Path   ParseSource = "path"
	Header ParseSource = "header"
	Cookie ParseSource = "cookie"
	Form   ParseSource = "form"
	Auto   ParseSource = "auto" // Smart parsing based on HTTP method
)

// FieldInfo contains parsing information for a field
type FieldInfo struct {
	Source      ParseSource
	Key         string // Custom key name (e.g., "user_id" for "UserId")
	Required    bool
	Default     interface{}
	Description string
}

// ParseError represents a parsing error
type ParseError struct {
	Field   string
	Source  string
	Message string
}

func (e *ParseError) Error() string {
	return e.Field + " (" + e.Source + "): " + e.Message
}

// Handler types for better type safety
type HandlerFunc func(*fiber.Ctx) error

type HandlerWithRequest[T any] func(*fiber.Ctx, *T) (interface{}, error)

// Global validator instance
var validate = validator.New()

// GetValidator returns the global validator instance
func GetValidator() *validator.Validate {
	return validate
}

// applyOptions applies a list of RouteOption to RouteOptions
func applyOptions(options []RouteOption) *RouteOptions {
	opts := &RouteOptions{}
	for _, opt := range options {
		opt(opts)
	}
	return opts
}
