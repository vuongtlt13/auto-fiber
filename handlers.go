// Package autofiber provides handler creation utilities for automatic request parsing, validation, and response handling.
package autofiber

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// createHandlerWithOptions returns a handler with the given options.
// Only two handler signatures are supported:
// 1. func(*fiber.Ctx) (interface{}, error) -- for endpoints without request schema
// 2. func(*fiber.Ctx, req *T) (interface{}, error) -- for endpoints with request schema
// All other signatures will panic.
func (af *AutoFiber) createHandlerWithOptions(handler interface{}, opts *RouteOptions) fiber.Handler {
	handlerType := reflect.TypeOf(handler)

	if handlerType.Kind() != reflect.Func {
		panic("Handler must be a function")
	}

	if opts.RequestSchema == nil {
		// Only allow func(*fiber.Ctx) (interface{}, error)
		if handlerType.NumIn() == 1 && handlerType.NumOut() == 2 {
			return func(c *fiber.Ctx) error {
				results := reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(c)})
				data := results[0].Interface()
				err, _ := results[1].Interface().(error)
				if err != nil {
					return err
				}
				// Validate response if schema is set
				if opts.ResponseSchema != nil {
					c.Locals("response_schema", opts.ResponseSchema)
					c.Locals("response_validator", af.validator)
					return ValidateAndJSON(c, data)
				}
				return c.JSON(data)
			}
		}
		panic("Handler must be func(*fiber.Ctx) (interface{}, error) when no request schema is provided")
	}

	// With request schema: only allow func(*fiber.Ctx, req *T) (interface{}, error)
	if handlerType.NumIn() == 2 && handlerType.NumOut() == 2 {
		return func(c *fiber.Ctx) error {
			// Parse and validate request
			parseMiddleware := AutoParseRequest(opts.RequestSchema, af.validator)
			if err := parseMiddleware(c); err != nil {
				// Handle parse errors
				if parseErr, ok := err.(*ParseError); ok {
					return &ValidationRequestError{
						Message: "Invalid request",
						Details: []FieldErrorDetail{{
							Field:   parseErr.Field,
							Message: parseErr.Message,
							Tag:     "parse",
						}},
					}
				}

				// Handle validation errors (from validator)
				if validationErrs, ok := err.(validator.ValidationErrors); ok {
					var details []FieldErrorDetail
					for _, verr := range validationErrs {
						details = append(details, FieldErrorDetail{
							Field:   verr.Namespace(),
							Message: verr.Error(),
							Tag:     verr.Tag(),
						})
					}
					return &ValidationRequestError{
						Message: "Validation failed",
						Details: details,
					}
				}
				// Unknown error
				return &ValidationRequestError{
					Message: err.Error(),
					Details: nil,
				}
			}
			req := c.Locals("parsed_request")
			if req == nil {
				return &ValidationRequestError{Message: "Invalid request"}
			}
			results := reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(req)})
			data := results[0].Interface()
			err, _ := results[1].Interface().(error)
			if err != nil {
				return err
			}
			// Validate response if schema is set
			if opts.ResponseSchema != nil {
				c.Locals("response_schema", opts.ResponseSchema)
				c.Locals("response_validator", af.validator)
				return ValidateAndJSON(c, data)
			}
			return c.JSON(data)
		}
	}

	panic("Handler must be func(*fiber.Ctx) (interface{}, error) or func(*fiber.Ctx, req *T) (interface{}, error)")
}
