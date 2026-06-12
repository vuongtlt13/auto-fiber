// Package autofiber provides a FastAPI-like wrapper for the Fiber web framework.
// It enables automatic request parsing, validation, and OpenAPI/Swagger documentation generation.
package autofiber

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AutoFiberOption is a function type for configuring AutoFiber options.
type AutoFiberOption func(*AutoFiber)

// WithOpenAPI sets the OpenAPI info for the documentation.
func WithOpenAPI(info OpenAPIInfo) AutoFiberOption {
	return func(af *AutoFiber) {
		af.docsGenerator.DocsInfo = &info
	}
}

// WithErrorHandler sets a custom error handler for request/response validation errors.
// When set, validation errors are passed through fn instead of being returned directly.
// This lets you control the response format for validation failures.
//
// Example:
//
//	app := autofiber.New(fiber.Config{}, autofiber.WithErrorHandler(func(c *fiber.Ctx, err error) error {
//	    return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
//	}))
func WithErrorHandler(fn func(*fiber.Ctx, error) error) AutoFiberOption {
	return func(af *AutoFiber) {
		af.errorHandler = fn
	}
}

// WithValidatorSetup runs fn against the instance's validator immediately after creation,
// allowing custom validation tags to be registered on the instance validator.
//
// Example:
//
//	app := autofiber.New(fiber.Config{}, autofiber.WithValidatorSetup(func(v *validator.Validate) {
//	    v.RegisterValidation("strong_password", validateStrongPassword)
//	}))
func WithValidatorSetup(fn func(*validator.Validate)) AutoFiberOption {
	return func(af *AutoFiber) {
		fn(af.validator)
	}
}

// AutoFiber is the main application struct for building APIs with automatic parsing, validation, and documentation.
type AutoFiber struct {
	App           *fiber.App
	docsGenerator *DocsGenerator
	validator     *validator.Validate
	errorHandler  func(*fiber.Ctx, error) error
}

// New creates a new AutoFiber application instance with custom options.
func New(config fiber.Config, options ...AutoFiberOption) *AutoFiber {
	af := &AutoFiber{
		App:           fiber.New(config),
		docsGenerator: NewDocsGenerator(),
		validator:     validator.New(),
	}
	for _, option := range options {
		option(af)
	}
	return af
}

// RegisterValidator registers a custom validation function on the instance's validator.
// Use this after New() to add validations that should apply to all routes on this instance.
func (af *AutoFiber) RegisterValidator(tag string, fn validator.Func) error {
	return af.validator.RegisterValidation(tag, fn)
}

// Group creates a new route group with the given prefix.
func (af *AutoFiber) Group(prefix string, handlers ...fiber.Handler) *AutoFiberGroup {
	group := af.App.Group(prefix, handlers...)
	return &AutoFiberGroup{
		Group:  group.(*fiber.Group),
		app:    af,
		Prefix: prefix,
	}
}

// Use adds middleware to the app.
func (af *AutoFiber) Use(args ...interface{}) fiber.Router {
	return af.App.Use(args...)
}

// Listen starts the Fiber application on the specified address.
func (af *AutoFiber) Listen(addr string) error {
	return af.App.Listen(addr)
}

// Test creates a test request for the Fiber application.
func (af *AutoFiber) Test(req *http.Request, msTimeout ...int) (*http.Response, error) {
	return af.App.Test(req, msTimeout...)
}

// handleError routes a validation/parse error through the custom error handler when one is set,
// or returns it directly for fiber's error handler otherwise.
func (af *AutoFiber) handleError(c *fiber.Ctx, err error) error {
	if af.errorHandler != nil {
		return af.errorHandler(c, err)
	}
	return err
}
