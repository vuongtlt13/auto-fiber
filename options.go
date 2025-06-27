// Package autofiber provides route configuration options for building APIs with automatic parsing, validation, and documentation.
package autofiber

import (
	"github.com/gofiber/fiber/v2"
)

// WithRequestSchema sets the request schema for auto-parsing.
// The schema should be a struct type that defines the expected request structure.
func WithRequestSchema(schema interface{}) RouteOption {
	return func(opts *RouteOptions) {
		opts.RequestSchema = schema
	}
}

// WithResponseSchema sets the response schema for documentation and validation.
// The schema should be a struct type that defines the expected response structure.
func WithResponseSchema(schema interface{}) RouteOption {
	return func(opts *RouteOptions) {
		opts.ResponseSchema = schema
	}
}

// WithMiddleware adds middleware to the route.
// Multiple middleware can be added and they will be executed in the order provided.
func WithMiddleware(middleware ...fiber.Handler) RouteOption {
	return func(opts *RouteOptions) {
		opts.Middleware = append(opts.Middleware, middleware...)
	}
}

// WithDescription sets the route description for API documentation.
// This description will appear in the generated OpenAPI/Swagger documentation.
func WithDescription(description string) RouteOption {
	return func(opts *RouteOptions) {
		opts.Description = description
	}
}

// WithTags sets the route tags for API documentation.
// Tags help organize and categorize routes in the generated documentation.
func WithTags(tags ...string) RouteOption {
	return func(opts *RouteOptions) {
		opts.Tags = tags
	}
}
