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
//
// Schema metadata (field info, parse tags) is pre-computed here at registration time,
// not on every request.
func AutoParseRequest(schema interface{}, customValidator *validator.Validate) fiber.Handler {
	if customValidator == nil {
		customValidator = GetValidator()
	}

	// Pre-compute schema type and field metadata once at registration time.
	// getOrCacheSchemaMeta also validates parse tag sources and panics on typos.
	schemaType := reflect.TypeOf(schema)
	if schemaType.Kind() == reflect.Ptr {
		schemaType = schemaType.Elem()
	}
	getOrCacheSchemaMeta(schemaType)

	return func(c *fiber.Ctx) error {
		req := reflect.New(schemaType).Interface()

		if err := parseFromMultipleSources(c, req); err != nil {
			return err
		}

		if err := customValidator.Struct(req); err != nil {
			return err
		}

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

	if schema == nil || validatorInstance == nil {
		return c.JSON(data)
	}

	if v, ok := validatorInstance.(*validator.Validate); ok {
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
