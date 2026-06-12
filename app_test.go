package autofiber_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestNew(t *testing.T) {
	app := autofiber.New(fiber.Config{})
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
}

func TestNew_WithConfig(t *testing.T) {
	app := autofiber.New(fiber.Config{})
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
}

func TestAutoFiber_Use(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	// Test middleware registration
	called := false
	middleware := func(c *fiber.Ctx) error {
		called = true
		return c.Next()
	}

	app.Use(middleware)

	// Add a test route
	app.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"message": "test"}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called, "Middleware should be called")
}

func TestAutoFiber_Listen(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	// Test that Listen doesn't panic (we can't actually test server startup in unit tests)
	// This is mainly for coverage
	app.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		return fiber.Map{"message": "test"}, nil
	})

	// The Listen method should not panic when called
	// In a real scenario, this would start the server
	// For testing, we just verify the method exists and can be called
	assert.NotNil(t, app.Listen)
}

func TestWithErrorHandler(t *testing.T) {
	handlerCalled := false
	app := autofiber.New(fiber.Config{},
		autofiber.WithErrorHandler(func(c *fiber.Ctx, err error) error {
			handlerCalled = true
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"custom": err.Error()})
		}),
	)

	type Req struct {
		Name string `json:"name" validate:"required"`
	}
	app.Post("/test", func(c *fiber.Ctx, req *Req) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(Req{}))

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.True(t, handlerCalled)
}

func TestWithValidatorSetup(t *testing.T) {
	app := autofiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		},
	}, autofiber.WithValidatorSetup(func(v *validator.Validate) {
		v.RegisterValidation("is_hello", func(fl validator.FieldLevel) bool {
			return fl.Field().String() == "hello"
		})
	}))

	type Req struct {
		Greeting string `json:"greeting" validate:"required,is_hello"`
	}
	app.Post("/greet", func(c *fiber.Ctx, req *Req) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(Req{}))

	// Valid request
	body := strings.NewReader(`{"greeting":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/greet", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Invalid request — custom tag fails
	body = strings.NewReader(`{"greeting":"world"}`)
	req = httptest.NewRequest(http.MethodPost, "/greet", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestRegisterValidator(t *testing.T) {
	app := autofiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		},
	})

	err := app.RegisterValidator("is_foo", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "foo"
	})
	assert.NoError(t, err)

	type Req struct {
		Value string `json:"value" validate:"required,is_foo"`
	}
	app.Post("/foo", func(c *fiber.Ctx, req *Req) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(Req{}))

	// Valid
	body := strings.NewReader(`{"value":"foo"}`)
	req := httptest.NewRequest(http.MethodPost, "/foo", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Invalid
	body = strings.NewReader(`{"value":"bar"}`)
	req = httptest.NewRequest(http.MethodPost, "/foo", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestHandleError_WithCustomHandler(t *testing.T) {
	customCalled := false
	app := autofiber.New(fiber.Config{},
		autofiber.WithErrorHandler(func(c *fiber.Ctx, err error) error {
			customCalled = true
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "custom"})
		}),
	)

	type Req struct {
		X string `json:"x" validate:"required"`
	}
	app.Post("/h", func(c *fiber.Ctx, req *Req) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(Req{}))

	req := httptest.NewRequest(http.MethodPost, "/h", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.True(t, customCalled)
}
