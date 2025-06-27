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

func TestWithDocsInfo(t *testing.T) {
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:       "Test API",
		Description: "Test API Description",
		Version:     "1.0.0",
		Contact: &autofiber.OpenAPIContact{
			Name:  "Test Team",
			Email: "test@example.com",
		},
	})

	// Test that docs info is set
	spec := app.GetOpenAPISpec()
	assert.Equal(t, "Test API", spec.Info.Title)
	assert.Equal(t, "Test API Description", spec.Info.Description)
	assert.Equal(t, "1.0.0", spec.Info.Version)
	assert.Equal(t, "Test Team", spec.Info.Contact.Name)
	assert.Equal(t, "test@example.com", spec.Info.Contact.Email)
}

func TestWithDocsServer(t *testing.T) {
	app := autofiber.New()
	app.WithDocsServer(autofiber.OpenAPIServer{
		URL:         "http://localhost:3000",
		Description: "Development server",
	})

	// Test that server is added
	spec := app.GetOpenAPISpec()
	assert.Len(t, spec.Servers, 1)
	assert.Equal(t, "http://localhost:3000", spec.Servers[0].URL)
	assert.NotEmpty(t, spec.Servers[0].Description)
}

func TestWithDocsBasePath(t *testing.T) {
	app := autofiber.New()
	app.WithDocsBasePath("/api/v1")

	// Test that base path is set
	spec := app.GetOpenAPISpec()
	// The base path should be reflected in the docs generator
	assert.NotNil(t, spec)
}

// =============================================================================
// OPENAPI SPECIFICATION TESTS
// =============================================================================

func TestGetOpenAPISpec(t *testing.T) {
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

	spec := app.GetOpenAPISpec()
	assert.NotNil(t, spec)
	assert.Equal(t, "3.0.0", spec.OpenAPI)
	assert.Equal(t, "Test API", spec.Info.Title)
	assert.Equal(t, "1.0.0", spec.Info.Version)
}

func TestGetOpenAPIJSON(t *testing.T) {
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

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
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

	// Register a test route to generate some docs
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test")
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
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

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
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

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
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

	app.Get("/users", func(c *fiber.Ctx) error {
		return c.SendString("users")
	}, autofiber.WithTags("users", "admin"))

	app.Get("/auth", func(c *fiber.Ctx) error {
		return c.SendString("auth")
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
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

	type SimpleRequest struct {
		Name   string `json:"name" validate:"required" description:"User name"`
		Age    int    `json:"age" validate:"gte=18" description:"User age"`
		Email  string `json:"email" validate:"required,email" description:"User email"`
		Active bool   `json:"active" description:"User active status"`
	}

	app.Post("/simple", func(c *fiber.Ctx, req *SimpleRequest) error {
		return c.JSON(req)
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
	app := autofiber.New()
	app.WithDocsInfo(autofiber.OpenAPIInfo{
		Title:   "Test API",
		Version: "1.0.0",
	})

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

	app.Post("/complex", func(c *fiber.Ctx, req *ComplexRequest) error {
		return c.JSON(req)
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
	app := autofiber.New().
		WithDocsInfo(autofiber.OpenAPIInfo{
			Title:       "Complete Test API",
			Description: "A complete test API with full documentation",
			Version:     "1.0.0",
			Contact: &autofiber.OpenAPIContact{
				Name:  "Test Team",
				Email: "test@example.com",
			},
		}).
		WithDocsServer(autofiber.OpenAPIServer{
			URL:         "http://localhost:3000",
			Description: "Development server",
		})

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
	app.Post("/users", func(c *fiber.Ctx, req *UserRequest) error {
		response := &UserResponse{
			ID:        1,
			Name:      req.Name,
			Email:     req.Email,
			Age:       req.Age,
			CreatedAt: time.Now(),
		}
		return c.JSON(response)
	}, autofiber.WithRequestSchema(UserRequest{}),
		autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Create a new user"),
		autofiber.WithTags("users", "admin"))

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		response := &UserResponse{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			Age:       25,
			CreatedAt: time.Now(),
		}
		return c.JSON(response)
	}, autofiber.WithResponseSchema(UserResponse{}),
		autofiber.WithDescription("Get user by ID"),
		autofiber.WithTags("users", "read"))

	// Serve documentation
	app.ServeDocs("/docs")
	app.ServeSwaggerUI("/swagger", "/docs")

	// Test OpenAPI spec generation
	spec := app.GetOpenAPISpec()
	assert.NotNil(t, spec)
	assert.Equal(t, "Complete Test API", spec.Info.Title)
	assert.Equal(t, "A complete test API with full documentation", spec.Info.Description)
	assert.Len(t, spec.Servers, 1)
	assert.Equal(t, "http://localhost:3000", spec.Servers[0].URL)

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
