// Package autofiber provides middleware functions for automatic request parsing, validation, and response handling.
package autofiber

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AutoParseRequest returns middleware for automatic request parsing and validation.
// The middleware parses request data from multiple sources (body, query, path, headers, cookies, form)
// based on struct tags and validates the parsed data using the provided schema.
// If customValidator is nil, it uses the global validator instance.
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

		// Parse from multiple sources based on struct tags and HTTP method
		if err := parseFromMultipleSources(c, req); err != nil {
			return err
		}

		// Validate the request
		if err := customValidator.Struct(req); err != nil {
			return err
		}

		// Store parsed request in context
		c.Locals("parsed_request", req)
		return nil
	}
}

// ValidateAndJSON validates response data and returns JSON.
// If response validation is configured, it validates the data against the response schema
// before returning the JSON response. If validation fails, it returns an error response.
// If no validation is configured, it simply returns the JSON response.
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
			return &ValidationResponseError{
				Message: "Response validation failed",
				Details: []FieldErrorDetail{{
					Field:   "response",
					Message: err.Error(),
				}},
			}
		}
	}

	return c.JSON(data)
}

// GetParsedRequest retrieves the parsed request from context.
// This function extracts the parsed request data that was stored by AutoParseRequest middleware.
// It returns nil if no parsed request is found or if the type assertion fails.
func GetParsedRequest[T any](c *fiber.Ctx) *T {
	if req, ok := c.Locals("parsed_request").(*T); ok {
		return req
	}
	return nil
}
