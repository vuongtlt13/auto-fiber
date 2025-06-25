package autofiber

import (
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"
)

// createHandlerWithOptions creates a handler with the given options
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
		return Simple(simpleHandler)
	}
	return handler.(fiber.Handler)
}

func printRoutes(app *fiber.App) {
	fmt.Println("Registered routes:")
	for _, routeList := range app.Stack() {
		for _, r := range routeList {
			fmt.Printf("[%s] %s\n", r.Method, r.Path)
		}
	}
}

// createStructHandler creates a handler for struct-based request schemas
func (af *AutoFiber) createStructHandler(handler interface{}, opts *RouteOptions) fiber.Handler {
	// Try to match handler signature with request schema
	handlerType := reflect.TypeOf(handler)

	// Check if handler is func(c *fiber.Ctx, req *SchemaType) (interface{}, error)
	if handlerType.Kind() == reflect.Func && handlerType.NumIn() == 2 && handlerType.NumOut() == 2 {
		// Create a wrapper that applies middleware and calls handler
		return func(c *fiber.Ctx) error {
			// Apply auto-parse middleware
			parseMiddleware := AutoParseRequest(opts.RequestSchema, nil)
			if err := parseMiddleware(c); err != nil {
				return err
			}

			// Get parsed request
			req := c.Locals("parsed_request")
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

			// Check for error first
			if len(results) > 1 && !results[1].IsNil() {
				return results[1].Interface().(error)
			}

			// Return data as JSON
			data := results[0].Interface()
			return c.JSON(data)
		}
	}

	// Check if handler is func(c *fiber.Ctx, req *SchemaType) error (legacy)
	if handlerType.Kind() == reflect.Func && handlerType.NumIn() == 2 && handlerType.NumOut() == 1 {
		// Create a wrapper that applies middleware and calls handler
		return func(c *fiber.Ctx) error {
			// Apply auto-parse middleware
			parseMiddleware := AutoParseRequest(opts.RequestSchema, nil)
			if err := parseMiddleware(c); err != nil {
				return err
			}

			// Get parsed request
			req := c.Locals("parsed_request")
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

	// Fallback to simple handler
	if simpleHandler, ok := handler.(func(*fiber.Ctx) error); ok {
		return Simple(simpleHandler)
	}

	return handler.(fiber.Handler)
}

// createHandlerWithRequest creates a handler with automatic request parsing and validation
func (af *AutoFiber) createHandlerWithRequest(handler interface{}, requestSchema interface{}, responseSchema interface{}) fiber.Handler {
	handlerType := reflect.TypeOf(handler)
	handlerValue := reflect.ValueOf(handler)

	// Get the first method (assuming it's the handler method)
	if handlerType.NumMethod() == 0 {
		panic("handler must have at least one method")
	}

	method := handlerType.Method(0)
	methodType := method.Type

	// Check if method has the correct signature
	if methodType.NumIn() != 2 || methodType.NumOut() != 2 {
		panic("handler method must have signature: func(*fiber.Ctx, *RequestType) (interface{}, error)")
	}

	// Check first parameter is *fiber.Ctx
	if methodType.In(0) != reflect.TypeOf((*fiber.Ctx)(nil)) {
		panic("first parameter must be *fiber.Ctx")
	}

	// Check second parameter is pointer to request schema
	expectedRequestType := reflect.PtrTo(reflect.TypeOf(requestSchema))
	if methodType.In(1) != expectedRequestType {
		panic("second parameter must be pointer to request schema")
	}

	// Check return types
	if methodType.Out(0) != reflect.TypeOf((*interface{})(nil)).Elem() || methodType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		panic("return types must be (interface{}, error)")
	}

	return func(c *fiber.Ctx) error {
		// Create request instance
		requestType := reflect.TypeOf(requestSchema)
		if requestType.Kind() == reflect.Ptr {
			requestType = requestType.Elem()
		}
		request := reflect.New(requestType).Interface()

		// Parse request body
		if err := c.BodyParser(request); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
		}

		// Validate request
		if err := af.validator.Struct(request); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		}

		// Call handler method
		args := []reflect.Value{
			reflect.ValueOf(c),
			reflect.ValueOf(request),
		}

		results := handlerValue.MethodByName(method.Name).Call(args)

		// Check for error
		if !results[1].IsNil() {
			return results[1].Interface().(error)
		}

		// Return response data
		responseData := results[0].Interface()

		// Use ValidateAndJSON if response schema is provided
		if responseSchema != nil {
			return ValidateAndJSON(c, responseData)
		}

		return c.JSON(responseData)
	}
}

// createHandlerWithRequestAndValidation creates a handler with request parsing, validation, and response validation
func (af *AutoFiber) createHandlerWithRequestAndValidation(handler interface{}, requestSchema interface{}, responseSchema interface{}) fiber.Handler {
	// Create the base handler
	baseHandler := af.createHandlerWithRequest(handler, requestSchema, responseSchema)

	// Wrap with response validation middleware if schema is provided
	if responseSchema != nil {
		return func(c *fiber.Ctx) error {
			// Apply response validation middleware
			validateMiddleware := ValidateResponse(responseSchema, nil)
			if err := validateMiddleware(c); err != nil {
				return err
			}

			// Call the base handler
			return baseHandler(c)
		}
	}

	return baseHandler
}
