package autofiber_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestGetValidator(t *testing.T) {
	validator := autofiber.GetValidator()
	assert.NotNil(t, validator)
}

func TestValidateResponseData(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type ValidResponse struct {
		ID    int    `json:"id" validate:"required"`
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	app.Get("/test-valid", func(c *fiber.Ctx) (interface{}, error) {
		// Set up response validation
		c.Locals("response_schema", ValidResponse{})
		c.Locals("response_validator", autofiber.GetValidator())

		// Return valid response data
		return ValidResponse{
			ID:    1,
			Name:  "John Doe",
			Email: "john@example.com",
		}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test-valid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateResponseData_Invalid(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type ValidResponse struct {
		ID    int    `json:"id" validate:"required"`
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	app.Get("/test-invalid", func(c *fiber.Ctx) (interface{}, error) {
		// Set up response validation
		c.Locals("response_schema", ValidResponse{})
		c.Locals("response_validator", autofiber.GetValidator())

		// Return invalid response data (missing required fields)
		invalidData := ValidResponse{
			ID: 1,
			// Name and Email are missing
		}

		// Use ValidateAndJSON to actually perform validation
		return autofiber.ValidateAndJSON(c, invalidData), nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test-invalid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestValidateResponseData_WithMap(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type ValidResponse struct {
		ID   int    `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	}

	app.Get("/test-map", func(c *fiber.Ctx) (interface{}, error) {
		// Set up response validation
		c.Locals("response_schema", ValidResponse{})
		c.Locals("response_validator", autofiber.GetValidator())

		// Return map data that should be validated
		return fiber.Map{
			"id":   1,
			"name": "John Doe",
		}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test-map", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateResponseData_WithInvalidMap(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type ValidResponse struct {
		ID   int    `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	}

	app.Get("/test-invalid-map", func(c *fiber.Ctx) (interface{}, error) {
		// Set up response validation
		c.Locals("response_schema", ValidResponse{})
		c.Locals("response_validator", autofiber.GetValidator())

		// Return invalid map data
		invalidData := fiber.Map{
			"id": 1,
			// name is missing
		}

		// Use ValidateAndJSON to actually perform validation
		return autofiber.ValidateAndJSON(c, invalidData), nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test-invalid-map", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestValidateResponseData_NoValidation(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	app.Get("/test-no-validation", func(c *fiber.Ctx) (interface{}, error) {
		// No response validation set up
		return fiber.Map{
			"message": "no validation",
		}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test-no-validation", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateResponseData_WithMapDataAndValidation(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type ValidResponse struct {
		ID   int    `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	}

	app.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		// Set up response validation
		c.Locals("response_schema", ValidResponse{})
		c.Locals("response_validator", autofiber.GetValidator())

		// Test with map data that should be validated
		err := autofiber.ValidateAndJSON(c, fiber.Map{
			"id":   1,
			"name": "test",
		})
		return nil, err
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
