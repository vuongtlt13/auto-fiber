package autofiber_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	app := newTestApp()

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
	app := newTestApp()

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
	app := newTestApp()

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
	app := newTestApp()

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
	app := newTestApp()

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
	app := newTestApp()

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

func TestParseEmbeddedStructs_Request(t *testing.T) {
	app := newTestApp()

	type UserBase struct {
		Email    string `json:"email" validate:"required,email"`
		FullName string `json:"fullName" validate:"required"`
	}
	type UserInfo struct {
		UserBase
		IsActive bool `json:"isActive"`
	}
	type UserRequest struct {
		UserInfo
		Age int `json:"age"`
	}

	var parsed *UserRequest
	app.Post("/embedded", func(c *fiber.Ctx, req *UserRequest) (interface{}, error) {
		parsed = req
		return req, nil
	}, autofiber.WithRequestSchema(&UserRequest{}))

	// Test parsing from JSON body (all embedded fields present)
	jsonBody := `{"email":"test@example.com","fullName":"Test User","isActive":true,"age":30}`
	req := httptest.NewRequest(http.MethodPost, "/embedded", bytes.NewReader([]byte(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if assert.NotNil(t, parsed) {
		assert.Equal(t, "test@example.com", parsed.Email)
		assert.Equal(t, "Test User", parsed.FullName)
		assert.Equal(t, true, parsed.IsActive)
		assert.Equal(t, 30, parsed.Age)
	}

	// Test parsing from query string (if using parse tag)
	type QueryUserBase struct {
		Email string `parse:"query:email"`
	}
	type QueryUserRequest struct {
		QueryUserBase
		Age int `parse:"query:age"`
	}
	var parsedQuery QueryUserRequest
	app.Get("/embedded-query", func(c *fiber.Ctx, req *QueryUserRequest) (interface{}, error) {
		parsedQuery = *req
		return req, nil
	}, autofiber.WithRequestSchema(&QueryUserRequest{}))

	req = httptest.NewRequest(http.MethodGet, "/embedded-query?email=abc@xyz.com&age=22", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "abc@xyz.com", parsedQuery.Email)
	assert.Equal(t, 22, parsedQuery.Age)
}

func TestParseEmbeddedPointerStructs_Request(t *testing.T) {
	app := newTestApp()

	type UserBase struct {
		Email    string `json:"email"`
		FullName string `json:"fullName"`
	}
	type UserInfo struct {
		*UserBase
		IsActive bool `json:"isActive"`
	}
	type UserRequest struct {
		*UserInfo
		Age int `json:"age"`
	}

	var parsed *UserRequest
	app.Post("/embedded-pointer", func(c *fiber.Ctx, req *UserRequest) (interface{}, error) {
		parsed = req
		return req, nil
	}, autofiber.WithRequestSchema(&UserRequest{}))

	// Test parsing from JSON body with all fields present
	jsonBody := `{"email":"ptr@example.com","fullName":"Pointer User","isActive":true,"age":25}`
	req := httptest.NewRequest(http.MethodPost, "/embedded-pointer", bytes.NewReader([]byte(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if assert.NotNil(t, parsed) && assert.NotNil(t, parsed.UserInfo) && assert.NotNil(t, parsed.UserInfo.UserBase) {
		assert.Equal(t, "ptr@example.com", parsed.UserInfo.UserBase.Email)
		assert.Equal(t, "Pointer User", parsed.UserInfo.UserBase.FullName)
		assert.Equal(t, true, parsed.UserInfo.IsActive)
		assert.Equal(t, 25, parsed.Age)
	}

	// Test parsing from query string (should also parse into embedded pointer fields)
	parsed = nil
	app.Get("/embedded-pointer-query", func(c *fiber.Ctx, req *UserRequest) (interface{}, error) {
		parsed = req
		return req, nil
	}, autofiber.WithRequestSchema(&UserRequest{}))

	req = httptest.NewRequest(http.MethodGet, "/embedded-pointer-query?email=abc@xyz.com&fullName=PointerQuery&isActive=true&age=22", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if assert.NotNil(t, parsed) && assert.NotNil(t, parsed.UserInfo) && assert.NotNil(t, parsed.UserInfo.UserBase) {
		assert.Equal(t, "abc@xyz.com", parsed.UserInfo.UserBase.Email)
		assert.Equal(t, "PointerQuery", parsed.UserInfo.UserBase.FullName)
		assert.Equal(t, true, parsed.UserInfo.IsActive)
		assert.Equal(t, 22, parsed.Age)
	}

	// Test parsing with missing embedded pointer (should still allocate and parse zero values)
	parsed = nil
	jsonBody = `{"isActive":false,"age":10}`
	req = httptest.NewRequest(http.MethodPost, "/embedded-pointer", bytes.NewReader([]byte(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if assert.NotNil(t, parsed) && assert.NotNil(t, parsed.UserInfo) && assert.NotNil(t, parsed.UserInfo.UserBase) {
		assert.Equal(t, "", parsed.UserInfo.UserBase.Email)
		assert.Equal(t, "", parsed.UserInfo.UserBase.FullName)
		assert.Equal(t, false, parsed.UserInfo.IsActive)
		assert.Equal(t, 10, parsed.Age)
	}
}

func TestParseEmbeddedPointerStructs_WithValidation(t *testing.T) {
	app := newTestApp()

	type UserBase struct {
		Email    string `json:"email" validate:"required,email"`
		FullName string `json:"fullName" validate:"required"`
	}
	type UserInfo struct {
		*UserBase
		IsActive bool `json:"isActive"`
	}
	type UserRequest struct {
		*UserInfo
		Age int `json:"age"`
	}

	app.Post("/embedded-pointer-validate", func(c *fiber.Ctx, req *UserRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(&UserRequest{}))

	// Case 1: missing required fields in pointer embedded struct should trigger validation error
	jsonBody := `{"isActive":true,"age":20}`
	req := httptest.NewRequest(http.MethodPost, "/embedded-pointer-validate", bytes.NewReader([]byte(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	var respBody map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Validation failed", respBody["error"])

	// Check that details is an array of objects and contains fields for Email and FullName
	details, ok := respBody["details"].([]interface{})
	assert.True(t, ok, "details should be an array")
	foundEmail := false
	foundFullName := false
	for _, d := range details {
		if m, ok := d.(map[string]interface{}); ok {
			if f, ok := m["field"].(string); ok {
				if strings.Contains(f, "Email") {
					foundEmail = true
				}
				if strings.Contains(f, "FullName") {
					foundFullName = true
				}
			}
		}
	}
	assert.True(t, foundEmail, "details should mention Email")
	assert.True(t, foundFullName, "details should mention FullName")

	// Case 2: valid request, should pass validation
	jsonBody = `{"email":"valid@example.com","fullName":"Valid User","isActive":true,"age":30}`
	req = httptest.NewRequest(http.MethodPost, "/embedded-pointer-validate", bytes.NewReader([]byte(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestResponseValidation_NoRequestSchema(t *testing.T) {
	app := newTestApp()

	type SimpleResponse struct {
		ID    int    `json:"id" validate:"required"`
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	// Handler returns invalid response (missing required fields)
	app.Get("/no-req-schema-invalid", func(c *fiber.Ctx) (interface{}, error) {
		return SimpleResponse{ID: 0, Name: "", Email: "not-an-email"}, nil
	}, autofiber.WithResponseSchema(SimpleResponse{}))

	req := httptest.NewRequest(http.MethodGet, "/no-req-schema-invalid", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Response validation failed", respBody["error"])
	// Should mention 'response' in details
	details, ok := respBody["details"].([]interface{})
	assert.True(t, ok, "details should be an array")
	foundResponse := false
	for _, d := range details {
		if m, ok := d.(map[string]interface{}); ok {
			if f, ok := m["field"].(string); ok {
				if strings.Contains(f, "response") {
					foundResponse = true
				}
			}
		}
	}
	assert.True(t, foundResponse, "details should mention response")

	// Handler returns valid response
	app.Get("/no-req-schema-valid", func(c *fiber.Ctx) (interface{}, error) {
		return SimpleResponse{ID: 1, Name: "Test", Email: "test@example.com"}, nil
	}, autofiber.WithResponseSchema(SimpleResponse{}))

	req = httptest.NewRequest(http.MethodGet, "/no-req-schema-valid", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
