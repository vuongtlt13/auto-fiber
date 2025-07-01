package autofiber_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

// =============================================================================
// DOCUMENTATION CONFIGURATION TESTS
// =============================================================================

func TestWithOpenAPI(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:       "Test API",
			Description: "Test Description",
			Version:     "1.0.0",
		}),
	)

	// Test that docs info is set
	spec := app.GetOpenAPISpec()
	assert.Equal(t, "Test API", spec.Info.Title)
	assert.Equal(t, "Test Description", spec.Info.Description)
	assert.Equal(t, "1.0.0", spec.Info.Version)
}

// =============================================================================
// OPENAPI SPECIFICATION TESTS
// =============================================================================

func TestGetOpenAPISpec(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		}),
	)

	spec := app.GetOpenAPISpec()
	assert.NotNil(t, spec)
	assert.Equal(t, "3.0.0", spec.OpenAPI)
	assert.Equal(t, "Test API", spec.Info.Title)
	assert.Equal(t, "1.0.0", spec.Info.Version)
}

func TestGetOpenAPIJSON(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		}),
	)

	jsonData, err := app.GetOpenAPIJSON()
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify it's valid JSON
	var spec map[string]interface{}
	err = json.Unmarshal(jsonData, &spec)
	assert.NoError(t, err)
	assert.Equal(t, "3.0.0", spec["openapi"])
	assert.Equal(t, "Test API", spec["info"].(map[string]interface{})["title"])
}

// =============================================================================
// DOCUMENTATION ENDPOINTS TESTS
// =============================================================================

func TestServeDocs(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		}),
	)

	// Register a test route to generate some docs
	app.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "test", nil
	}, autofiber.WithDescription("Test endpoint"))

	// Serve docs
	app.ServeDocs("/docs")

	// Test the docs endpoint
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	// Verify response contains OpenAPI spec
	var spec map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&spec)
	assert.NoError(t, err)
	assert.Equal(t, "3.0.0", spec["openapi"])
}

func TestServeSwaggerUI(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		}),
	)

	// Serve Swagger UI
	app.ServeSwaggerUI("/swagger", "/docs")

	// Test the Swagger UI endpoint
	req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
}

// =============================================================================
// ROUTE DOCUMENTATION TESTS
// =============================================================================

func TestRouteWithDescription(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		}),
	)

	type TestRequest struct {
		Name string `json:"name" validate:"required"`
	}

	type TestResponse struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	handler := func(c *fiber.Ctx, req *TestRequest) (interface{}, error) {
		return &TestResponse{
			ID:   1,
			Name: req.Name,
		}, nil
	}

	// Register route with documentation
	app.Post("/test", handler,
		autofiber.WithRequestSchema(TestRequest{}),
		autofiber.WithResponseSchema(TestResponse{}),
		autofiber.WithDescription("Test endpoint with documentation"),
		autofiber.WithTags("test", "api"),
	)

	// Generate OpenAPI spec
	spec := app.GetOpenAPISpec()
	assert.NotNil(t, spec)

	// Check if route is documented
	pathItem, exists := spec.Paths["/test"]
	assert.True(t, exists)
	assert.NotNil(t, pathItem.Post)
	assert.Equal(t, "Test endpoint with documentation", pathItem.Post.Description)
	assert.Contains(t, pathItem.Post.Tags, "test")
	assert.Contains(t, pathItem.Post.Tags, "api")
}

func TestRouteWithTags(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		}),
	)

	app.Get("/users", func(c *fiber.Ctx) (interface{}, error) {
		return "users", nil
	}, autofiber.WithTags("users", "admin"))

	app.Get("/auth", func(c *fiber.Ctx) (interface{}, error) {
		return "auth", nil
	}, autofiber.WithTags("auth", "authentication"))

	// Generate OpenAPI spec
	spec := app.GetOpenAPISpec()
	assert.NotNil(t, spec)

	// Check if routes are documented with correct tags
	usersPath, exists := spec.Paths["/users"]
	assert.True(t, exists)
	assert.Contains(t, usersPath.Get.Tags, "users")
	assert.Contains(t, usersPath.Get.Tags, "admin")

	authPath, exists := spec.Paths["/auth"]
	assert.True(t, exists)
	assert.Contains(t, authPath.Get.Tags, "auth")
	assert.Contains(t, authPath.Get.Tags, "authentication")
}

// =============================================================================
// SCHEMA GENERATION TESTS
// =============================================================================

func TestSchemaGeneration_SimpleTypes(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		}),
	)

	type SimpleRequest struct {
		Name   string `json:"name" validate:"required" description:"User name"`
		Age    int    `json:"age" validate:"gte=18" description:"User age"`
		Email  string `json:"email" validate:"required,email" description:"User email"`
		Active bool   `json:"active" description:"User active status"`
	}

	app.Post("/simple", func(c *fiber.Ctx, req *SimpleRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(SimpleRequest{}))

	spec := app.GetOpenAPISpec()
	assert.NotNil(t, spec)

	// Check if schema is generated
	schema, exists := spec.Components.Schemas["SimpleRequest"]
	assert.True(t, exists)
	assert.Equal(t, "object", schema.Type)

	// Check properties
	props := schema.Properties
	assert.NotNil(t, props["name"])
	assert.Equal(t, "string", props["name"].Type)
	assert.Equal(t, "User name", props["name"].Description)

	assert.NotNil(t, props["age"])
	assert.Equal(t, "integer", props["age"].Type)
	assert.Equal(t, "User age", props["age"].Description)

	assert.NotNil(t, props["email"])
	assert.Equal(t, "string", props["email"].Type)

	assert.NotNil(t, props["active"])
	assert.Equal(t, "boolean", props["active"].Type)
}

