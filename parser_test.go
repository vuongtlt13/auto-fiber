package autofiber_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestParseError_Error(t *testing.T) {
	err := &autofiber.ParseError{
		Field:   "age",
		Source:  "query",
		Message: "invalid",
	}
	assert.Equal(t, "age (query): invalid", err.Error())
}

func TestParseSource_Constants(t *testing.T) {
	assert.Equal(t, autofiber.ParseSource("body"), autofiber.Body)
	assert.Equal(t, autofiber.ParseSource("query"), autofiber.Query)
	assert.Equal(t, autofiber.ParseSource("path"), autofiber.Path)
	assert.Equal(t, autofiber.ParseSource("header"), autofiber.Header)
	assert.Equal(t, autofiber.ParseSource("cookie"), autofiber.Cookie)
	assert.Equal(t, autofiber.ParseSource("form"), autofiber.Form)
	assert.Equal(t, autofiber.ParseSource("auto"), autofiber.Auto)
}

func TestParseFromMultipleSources_EdgeCases(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type TestRequest struct {
		ID       int    `parse:"path:id"`
		Name     string `parse:"query:name"`
		Token    string `parse:"header:Authorization"`
		Email    string `json:"email"`
		Age      int    `json:"age"`
		IsActive bool   `json:"is_active"`
	}

	app.Post("/users/:id", func(c *fiber.Ctx, req *TestRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(&TestRequest{}))

	// Test with empty body but valid path and query
	req := httptest.NewRequest(http.MethodPost, "/users/123?name=John", nil)
	req.Header.Set("Authorization", "Bearer token123")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with invalid path parameter
	req = httptest.NewRequest(http.MethodPost, "/users/invalid?name=John", nil)
	req.Header.Set("Authorization", "Bearer token123")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestParseFieldFromSource_ComplexTypes(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type ComplexRequest struct {
		IDs      []int                  `parse:"query:ids"`
		Names    []string               `parse:"query:names"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	app.Post("/complex", func(c *fiber.Ctx, req *ComplexRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(&ComplexRequest{}))

	// Test with array query parameters
	req := httptest.NewRequest(http.MethodPost, "/complex?ids=1,2,3&names=John,Jane", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = httptest.NewRecorder().Result().Body
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Should handle array parsing gracefully
	assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestSetFieldValue_EdgeCases(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type EdgeCaseRequest struct {
		IntField    int     `parse:"query:int"`
		FloatField  float64 `parse:"query:float"`
		BoolField   bool    `parse:"query:bool"`
		StringField string  `parse:"query:string"`
	}

	app.Get("/edge", func(c *fiber.Ctx, req *EdgeCaseRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(&EdgeCaseRequest{}))

	// Test with various data types
	req := httptest.NewRequest(http.MethodGet, "/edge?int=42&float=3.14&bool=true&string=test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with invalid number
	req = httptest.NewRequest(http.MethodGet, "/edge?int=invalid&float=3.14&bool=true&string=test", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetSmartSource(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type SmartRequest struct {
		ID   int    `parse:"auto:id"`
		Name string `parse:"auto:name"`
	}

	// Test GET request (should parse from query)
	app.Get("/smart/:id", func(c *fiber.Ctx, req *SmartRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(&SmartRequest{}))

	req := httptest.NewRequest(http.MethodGet, "/smart/123?name=John", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test POST request (should parse from body)
	app.Post("/smart/:id", func(c *fiber.Ctx, req *SmartRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(&SmartRequest{}))

	req = httptest.NewRequest(http.MethodPost, "/smart/123", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = httptest.NewRecorder().Result().Body
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestConvertDefaultValue(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type DefaultRequest struct {
		IntField    int     `parse:"query:int" default:"42"`
		FloatField  float64 `parse:"query:float" default:"3.14"`
		StringField string  `parse:"query:string" default:"default"`
		BoolField   bool    `parse:"query:bool" default:"true"`
	}

	app.Get("/defaults", func(c *fiber.Ctx, req *DefaultRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(&DefaultRequest{}))

	// Test with missing fields (should use defaults)
	req := httptest.NewRequest(http.MethodGet, "/defaults", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with provided values
	req = httptest.NewRequest(http.MethodGet, "/defaults?int=100&float=2.5&string=custom&bool=false", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestParseFloat(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type FloatRequest struct {
		Price float64 `parse:"query:price"`
		Rate  float64 `parse:"query:rate"`
	}

	app.Get("/float", func(c *fiber.Ctx, req *FloatRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(&FloatRequest{}))

	// Test with valid float values
	req := httptest.NewRequest(http.MethodGet, "/float?price=19.99&rate=0.05", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test with invalid float
	req = httptest.NewRequest(http.MethodGet, "/float?price=invalid&rate=0.05", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
