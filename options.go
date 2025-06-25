package autofiber

import (
	"github.com/gofiber/fiber/v2"
)

// WithRequestSchema sets the request schema for auto-parsing
func WithRequestSchema(schema interface{}) RouteOption {
	return func(opts *RouteOptions) {
		opts.RequestSchema = schema
	}
}

// WithResponseSchema sets the response schema for documentation
func WithResponseSchema(schema interface{}) RouteOption {
	return func(opts *RouteOptions) {
		opts.ResponseSchema = schema
	}
}

// WithMiddleware adds middleware to the route
func WithMiddleware(middleware ...fiber.Handler) RouteOption {
	return func(opts *RouteOptions) {
		opts.Middleware = append(opts.Middleware, middleware...)
	}
}

// WithDescription sets the route description
func WithDescription(description string) RouteOption {
	return func(opts *RouteOptions) {
		opts.Description = description
	}
}

// WithTags sets the route tags
func WithTags(tags ...string) RouteOption {
	return func(opts *RouteOptions) {
		opts.Tags = tags
	}
}
