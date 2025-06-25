package autofiber

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// New creates a new AutoFiber instance
func New(config ...fiber.Config) *AutoFiber {
	app := fiber.New(config...)
	return &AutoFiber{
		App:           app,
		docsGenerator: NewDocsGenerator(""),
		validator:     validator.New(),
	}
}

// Group creates a new route group
func (af *AutoFiber) Group(prefix string, handlers ...fiber.Handler) *AutoFiberGroup {
	group := af.App.Group(prefix, handlers...)
	return &AutoFiberGroup{
		Group: group.(*fiber.Group),
		app:   af,
	}
}

// Use adds middleware to the app
func (af *AutoFiber) Use(args ...interface{}) fiber.Router {
	return af.App.Use(args...)
}

// Listen starts the server
func (af *AutoFiber) Listen(addr string) error {
	return af.App.Listen(addr)
}

// Test creates a test request
func (af *AutoFiber) Test(req *http.Request, msTimeout ...int) (*http.Response, error) {
	return af.App.Test(req, msTimeout...)
}
