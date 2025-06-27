package autofiber

import (
	"reflect"

	"github.com/gofiber/fiber/v2"
)

// createHandlerWithOptions returns a handler with the given options
func (af *AutoFiber) createHandlerWithOptions(handler interface{}, opts *RouteOptions) fiber.Handler {
	if opts.RequestSchema == nil {
		if simpleHandler, ok := handler.(func(*fiber.Ctx) error); ok {
			return Simple(simpleHandler)
		}
		// Fallback to direct handler
		return handler.(fiber.Handler)
	}

	// Create auto-parse handler based on request schema type
	return af.createAutoParseHandler(handler, opts)
}

// createAutoParseHandler returns an auto-parse handler based on the request schema
func (af *AutoFiber) createAutoParseHandler(handler interface{}, opts *RouteOptions) fiber.Handler {
	reqType := reflect.TypeOf(opts.RequestSchema)

	// Create a generic handler based on the request type
	switch reqType.Kind() {
	case reflect.Struct:
		return af.createStructHandler(handler, opts)
	case reflect.Ptr:
		if reqType.Elem().Kind() == reflect.Struct {
			return af.createStructHandler(handler, opts)
		}
	}

	// Fallback to simple handler
	if simpleHandler, ok := handler.(func(*fiber.Ctx) error); ok {
		return Simple(simpleHandler)
	}
	return handler.(fiber.Handler)
}

// createStructHandler returns a handler for struct-based request schemas
func (af *AutoFiber) createStructHandler(handler interface{}, opts *RouteOptions) fiber.Handler {
	// Try to match handler signature with request schema
	handlerType := reflect.TypeOf(handler)

	// Handler: func(c *fiber.Ctx, req *SchemaType) (interface{}, error)
	if handlerType.Kind() == reflect.Func && handlerType.NumIn() == 2 && handlerType.NumOut() == 2 {
		return func(c *fiber.Ctx) error {
			// Set up response validation if schema is provided
			if opts.ResponseSchema != nil {
				c.Locals("response_schema", opts.ResponseSchema)
				c.Locals("response_validator", af.validator)
			}

			// Apply auto-parse middleware
			parseMiddleware := AutoParseRequest(opts.RequestSchema, af.validator)
			if err := parseMiddleware(c); err != nil {
				// Handle parse errors
				if parseErr, ok := err.(*ParseError); ok {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   "Invalid request",
						"details": parseErr.Error(),
					})
				}
				// Handle validation errors
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"error":   "Validation failed",
					"details": err.Error(),
				})
			}

			// Get parsed request
			req := c.Locals("parsed_request")
			if req == nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
			}

			// Call the handler using reflection
			handlerValue := reflect.ValueOf(handler)
			args := []reflect.Value{
				reflect.ValueOf(c),
				reflect.ValueOf(req),
			}
			results := handlerValue.Call(args)

			// Check for error first
			if len(results) > 1 && !results[1].IsNil() {
				return results[1].Interface().(error)
			}

			// Return data with validation if response schema is set
			data := results[0].Interface()
			return ValidateAndJSON(c, data)
		}
	}

	// Handler: func(c *fiber.Ctx, req *SchemaType) error (legacy)
	if handlerType.Kind() == reflect.Func && handlerType.NumIn() == 2 && handlerType.NumOut() == 1 {
		return func(c *fiber.Ctx) error {
			// Apply auto-parse middleware
			parseMiddleware := AutoParseRequest(opts.RequestSchema, af.validator)
			if err := parseMiddleware(c); err != nil {
				// Handle parse errors
				if parseErr, ok := err.(*ParseError); ok {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   "Invalid request",
						"details": parseErr.Error(),
					})
				}
				// Handle validation errors
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"error":   "Validation failed",
					"details": err.Error(),
				})
			}

			// Get parsed request
			req := c.Locals("parsed_request")
			if req == nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
			}

			// Call the handler using reflection
			handlerValue := reflect.ValueOf(handler)
			args := []reflect.Value{
				reflect.ValueOf(c),
				reflect.ValueOf(req),
			}
			results := handlerValue.Call(args)

			if len(results) > 0 && !results[0].IsNil() {
				return results[0].Interface().(error)
			}
			return nil
		}
	}

	// Fallback to simple handler
	if simpleHandler, ok := handler.(func(*fiber.Ctx) error); ok {
		return Simple(simpleHandler)
	}

	return handler.(fiber.Handler)
}
