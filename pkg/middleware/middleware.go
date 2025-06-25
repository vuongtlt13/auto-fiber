package middleware

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// Simple wraps a simple handler
func Simple(handler func(*fiber.Ctx) error) fiber.Handler {
	return handler
}

// AutoParseRequest creates middleware for auto-parsing request body
func AutoParseRequest(schema interface{}, customValidator *validator.Validate) fiber.Handler {
	if customValidator == nil {
		customValidator = validate
	}

	return func(c *fiber.Ctx) error {
		// Create a new instance of the schema
		schemaType := reflect.TypeOf(schema)
		if schemaType.Kind() == reflect.Ptr {
			schemaType = schemaType.Elem()
		}

		req := reflect.New(schemaType).Interface()

		// Parse request body
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
		}

		// Validate the request
		if err := customValidator.Struct(req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		}

		// Store parsed request in context
		c.Locals("parsed_request", req)
		return c.Next()
	}
}

// GetParsedRequest retrieves the parsed request from context
func GetParsedRequest[T any](c *fiber.Ctx) *T {
	if req, ok := c.Locals("parsed_request").(*T); ok {
		return req
	}
	return nil
}
