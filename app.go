// Package autofiber provides a FastAPI-like wrapper for the Fiber web framework.
// It enables automatic request parsing, validation, and OpenAPI/Swagger documentation generation.
package autofiber

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AutoFiber is the main application struct for building APIs with automatic parsing, validation, and documentation.
type AutoFiber struct {
	App           *fiber.App
	docsGenerator *DocsGenerator
	validator     *validator.Validate
	docsInfo      *OpenAPIInfo
	docsServers   []OpenAPIServer
}

// New creates a new AutoFiber application instance with default configuration.
func New(config ...fiber.Config) *AutoFiber {
	app := fiber.New(config...)
	return &AutoFiber{
		App:           app,
		docsGenerator: NewDocsGenerator(""),
		validator:     validator.New(),
	}
}

// Group creates a new route group with the given prefix.
func (af *AutoFiber) Group(prefix string, handlers ...fiber.Handler) *AutoFiberGroup {
	group := af.App.Group(prefix, handlers...)
	return &AutoFiberGroup{
		Group: group.(*fiber.Group),
		app:   af,
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
