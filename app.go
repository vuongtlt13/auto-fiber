// Package autofiber provides a FastAPI-like wrapper for the Fiber web framework.
// It enables automatic request parsing, validation, and OpenAPI/Swagger documentation generation.
package autofiber

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AutoFiberOption is a function type for configuring AutoFiber options
type AutoFiberOption func(*AutoFiber)

// WithOpenAPI sets the OpenAPI info for the documentation (no server info).
func WithOpenAPI(info OpenAPIInfo) AutoFiberOption {
	return func(af *AutoFiber) {
		af.docsGenerator.DocsInfo = &info
	}
}

// AutoFiber is the main application struct for building APIs with automatic parsing, validation, and documentation.
type AutoFiber struct {
	App           *fiber.App
	docsGenerator *DocsGenerator
	validator     *validator.Validate
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
