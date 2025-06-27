package autofiber_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

// testAutoParseRequest is a helper function to test AutoFiber with proper method registration
func testAutoParseRequest(t *testing.T, schema interface{}, handler interface{}, req *http.Request) *http.Response {
	af := autofiber.New()

	// Determine HTTP method from request
	method := req.Method
	switch method {
	case http.MethodGet:
		af.Get("/", handler, autofiber.WithRequestSchema(schema))
	case http.MethodPost:
		af.Post("/", handler, autofiber.WithRequestSchema(schema))
	case http.MethodPut:
		af.Put("/", handler, autofiber.WithRequestSchema(schema))
	case http.MethodDelete:
		af.Delete("/", handler, autofiber.WithRequestSchema(schema))
	default:
		af.Get("/", handler, autofiber.WithRequestSchema(schema))
	}

	resp, err := af.Test(req)
	assert.NoError(t, err)
	return resp
}

// =============================================================================
// GROUP 1: FOCUS ON PARSING (NO VALIDATION RULES)
// =============================================================================

func TestSimple(t *testing.T) {
	called := false
	h := autofiber.Simple(func(c *fiber.Ctx) error {
		called = true
		return c.SendString("ok")
	})

	app := fiber.New()
	app.Get("/", h)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}

func TestAutoParseRequest_ParseFromQuery(t *testing.T) {
	type Req struct {
		Name string `parse:"query:name"`
		Age  string `parse:"query:age"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	req := httptest.NewRequest(http.MethodGet, "/?name=John&age=25", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoParseRequest_ParseFromPath(t *testing.T) {
	type Req struct {
		ID   string `parse:"path:id"`
		Name string `parse:"query:name"`
	}

	app := autofiber.New()
	app.Get("/users/:id", func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}, autofiber.WithRequestSchema(&Req{}))

	req := httptest.NewRequest(http.MethodGet, "/users/123?name=John", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoParseRequest_ParseFromHeader(t *testing.T) {
	type Req struct {
		AuthToken string `parse:"header:Authorization"`
		UserAgent string `parse:"header:User-Agent"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("User-Agent", "TestAgent/1.0")

	resp := testAutoParseRequest(t, &Req{}, handler, req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoParseRequest_ParseFromBody(t *testing.T) {
	type Req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Body = httptest.NewRecorder().Result().Body // This will be empty, but middleware should handle it

	resp := testAutoParseRequest(t, &Req{}, handler, req)
	// Should not return 500 even with empty body
	assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetParsedRequest(t *testing.T) {
	type Req struct {
		Name string
	}

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		// Manually set parsed request in context
		req := &Req{Name: "test"}
		c.Locals("parsed_request", req)

		// Test GetParsedRequest
		parsedReq := autofiber.GetParsedRequest[Req](c)
		assert.NotNil(t, parsedReq)
		assert.Equal(t, "test", parsedReq.Name)

		return c.JSON(parsedReq)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// GROUP 2: FOCUS ON VALIDATION (COMBINED WITH PARSING)
// =============================================================================

func TestAutoParseRequest_WithValidation_ValidData(t *testing.T) {
	type Req struct {
		Name  string `parse:"query:name" validate:"required"`
		Age   int    `parse:"query:age" validate:"required,min=18"`
		Email string `parse:"query:email" validate:"required,email"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	req := httptest.NewRequest(http.MethodGet, "/?name=John&age=25&email=john@example.com", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoParseRequest_WithValidation_MissingRequiredField(t *testing.T) {
	type Req struct {
		Name string `parse:"query:name" validate:"required"`
		Age  int    `parse:"query:age" validate:"required"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	// Missing 'age' field
	req := httptest.NewRequest(http.MethodGet, "/?name=John", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	// Should return validation error
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestAutoParseRequest_WithValidation_InvalidEmail(t *testing.T) {
	type Req struct {
		Email string `parse:"query:email" validate:"required,email"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	// Invalid email format
	req := httptest.NewRequest(http.MethodGet, "/?email=invalid-email", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	// Should return validation error
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestAutoParseRequest_WithValidation_AgeBelowMinimum(t *testing.T) {
	type Req struct {
		Age int `parse:"query:age" validate:"required,min=18"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	// Age below minimum
	req := httptest.NewRequest(http.MethodGet, "/?age=15", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	// Should return validation error
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestAutoParseRequest_WithValidation_ComplexRules(t *testing.T) {
	type Req struct {
		Username string `parse:"query:username" validate:"required,min=3,max=20"`
		Password string `parse:"query:password" validate:"required,min=8"`
		Age      int    `parse:"query:age" validate:"required,min=18,max=100"`
		Email    string `parse:"query:email" validate:"required,email"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	// Valid data
	req := httptest.NewRequest(http.MethodGet, "/?username=john_doe&password=securepass123&age=25&email=john@example.com", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoParseRequest_WithValidation_ComplexRules_InvalidData(t *testing.T) {
	type Req struct {
		Username string `parse:"query:username" validate:"required,min=3,max=20"`
		Password string `parse:"query:password" validate:"required,min=8"`
		Age      int    `parse:"query:age" validate:"required,min=18,max=100"`
		Email    string `parse:"query:email" validate:"required,email"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		return c.JSON(req)
	}

	// Invalid data: short username, short password, invalid age, invalid email
	req := httptest.NewRequest(http.MethodGet, "/?username=jo&password=123&age=15&email=invalid", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	// Should return validation error
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

// =============================================================================
// OTHER MIDDLEWARE TESTS
// =============================================================================

func TestValidateAndJSON(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return autofiber.ValidateAndJSON(c, fiber.Map{"msg": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// 1. Trường hợp có parse tag
// =============================================================================
func TestAutoParseRequest_ParseTag_Query(t *testing.T) {
	type Req struct {
		Name string `parse:"query:name" validate:"required"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		assert.Equal(t, "John", req.Name)
		return c.JSON(req)
	}

	req := httptest.NewRequest(http.MethodGet, "/?name=John", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoParseRequest_ParseTag_Path(t *testing.T) {
	type Req struct {
		ID string `parse:"path:id" validate:"required"`
	}

	app := autofiber.New()
	app.Get("/user/:id", func(c *fiber.Ctx, req *Req) error {
		assert.Equal(t, "42", req.ID)
		return c.JSON(req)
	}, autofiber.WithRequestSchema(&Req{}))

	req := httptest.NewRequest(http.MethodGet, "/user/42", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// 2. Không có parse tag nhưng có json tag
// =============================================================================
func TestAutoParseRequest_JsonTag_Query(t *testing.T) {
	type Req struct {
		Age int `json:"age" validate:"required"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		assert.Equal(t, 30, req.Age)
		return c.JSON(req)
	}

	req := httptest.NewRequest(http.MethodGet, "/?age=30", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAutoParseRequest_JsonTag_Body(t *testing.T) {
	type Req struct {
		Email string `json:"email" validate:"required,email"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		assert.Equal(t, "a@b.com", req.Email)
		return c.JSON(req)
	}

	body := `{"email":"a@b.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	// Không lỗi validate
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// 3. Không có parse tag, không có json tag (dùng tên field)
// =============================================================================
func TestAutoParseRequest_NoTag_FieldName_Query(t *testing.T) {
	type Req struct {
		Username string `validate:"required"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		assert.Equal(t, "abc", req.Username)
		return c.JSON(req)
	}

	req := httptest.NewRequest(http.MethodGet, "/?Username=abc", nil)
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// 4. Không có parse tag, không có json tag, không truyền đúng key (báo lỗi required)
// =============================================================================
func TestAutoParseRequest_NoTag_FieldName_RequiredError(t *testing.T) {
	type Req struct {
		Username string `validate:"required"`
	}

	handler := func(c *fiber.Ctx, req *Req) error {
		// This should not be called due to validation error
		return c.SendStatus(500)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil) // Không truyền Username
	resp := testAutoParseRequest(t, &Req{}, handler, req)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

// =============================================================================
// RESPONSE VALIDATION TESTS
// =============================================================================

func TestValidateAndJSON_ValidData(t *testing.T) {
	type ValidResponse struct {
		ID    int    `json:"id" validate:"required"`
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	app := autofiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		data := &ValidResponse{
			ID:    1,
			Name:  "John Doe",
			Email: "john@example.com",
		}
		return autofiber.ValidateAndJSON(c, data)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateAndJSON_MapData(t *testing.T) {
	app := autofiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		data := fiber.Map{
			"message": "Hello World",
			"status":  "success",
		}
		return autofiber.ValidateAndJSON(c, data)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// INTEGRATION TESTS: REQUEST + RESPONSE VALIDATION
// =============================================================================

func TestRequestAndResponseValidation_Valid(t *testing.T) {
	type Request struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	type Response struct {
		ID    int    `json:"id" validate:"required"`
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	handler := func(c *fiber.Ctx, req *Request) (interface{}, error) {
		return &Response{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		}, nil
	}

	body := `{"name":"John Doe","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	resp := testAutoParseRequest(t, &Request{}, handler, req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestAndResponseValidation_InvalidRequest(t *testing.T) {
	type Request struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	type Response struct {
		ID    int    `json:"id" validate:"required"`
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	handler := func(c *fiber.Ctx, req *Request) (interface{}, error) {
		return &Response{
			ID:    1,
			Name:  req.Name,
			Email: req.Email,
		}, nil
	}

	body := `{"name":"John Doe","email":"invalid-email"}` // Invalid email
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	resp := testAutoParseRequest(t, &Request{}, handler, req)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}
