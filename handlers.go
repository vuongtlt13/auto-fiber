// Package autofiber provides handler creation utilities for automatic request parsing, validation, and response handling.
package autofiber

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// createHandlerWithOptions returns a handler with the given options.
// Supported handler signatures:
// 1. func(*fiber.Ctx) (interface{}, error) or (*ResponseSchema, error) -- for endpoints without request schema
// 2. func(*fiber.Ctx, req *T) (interface{}, error) or (*ResponseSchema, error) -- for endpoints with request schema
// When WithResponseSchema is provided, you can return the concrete schema type instead of interface{}.
// All other signatures will panic.
func (af *AutoFiber) createHandlerWithOptions(handler interface{}, opts *RouteOptions) fiber.Handler {
	handlerType := reflect.TypeOf(handler)

	if handlerType.Kind() != reflect.Func {
		panic("Handler must be a function")
	}

	if opts.RequestSchema == nil {
		// Allow func(*fiber.Ctx) (interface{}, error) or (*ResponseSchema, error)
		if handlerType.NumIn() == 1 && handlerType.NumOut() == 2 {
			return func(c *fiber.Ctx) error {
				// Enforce Authorization header when JWT auth is required (no request schema to validate it)
				if opts.RequireJWTAuth && c.Get("Authorization") == "" {
					return fiber.NewError(fiber.StatusUnauthorized, "Missing Authorization header")
				}

				results := reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(c)})
				data := results[0].Interface()
				err, _ := results[1].Interface().(error)
				if err != nil {
					return err
				}

				// If handler returned a FileResponse, send file directly (no JSON / validation).
				if fr, ok := data.(FileResponse); ok {
					return fr.SendFileResponse(c)
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
		panic("Handler must be func(*fiber.Ctx) (interface{}, error) or (*ResponseSchema, error) when no request schema is provided")
	}

	// With request schema: allow func(*fiber.Ctx, req *T) (interface{}, error) or (*ResponseSchema, error)
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

				// If JWT auth is required and Authorization header is missing -> 401
				if opts.RequireJWTAuth && c.Get("Authorization") == "" {
					return fiber.NewError(fiber.StatusUnauthorized, "Missing Authorization header")
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

			// Enforce Authorization header when JWT auth is required.
			// Even though RequestSchema could already require it, this guarantees presence.
			if opts.RequireJWTAuth && c.Get("Authorization") == "" {
				return fiber.NewError(fiber.StatusUnauthorized, "Missing Authorization header")
			}

			results := reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(c), reflect.ValueOf(req)})
			data := results[0].Interface()
			err, _ := results[1].Interface().(error)
			if err != nil {
				return err
			}

			// If handler returned a FileResponse, send file directly (no JSON / validation).
			if fr, ok := data.(FileResponse); ok {
				return fr.SendFileResponse(c)
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

	panic("Handler must be func(*fiber.Ctx) (interface{}, error) or (*ResponseSchema, error), or func(*fiber.Ctx, req *T) (interface{}, error) or (*ResponseSchema, error)")
}
