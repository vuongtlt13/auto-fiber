package autofiber

import (
	"reflect"

	"autofiber/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

// createHandlerWithOptions creates a handler with auto-parse middleware based on options
func (af *AutoFiber) createHandlerWithOptions(handler interface{}, opts *RouteOptions) fiber.Handler {
	// Add route to docs generator
	af.docsGenerator.AddRoute("", "", handler, opts)

	// If no request schema, use simple handler
	if opts.RequestSchema == nil {
		if simpleHandler, ok := handler.(func(*fiber.Ctx) error); ok {
			return middleware.Simple(simpleHandler)
		}
		// Fallback to direct handler
		return handler.(fiber.Handler)
	}

	// Create auto-parse handler based on request schema type
	return af.createAutoParseHandler(handler, opts)
}

// createAutoParseHandler creates an auto-parse handler based on the request schema
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
		return middleware.Simple(simpleHandler)
	}
	return handler.(fiber.Handler)
}

// createStructHandler creates a handler for struct-based request schemas
func (af *AutoFiber) createStructHandler(handler interface{}, opts *RouteOptions) fiber.Handler {
	// Try to match handler signature with request schema
	handlerType := reflect.TypeOf(handler)

	// Check if handler is func(c *fiber.Ctx, req *SchemaType) error
	if handlerType.Kind() == reflect.Func && handlerType.NumIn() == 2 {
		// Create a wrapper that applies middleware and calls handler
		return func(c *fiber.Ctx) error {
			// Apply auto-parse middleware
			parseMiddleware := middleware.AutoParseRequest(opts.RequestSchema, nil)
			if err := parseMiddleware(c); err != nil {
				return err
			}

			// Get parsed request
			req := middleware.GetParsedRequest[interface{}](c)
			if req == nil {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
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

	// Check if handler is func(c *fiber.Ctx, req *SchemaType) (interface{}, error)
	if handlerType.Kind() == reflect.Func && handlerType.NumIn() == 2 && handlerType.NumOut() == 2 {
		// Create a wrapper that applies middleware and calls handler
		return func(c *fiber.Ctx) error {
			// Apply auto-parse middleware
			parseMiddleware := middleware.AutoParseRequest(opts.RequestSchema, nil)
			if err := parseMiddleware(c); err != nil {
				return err
			}

			// Get parsed request
			req := middleware.GetParsedRequest[interface{}](c)
			if req == nil {
				return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
			}

			// Call the handler using reflection
			handlerValue := reflect.ValueOf(handler)
			args := []reflect.Value{
				reflect.ValueOf(c),
				reflect.ValueOf(req),
			}
			results := handlerValue.Call(args)

			if len(results) > 1 && !results[1].IsNil() {
				return results[1].Interface().(error)
			}

			// Return data directly like normal Fiber
			data := results[0].Interface()
			return c.JSON(data)
		}
	}

	// Fallback to simple handler
	if simpleHandler, ok := handler.(func(*fiber.Ctx) error); ok {
		return middleware.Simple(simpleHandler)
	}

	return handler.(fiber.Handler)
}