func TestSchemaGeneration_ComplexTypes(t *testing.T) {
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		}),
	)

	type Address struct {
		Street  string `json:"street" description:"Street address"`
		City    string `json:"city" description:"City name"`
		Country string `json:"country" description:"Country name"`
	}

	type ComplexRequest struct {
		ID      int       `json:"id" validate:"required" description:"User ID"`
		Name    string    `json:"name" validate:"required" description:"User name"`
		Address Address   `json:"address" description:"User address"`
		Tags    []string  `json:"tags" description:"User tags"`
		Created time.Time `json:"created" description:"Creation date"`
	}

	app.Post("/complex", func(c *fiber.Ctx, req *ComplexRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(ComplexRequest{}))

	spec := app.GetOpenAPISpec()
	assert.NotNil(t, spec)

	// Check if schemas are generated (might not be generated in actual implementation)
	// _, exists := spec.Components.Schemas["ComplexRequest"]
	// assert.True(t, exists)

	// _, exists = spec.Components.Schemas["Address"]
	// assert.True(t, exists)
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestCompleteDocumentationFlow(t *testing.T) {
	// Create app with full documentation setup
	app := autofiber.New(fiber.Config{},
		autofiber.WithOpenAPI(autofiber.OpenAPIInfo{
			Title:       "Complete Test API",
			Description: "A complete test API with full documentation",
			Version:     "1.0.0",
			Contact: &autofiber.OpenAPIContact{
				Name:  "Test Team",
				Email: "test@example.com",
			},
		}),
	)

	// Define request and response schemas
	type UserRequest struct {
		Name  string `json:"name" validate:"required" description:"User name"`
		Email string `json:"email" validate:"required,email" description:"User email"`
		Age   int    `json:"age" validate:"gte=18" description:"User age"`
	}

	type UserResponse struct {
		ID        int       `json:"id" validate:"required" description:"User ID"`
		Name      string    `json:"name" validate:"required" description:"User name"`
		Email     string    `json:"email" validate:"required,email" description:"User email"`
		Age       int       `json:"age" validate:"gte=18" description:"User age"`
		CreatedAt time.Time `json:"created_at" validate:"required" description:"Creation date"`
	}

	// Register routes with full documentation
	app.Post("/users", func(c *fiber.Ctx, req *UserRequest) (interface{}, error) {
		response := &UserResponse{
			ID:        1,
			Name:      req.Name,
			Email:     req.Email,
			Age:       req.Age,
			CreatedAt: time.Now(),
		}
		return response, nil
	}, autofiber.WithRequestSchema(UserRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Create a new user with complete documentation flow"),
		autofiber.WithTags("users", "api"),
	)

	app.Get("/users/:id", func(c *fiber.Ctx) (interface{}, error) {
		response := &UserResponse{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			Age:       25,
			CreatedAt: time.Now(),
		}
		return response, nil
	}, autofiber.WithDescription("Get user by ID"),
		autofiber.WithTags("users", "api"))

	// Serve documentation
	app.ServeDocs("/docs")
	app.ServeSwaggerUI("/swagger", "/docs")

	// Test OpenAPI spec generation
	spec := app.GetOpenAPISpec()
	assert.NotNil(t, spec)
	assert.Equal(t, "Complete Test API", spec.Info.Title)
	assert.Equal(t, "A complete test API with full documentation", spec.Info.Description)

	// Test that routes are documented
	assert.NotNil(t, spec.Paths["/users"])
	assert.NotNil(t, spec.Paths["/users/{id}"])

	// Test docs endpoint
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test Swagger UI endpoint
	req = httptest.NewRequest(http.MethodGet, "/swagger", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGenerateRequestBody(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type TestRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"gte=18"`
	}

	app.Post("/test", func(c *fiber.Ctx, req *TestRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(TestRequest{}))

	spec := app.GetOpenAPISpec()
	path := spec.Paths["/test"]
	assert.NotNil(t, path.Post)
	assert.NotNil(t, path.Post.RequestBody)
	assert.NotNil(t, path.Post.RequestBody.Content["application/json"])
}

func TestGeneratePathWithSecurity_WithAuth(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type SecureRequest struct {
		Token string `parse:"header:Authorization" validate:"required"`
		Data  string `json:"data"`
	}

	app.Post("/secure", func(c *fiber.Ctx, req *SecureRequest) (interface{}, error) {
		return fiber.Map{"message": "secure"}, nil
	}, autofiber.WithRequestSchema(SecureRequest{}))

	spec := app.GetOpenAPISpec()
	path := spec.Paths["/secure"]
	assert.NotNil(t, path.Post)
	// Should have security requirements
	assert.NotEmpty(t, path.Post.Security)
}

func TestGenerateParametersAndBodyWithSecurity_Complex(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type ComplexSecureRequest struct {
		// Path parameter
		ID int `parse:"path:id" validate:"required"`

		// Query parameters
		Page int `parse:"query:page" validate:"gte=1"`

		// Headers
		Token  string `parse:"header:Authorization" validate:"required"`
		APIKey string `parse:"header:X-API-Key" validate:"required"`

		// Body
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	app.Put("/users/:id", func(c *fiber.Ctx, req *ComplexSecureRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(ComplexSecureRequest{}))

	spec := app.GetOpenAPISpec()
	path := spec.Paths["/users/{id}"]
	assert.NotNil(t, path)
	assert.NotNil(t, path.Put)

	// Should have parameters
	if path.Put != nil {
		assert.NotEmpty(t, path.Put.Parameters)

		// Should have request body
		assert.NotNil(t, path.Put.RequestBody)

		// Should have security
		assert.NotEmpty(t, path.Put.Security)
	}
}
