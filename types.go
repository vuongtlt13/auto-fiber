package autofiber

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AutoFiber wraps Fiber with auto-parse capabilities
type AutoFiber struct {
	*fiber.App
	docsGenerator *DocsGenerator
	docsInfo      *OpenAPIInfo
	docsServers   []OpenAPIServer
	validator     *validator.Validate
}

// AutoFiberGroup wraps Fiber group with auto-parse capabilities
type AutoFiberGroup struct {
	*fiber.Group
	app *AutoFiber
}

// RouteOptions defines options for route registration
type RouteOptions struct {
	RequestSchema  interface{}
	ResponseSchema interface{}
	Middleware     []fiber.Handler
	Description    string
	Tags           []string
}

// RouteOption is a function that modifies RouteOptions
type RouteOption func(*RouteOptions)

// defaultRouteOptions returns default route options
func defaultRouteOptions() *RouteOptions {
	return &RouteOptions{
		Middleware: []fiber.Handler{},
		Tags:       []string{},
	}
}

// applyOptions applies route options
func applyOptions(options []RouteOption) *RouteOptions {
	opts := defaultRouteOptions()
	for _, option := range options {
		option(opts)
	}
	return opts
}
