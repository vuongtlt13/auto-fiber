package autofiber

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Simple wraps a simple handler
func Simple(handler func(*fiber.Ctx) error) fiber.Handler {
	return handler
}

// AutoParseRequest creates middleware for auto-parsing request from multiple sources
func AutoParseRequest(schema interface{}, customValidator *validator.Validate) fiber.Handler {
	if customValidator == nil {
		customValidator = GetValidator()
	}

	return func(c *fiber.Ctx) error {
		// Create a new instance of the schema
		schemaType := reflect.TypeOf(schema)
		if schemaType.Kind() == reflect.Ptr {
			schemaType = schemaType.Elem()
		}

		req := reflect.New(schemaType).Interface()

		// Parse from different sources based on struct tags and HTTP method
		if err := parseFromMultipleSources(c, req); err != nil {
			if _, ok := err.(*ParseError); ok {
				return err
				//return c.Status(400).JSON(fiber.Map{
				//	"error":   "Invalid request",
				//	"details": err.Error(),
				//})
			}
			// fallback: unknown error
			return err
			//return c.Status(400).JSON(fiber.Map{
			//	"error":   "Invalid request",
			//	"details": err.Error(),
			//})
		}

		// Validate the request
		if err := customValidator.Struct(req); err != nil {
			return err
			//return c.Status(422).JSON(fiber.Map{
			//	"error":   "Validation failed",
			//	"details": err.Error(),
			//})
		}

		// Store parsed request in context
		c.Locals("parsed_request", req)
		return nil
	}
}

// ValidateResponse creates middleware for response validation
func ValidateResponse(schema interface{}, customValidator *validator.Validate) fiber.Handler {
	if customValidator == nil {
		customValidator = GetValidator()
	}

	return func(c *fiber.Ctx) error {
		// Store validation info in context
		c.Locals("response_schema", schema)
		c.Locals("response_validator", customValidator)
		return nil
	}
}

// ValidateAndJSON validates response data and returns JSON
func ValidateAndJSON(c *fiber.Ctx, data interface{}) error {
	schema := c.Locals("response_schema")
	validatorInstance := c.Locals("response_validator")

	// If no validation is set up, just return JSON
	if schema == nil || validatorInstance == nil {
		return c.JSON(data)
	}

	// Type assert validator
	if v, ok := validatorInstance.(*validator.Validate); ok {
		// Validate response data
		if err := validateResponseData(data, schema, v); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Response validation failed",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(data)
}

// GetParsedRequest retrieves the parsed request from context
func GetParsedRequest[T any](c *fiber.Ctx) *T {
	if req, ok := c.Locals("parsed_request").(*T); ok {
		return req
	}
	return nil
}
