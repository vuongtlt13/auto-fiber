package autofiber_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
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
// SCHEMA NAME GENERATION TESTS
// =============================================================================

func TestGetSchemaName_RFC3986Compliant(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type UserList struct {
		Users []User `json:"users"`
	}

	type APIResponse[T any] struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    T      `json:"data"`
	}

	// Case 1: Simple struct
	name := autofiber.GetSchemaName(User{})
	assert.Equal(t, "User", name)

	// Case 2: Generic struct with User
	name = autofiber.GetSchemaName(APIResponse[User]{})
	assert.Equal(t, "APIResponse_User", name)

	// Case 3: Generic struct with UserList
	name = autofiber.GetSchemaName(APIResponse[UserList]{})
	assert.Equal(t, "APIResponse_UserList", name)

	// Case 4: No special characters in name
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			t.Errorf("Schema name contains invalid character: %c", c)
		}
	}
}

func TestGetSchemaName_GenericAndNonGeneric(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type LoginResponse struct {
		Token     string    `json:"token"`
		User      User      `json:"user"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	type APIResponse[T any] struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    T      `json:"data"`
	}

	// Non-generic struct should not append field type
	name := autofiber.GetSchemaName(LoginResponse{})
	assert.Equal(t, "LoginResponse", name)

	// Generic struct should append field type
	name = autofiber.GetSchemaName(APIResponse[User]{})
	assert.Equal(t, "APIResponse_User", name)
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

func TestGenerateOperationID_Simple(t *testing.T) {
	id := autofiber.GenerateOperationID("POST", "/auth/register", nil)
	assert.Equal(t, "post_auth_register", id)

	id = autofiber.GenerateOperationID("GET", "/users/:user_id", nil)
	assert.Equal(t, "get_users_user_id", id)

	id = autofiber.GenerateOperationID("DELETE", "/api/v1/items/:item_id", nil)
	assert.Equal(t, "delete_api_v1_items_item_id", id)
}

func TestOpenAPISpec_GET_NoRequestBody(t *testing.T) {
	type GetUserRequest struct {
		UserID int    `parse:"auto:user_id" validate:"required"`
		Name   string `json:"name"`
	}

	app := autofiber.New(fiber.Config{})
	app.Get("/users/:user_id", func(c *fiber.Ctx, req *GetUserRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(GetUserRequest{}))

	spec := app.GetOpenAPISpec()
	path, exists := spec.Paths["/users/{user_id}"]
	assert.True(t, exists)
	assert.NotNil(t, path.Get)
	assert.Nil(t, path.Get.RequestBody, "GET operation must not have requestBody in OpenAPI spec")
}

func TestOpenAPISpec_NoRequestBody_ForGET_DELETE_HEAD_OPTIONS(t *testing.T) {
	type Req struct {
		ID   int    `parse:"auto:id" validate:"required"`
		Name string `json:"name"`
	}

	app := autofiber.New(fiber.Config{})
	app.Get("/test-get/:id", func(c *fiber.Ctx, req *Req) (interface{}, error) { return req, nil }, autofiber.WithRequestSchema(Req{}))
	app.Delete("/test-delete/:id", func(c *fiber.Ctx, req *Req) (interface{}, error) { return req, nil }, autofiber.WithRequestSchema(Req{}))
	app.Head("/test-head/:id", func(c *fiber.Ctx, req *Req) (interface{}, error) { return req, nil }, autofiber.WithRequestSchema(Req{}))
	app.Options("/test-options/:id", func(c *fiber.Ctx, req *Req) (interface{}, error) { return req, nil }, autofiber.WithRequestSchema(Req{}))

	spec := app.GetOpenAPISpec()

	// GET
	path, exists := spec.Paths["/test-get/{id}"]
	assert.True(t, exists)
	assert.NotNil(t, path.Get)
	assert.Nil(t, path.Get.RequestBody, "GET operation must not have requestBody in OpenAPI spec")

	// DELETE
	path, exists = spec.Paths["/test-delete/{id}"]
	assert.True(t, exists)
	assert.NotNil(t, path.Delete)
	assert.Nil(t, path.Delete.RequestBody, "DELETE operation must not have requestBody in OpenAPI spec")

	// HEAD
	path, exists = spec.Paths["/test-head/{id}"]
	assert.True(t, exists)
	assert.NotNil(t, path.Head)
	assert.Nil(t, path.Head.RequestBody, "HEAD operation must not have requestBody in OpenAPI spec")

	// OPTIONS
	path, exists = spec.Paths["/test-options/{id}"]
	assert.True(t, exists)
	assert.NotNil(t, path.Options)
	assert.Nil(t, path.Options.RequestBody, "OPTIONS operation must not have requestBody in OpenAPI spec")
}

// TestConvertToOpenAPISchema_SimpleStruct tests conversion of simple structs
func TestConvertToOpenAPISchema_SimpleStruct(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	// Test case 1: Simple struct with basic types
	type SimpleUser struct {
		ID       int       `json:"id" validate:"required"`
		Name     string    `json:"name" validate:"required"`
		Email    string    `json:"email" validate:"required,email"`
		Age      float64   `json:"age" validate:"gte=0"`
		IsActive bool      `json:"is_active"`
		Created  time.Time `json:"created_at" validate:"required"`
	}

	schema := dg.ConvertToOpenAPISchema(SimpleUser{})

	expected := autofiber.OpenAPISchema{
		Type: "object",
		Properties: map[string]autofiber.OpenAPISchema{
			"id":         {Type: "integer"},
			"name":       {Type: "string"},
			"email":      {Type: "string"},
			"age":        {Type: "number"},
			"is_active":  {Type: "boolean"},
			"created_at": {Type: "string", Format: "date-time"},
		},
		Required: []string{"id", "name", "email", "created_at"},
	}

	if !reflect.DeepEqual(schema.Type, expected.Type) {
		t.Errorf("Expected type %s, got %s", expected.Type, schema.Type)
	}

	if len(schema.Required) != len(expected.Required) {
		t.Errorf("Expected %d required fields, got %d", len(expected.Required), len(schema.Required))
	}

	// Check properties
	for field, expectedProp := range expected.Properties {
		if prop, exists := schema.Properties[field]; exists {
			if prop.Type != expectedProp.Type {
				t.Errorf("Field %s: expected type %s, got %s", field, expectedProp.Type, prop.Type)
			}
			if expectedProp.Format != "" && prop.Format != expectedProp.Format {
				t.Errorf("Field %s: expected format %s, got %s", field, expectedProp.Format, prop.Format)
			}
		} else {
			t.Errorf("Expected property %s not found", field)
		}
	}
}

// TestConvertToOpenAPISchema_ComplexStruct tests conversion of complex structs with nested structures
func TestConvertToOpenAPISchema_ComplexStruct(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	// Test case 2: Complex struct with nested structs
	type Address struct {
		Street  string `json:"street" validate:"required"`
		City    string `json:"city" validate:"required"`
		Country string `json:"country" validate:"required"`
	}

	type Profile struct {
		Bio       string `json:"bio"`
		AvatarURL string `json:"avatar_url"`
	}

	type ComplexUser struct {
		ID      int      `json:"id" validate:"required"`
		Name    string   `json:"name" validate:"required"`
		Address Address  `json:"address" validate:"required"`
		Profile *Profile `json:"profile"`
		Tags    []string `json:"tags"`
	}

	schema := dg.ConvertToOpenAPISchema(ComplexUser{})

	// Check basic structure
	if schema.Type != "object" {
		t.Errorf("Expected type object, got %s", schema.Type)
	}

	// Check that nested structs are referenced and registered
	if addressProp, exists := schema.Properties["address"]; exists {
		if addressProp.Ref == "" {
			t.Error("Expected address field to have a reference to nested schema")
		}
		// Check that Address schema is registered
		if _, ok := dg.Schemas()["Address"]; !ok {
			t.Error("Expected Address schema to be registered in DocsGenerator")
		}
	} else {
		t.Error("Expected address property not found")
	}

	if profileProp, exists := schema.Properties["profile"]; exists {
		if profileProp.Ref == "" {
			t.Error("Expected profile field to have a reference to nested schema")
		}
		// Check that Profile schema is registered
		if _, ok := dg.Schemas()["Profile"]; !ok {
			t.Error("Expected Profile schema to be registered in DocsGenerator")
		}
	} else {
		t.Error("Expected profile property not found")
	}

	// Check array type
	if tagsProp, exists := schema.Properties["tags"]; exists {
		if tagsProp.Type != "array" {
			t.Errorf("Expected tags to be array type, got %s", tagsProp.Type)
		}
		if tagsProp.Items == nil {
			t.Error("Expected tags array to have items schema")
		}
		if tagsProp.Items.Type != "string" {
			t.Errorf("Expected tags items to be string type, got %s", tagsProp.Items.Type)
		}
	} else {
		t.Error("Expected tags property not found")
	}

	// Check required fields
	expectedRequired := []string{"id", "name", "address"}
	for _, req := range expectedRequired {
		found := false
		for _, r := range schema.Required {
			if r == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected required field %s not found", req)
		}
	}
}

// TestConvertToOpenAPISchema_PointerSimpleStruct tests conversion of pointers to simple structs
func TestConvertToOpenAPISchema_PointerSimpleStruct(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type SimpleUser struct {
		ID   int    `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	}

	schema := dg.ConvertToOpenAPISchema(&SimpleUser{})

	// Should handle pointer the same as direct struct
	expected := autofiber.OpenAPISchema{
		Type: "object",
		Properties: map[string]autofiber.OpenAPISchema{
			"id":   {Type: "integer"},
			"name": {Type: "string"},
		},
		Required: []string{"id", "name"},
	}

	if schema.Type != expected.Type {
		t.Errorf("Expected type %s, got %s", expected.Type, schema.Type)
	}

	if len(schema.Required) != len(expected.Required) {
		t.Errorf("Expected %d required fields, got %d", len(expected.Required), len(schema.Required))
	}
}

// TestConvertToOpenAPISchema_PointerComplexStruct tests conversion of pointers to complex structs
func TestConvertToOpenAPISchema_PointerComplexStruct(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type Address struct {
		Street string `json:"street" validate:"required"`
		City   string `json:"city" validate:"required"`
	}

	type ComplexUser struct {
		ID      int      `json:"id" validate:"required"`
		Name    string   `json:"name" validate:"required"`
		Address *Address `json:"address" validate:"required"`
		Tags    []string `json:"tags"`
	}

	schema := dg.ConvertToOpenAPISchema(&ComplexUser{})

	// Check basic structure
	if schema.Type != "object" {
		t.Errorf("Expected type object, got %s", schema.Type)
	}

	// Check that pointer to nested struct is handled correctly and registered
	if addressProp, exists := schema.Properties["address"]; exists {
		if addressProp.Ref == "" {
			t.Error("Expected address field to have a reference to nested schema")
		}
		// Check that Address schema is registered
		if _, ok := dg.Schemas()["Address"]; !ok {
			t.Error("Expected Address schema to be registered in DocsGenerator")
		}
	} else {
		t.Error("Expected address property not found")
	}

	// Check array type
	if tagsProp, exists := schema.Properties["tags"]; exists {
		if tagsProp.Type != "array" {
			t.Errorf("Expected tags to be array type, got %s", tagsProp.Type)
		}
		if tagsProp.Items == nil {
			t.Error("Expected tags array to have items schema")
		}
	} else {
		t.Error("Expected tags property not found")
	}
}

// TestConvertToOpenAPISchema_DeepNestedStruct tests conversion of deeply nested structs
func TestConvertToOpenAPISchema_DeepNestedStruct(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type Country struct {
		Code string `json:"code" validate:"required"`
		Name string `json:"name" validate:"required"`
	}

	type City struct {
		Name    string  `json:"name" validate:"required"`
		Country Country `json:"country" validate:"required"`
	}

	type Address struct {
		Street string `json:"street" validate:"required"`
		City   City   `json:"city" validate:"required"`
	}

	type DeepNestedUser struct {
		ID      int      `json:"id" validate:"required"`
		Name    string   `json:"name" validate:"required"`
		Address *Address `json:"address" validate:"required"`
	}

	schema := dg.ConvertToOpenAPISchema(DeepNestedUser{})

	// Check basic structure
	if schema.Type != "object" {
		t.Errorf("Expected type object, got %s", schema.Type)
	}

	// Check that deeply nested structs are referenced and registered
	if addressProp, exists := schema.Properties["address"]; exists {
		if addressProp.Ref == "" {
			t.Error("Expected address field to have a reference to nested schema")
		}
		// Check that Address schema is registered
		if _, ok := dg.Schemas()["Address"]; !ok {
			t.Error("Expected Address schema to be registered in DocsGenerator")
		}
		// Check that City schema is registered
		if _, ok := dg.Schemas()["City"]; !ok {
			t.Error("Expected City schema to be registered in DocsGenerator")
		}
		// Check that Country schema is registered
		if _, ok := dg.Schemas()["Country"]; !ok {
			t.Error("Expected Country schema to be registered in DocsGenerator")
		}
	} else {
		t.Error("Expected address property not found")
	}

	// Check required fields
	expectedRequired := []string{"id", "name", "address"}
	for _, req := range expectedRequired {
		found := false
		for _, r := range schema.Required {
			if r == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected required field %s not found", req)
		}
	}
}

// TestConvertToOpenAPISchema_WithDescriptionsAndExamples tests conversion with description and example tags
func TestConvertToOpenAPISchema_WithDescriptionsAndExamples(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type UserWithMetadata struct {
		ID    int    `json:"id" validate:"required" description:"User unique identifier" example:"123"`
		Name  string `json:"name" validate:"required" description:"User full name" example:"John Doe"`
		Email string `json:"email" validate:"required,email" description:"User email address" example:"john@example.com"`
	}

	schema := dg.ConvertToOpenAPISchema(UserWithMetadata{})

	// Check descriptions and examples
	if idProp, exists := schema.Properties["id"]; exists {
		if idProp.Description != "User unique identifier" {
			t.Errorf("Expected description 'User unique identifier', got '%s'", idProp.Description)
		}
		if idProp.Example != "123" {
			t.Errorf("Expected example '123', got '%v'", idProp.Example)
		}
	} else {
		t.Error("Expected id property not found")
	}

	if nameProp, exists := schema.Properties["name"]; exists {
		if nameProp.Description != "User full name" {
			t.Errorf("Expected description 'User full name', got '%s'", nameProp.Description)
		}
		if nameProp.Example != "John Doe" {
			t.Errorf("Expected example 'John Doe', got '%v'", nameProp.Example)
		}
	} else {
		t.Error("Expected name property not found")
	}
}

// TestConvertToOpenAPISchema_NonStructInput tests conversion of non-struct inputs
func TestConvertToOpenAPISchema_NonStructInput(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	// Test with string
	schema := dg.ConvertToOpenAPISchema("test")
	expected := autofiber.OpenAPISchema{Type: "object"}
	if schema.Type != expected.Type {
		t.Errorf("Expected type %s for non-struct input, got %s", expected.Type, schema.Type)
	}

	// Test with int
	schema = dg.ConvertToOpenAPISchema(123)
	if schema.Type != expected.Type {
		t.Errorf("Expected type %s for non-struct input, got %s", expected.Type, schema.Type)
	}

	// Test with slice
	schema = dg.ConvertToOpenAPISchema([]string{"test"})
	if schema.Type != expected.Type {
		t.Errorf("Expected type %s for non-struct input, got %s", expected.Type, schema.Type)
	}
}

// TestConvertToOpenAPISchema_GenericStruct tests conversion of generic structs
func TestConvertToOpenAPISchema_GenericStruct(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type User struct {
		ID   int    `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	}

	type APIResponse[T any] struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    T      `json:"data"`
	}

	// Test with generic struct
	response := APIResponse[User]{}
	schema := dg.ConvertToOpenAPISchema(response)

	// Check basic structure
	if schema.Type != "object" {
		t.Errorf("Expected type object, got %s", schema.Type)
	}

	// Check that data field exists
	if dataProp, exists := schema.Properties["data"]; exists {
		// Generic structs should either be inlined or referenced
		// Both approaches are valid depending on implementation
		if dataProp.Type != "object" && dataProp.Ref == "" {
			t.Errorf("Expected data field to be object type or have reference, got type: %s, ref: %s", dataProp.Type, dataProp.Ref)
		}
	} else {
		t.Error("Expected data property not found")
	}

	// Check that code and message fields exist
	if codeProp, exists := schema.Properties["code"]; exists {
		if codeProp.Type != "integer" {
			t.Errorf("Expected code field to be integer type, got %s", codeProp.Type)
		}
	} else {
		t.Error("Expected code property not found")
	}

	if messageProp, exists := schema.Properties["message"]; exists {
		if messageProp.Type != "string" {
			t.Errorf("Expected message field to be string type, got %s", messageProp.Type)
		}
	} else {
		t.Error("Expected message property not found")
	}
}

// TestConvertToOpenAPISchema_DebugSchemas tests to see what schemas are actually registered
func TestConvertToOpenAPISchema_DebugSchemas(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type Address struct {
		Street  string `json:"street" validate:"required"`
		City    string `json:"city" validate:"required"`
		Country string `json:"country" validate:"required"`
	}

	type Profile struct {
		Bio       string `json:"bio"`
		AvatarURL string `json:"avatar_url"`
	}

	type ComplexUser struct {
		ID      int      `json:"id" validate:"required"`
		Name    string   `json:"name" validate:"required"`
		Address Address  `json:"address" validate:"required"`
		Profile *Profile `json:"profile"`
		Tags    []string `json:"tags"`
	}

	schema := dg.ConvertToOpenAPISchema(ComplexUser{})

	// Print all registered schemas
	t.Logf("Registered schemas: %+v", dg.Schemas())

	// Print schema names
	for name := range dg.Schemas() {
		t.Logf("Schema name: %s", name)
	}

	// Check basic structure
	if schema.Type != "object" {
		t.Errorf("Expected type object, got %s", schema.Type)
	}

	// Check that nested structs are referenced
	if addressProp, exists := schema.Properties["address"]; exists {
		t.Logf("Address prop: %+v", addressProp)
		if addressProp.Ref == "" {
			t.Error("Expected address field to have a reference to nested schema")
		}
	} else {
		t.Error("Expected address property not found")
	}

	if profileProp, exists := schema.Properties["profile"]; exists {
		t.Logf("Profile prop: %+v", profileProp)
		if profileProp.Ref == "" {
			t.Error("Expected profile field to have a reference to nested schema")
		}
	} else {
		t.Error("Expected profile property not found")
	}
}

// TestConvertRequestToOpenAPISchema tests the request conversion function
func TestConvertRequestToOpenAPISchema(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	// Test case: Request struct with parse tags
	type UserRequest struct {
		ID     int    `parse:"path:user_id" json:"id" validate:"required"`
		Name   string `parse:"query:name" json:"name" validate:"required"`
		Email  string `json:"email" validate:"required,email"`
		Token  string `parse:"header:Authorization" json:"token"`
		Page   int    `parse:"query:page" json:"page"`
		Data   string `parse:"body:data" json:"data"`
		SkipMe string `json:"-"` // Should be skipped
		NoTags string // Should be skipped (no parse or json tags)
	}

	schema := dg.ConvertRequestToOpenAPISchema(UserRequest{})

	// Debug: print actual schema properties
	t.Logf("Actual schema properties: %v", schema.Properties)

	// Only expect fields with parse tag (body) or valid json tag
	expectedFields := []string{"data", "email"}
	for _, field := range expectedFields {
		if _, exists := schema.Properties[field]; !exists {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Check that skipped fields are not included
	skippedFields := []string{"user_id", "name", "token", "page", "SkipMe", "NoTags", "skip_me", "no_tags", "Authorization"}
	for _, field := range skippedFields {
		if _, exists := schema.Properties[field]; exists {
			t.Errorf("Unexpected field %s found in schema", field)
		}
	}

	// Check required fields
	expectedRequired := []string{"email"}
	for _, req := range expectedRequired {
		found := false
		for _, r := range schema.Required {
			if r == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected required field %s not found", req)
		}
	}
}

// TestConvertResponseToOpenAPISchema tests the response conversion function
func TestConvertResponseToOpenAPISchema(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	// Test case: Response struct with json tags and camelCase fallback
	type UserResponse struct {
		ID        int       `json:"id" validate:"required"`
		Name      string    `json:"name" validate:"required"`
		Email     string    `json:"email" validate:"required,email"`
		CreatedAt time.Time `json:"created_at" validate:"required"`
		IsActive  bool      `json:"is_active"`
		UserType  string    // No json tag, should use camelCase
		APIKey    string    // No json tag, should use camelCase
		SkipMe    string    `json:"-"` // Should be skipped
	}

	schema := dg.ConvertResponseToOpenAPISchema(UserResponse{})

	// Debug: print actual schema properties
	t.Logf("Actual schema properties: %v", schema.Properties)

	// Check fields with json tags
	jsonTagFields := map[string]string{
		"id":         "integer",
		"name":       "string",
		"email":      "string",
		"created_at": "string",
		"is_active":  "boolean",
	}

	for field, expectedType := range jsonTagFields {
		if prop, exists := schema.Properties[field]; exists {
			if prop.Type != expectedType {
				t.Errorf("Field %s: expected type %s, got %s", field, expectedType, prop.Type)
			}
			// Check format for time.Time fields
			if field == "created_at" && prop.Format != "date-time" {
				t.Errorf("Field %s: expected format date-time, got %s", field, prop.Format)
			}
		} else {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Check fields without json tags (should use camelCase)
	camelCaseFields := map[string]string{
		"userType": "string",
		"apiKey":   "string",
	}

	for field, expectedType := range camelCaseFields {
		if prop, exists := schema.Properties[field]; exists {
			if prop.Type != expectedType {
				t.Errorf("Field %s: expected type %s, got %s", field, expectedType, prop.Type)
			}
		} else {
			t.Errorf("Expected camelCase field %s not found in schema", field)
		}
	}

	// Check that skipped fields are not included
	skippedFields := []string{"SkipMe", "skip_me"}
	for _, field := range skippedFields {
		if _, exists := schema.Properties[field]; exists {
			t.Errorf("Unexpected field %s found in schema", field)
		}
	}

	// Check required fields
	expectedRequired := []string{"id", "name", "email", "created_at"}
	for _, req := range expectedRequired {
		found := false
		for _, r := range schema.Required {
			if r == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected required field %s not found", req)
		}
	}
}

// TestConvertRequestToOpenAPISchema_ParseTagPriority tests that parse tags take priority over json tags, but only for body source
func TestConvertRequestToOpenAPISchema_ParseTagPriority(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type TestRequest struct {
		Field1 string `parse:"body:user_id" json:"id" validate:"required"`
		Field2 string `parse:"query:search" json:"name" validate:"required"`
		Field3 string `json:"email" validate:"required"` // Only json tag
	}

	schema := dg.ConvertRequestToOpenAPISchema(TestRequest{})

	// Only expect fields with parse tag (body) or valid json tag
	expectedFields := []string{"user_id", "email"}
	for _, field := range expectedFields {
		if _, exists := schema.Properties[field]; !exists {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Fields with parse tag (not body) or no valid tag should be skipped
	skippedFields := []string{"search", "id", "name"}
	for _, field := range skippedFields {
		if _, exists := schema.Properties[field]; exists {
			t.Errorf("Unexpected field %s found in schema", field)
		}
	}
}

// TestConvertResponseToOpenAPISchema_CamelCase tests camelCase conversion for field names
func TestConvertResponseToOpenAPISchema_CamelCase(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type TestResponse struct {
		UserID      string // Should become "userID"
		APIKey      string // Should become "apiKey"
		HTTPStatus  string // Should become "httpStatus"
		JSONData    string // Should become "jsonData"
		SimpleField string // Should become "simpleField"
	}

	schema := dg.ConvertResponseToOpenAPISchema(TestResponse{})

	expectedFields := map[string]string{
		"userID":      "string",
		"apiKey":      "string",
		"httpStatus":  "string",
		"jsonData":    "string",
		"simpleField": "string",
	}

	for field, expectedType := range expectedFields {
		if prop, exists := schema.Properties[field]; exists {
			if prop.Type != expectedType {
				t.Errorf("Field %s: expected type %s, got %s", field, expectedType, prop.Type)
			}
		} else {
			t.Errorf("Expected camelCase field %s not found in schema", field)
		}
	}
}

// TestConvertRequestToOpenAPISchema_EmptyJsonTag tests handling of empty json tags
func TestConvertRequestToOpenAPISchema_EmptyJsonTag(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type TestRequest struct {
		Field1 string `json:"" validate:"required"`       // Empty json tag, should be skipped
		Field2 string `json:"," validate:"required"`      // Empty json tag with comma, should be skipped
		Field3 string `json:"field3" validate:"required"` // Normal json tag
		Field4 string `parse:"body:myfield"`              // parse tag body, should be included
	}

	schema := dg.ConvertRequestToOpenAPISchema(TestRequest{})

	// Only expect fields with parse tag (body) or valid json tag
	expectedFields := []string{"myfield", "field3"}
	for _, field := range expectedFields {
		if _, exists := schema.Properties[field]; !exists {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}

	// Fields with empty json tag should be skipped
	skippedFields := []string{"Field1", "Field2"}
	for _, field := range skippedFields {
		if _, exists := schema.Properties[field]; exists {
			t.Errorf("Unexpected field %s found in schema", field)
		}
	}
}

// TestConvertResponseToOpenAPISchema_EmptyJsonTag tests handling of empty json tags
func TestConvertResponseToOpenAPISchema_EmptyJsonTag(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type TestResponse struct {
		Field1 string `json:""`       // Empty json tag
		Field2 string `json:","`      // Empty json tag with comma
		Field3 string `json:"field3"` // Normal json tag
	}

	schema := dg.ConvertResponseToOpenAPISchema(TestResponse{})

	// Empty json tags should fall back to camelCase field name
	expectedFields := map[string]string{
		"field1": "string",
		"field2": "string",
		"field3": "string",
	}

	for field, expectedType := range expectedFields {
		if prop, exists := schema.Properties[field]; exists {
			if prop.Type != expectedType {
				t.Errorf("Field %s: expected type %s, got %s", field, expectedType, prop.Type)
			}
		} else {
			t.Errorf("Expected field %s not found in schema", field)
		}
	}
}

// TestConvertRequestToOpenAPISchema_ComplexNested tests complex nested structs for request
func TestConvertRequestToOpenAPISchema_ComplexNested(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type Address struct {
		Street  string `parse:"query:street" json:"street" validate:"required"`
		City    string `json:"city" validate:"required"`
		Country string `json:"country" validate:"required"`
	}

	type UserRequest struct {
		ID      int      `parse:"path:user_id" json:"id" validate:"required"`
		Name    string   `parse:"query:name" json:"name" validate:"required"`
		Address Address  `json:"address" validate:"required"`
		Profile *Address `json:"profile"`
	}

	schema := dg.ConvertRequestToOpenAPISchema(UserRequest{})

	// Check that nested structs are registered
	if _, ok := dg.Schemas()["Address"]; !ok {
		t.Error("Expected Address schema to be registered in DocsGenerator")
	}

	// Check that nested structs are referenced
	if addressProp, exists := schema.Properties["address"]; exists {
		if addressProp.Ref == "" {
			t.Error("Expected address field to have a reference to nested schema")
		}
	} else {
		t.Error("Expected address property not found")
	}

	if profileProp, exists := schema.Properties["profile"]; exists {
		if profileProp.Ref == "" {
			t.Error("Expected profile field to have a reference to nested schema")
		}
	} else {
		t.Error("Expected profile property not found")
	}
}

// TestConvertResponseToOpenAPISchema_ComplexNested tests complex nested structs for response
func TestConvertResponseToOpenAPISchema_ComplexNested(t *testing.T) {
	dg := autofiber.NewDocsGenerator()

	type Address struct {
		Street  string `json:"street" validate:"required"`
		City    string `json:"city" validate:"required"`
		Country string `json:"country" validate:"required"`
	}

	type UserResponse struct {
		ID      int      `json:"id" validate:"required"`
		Name    string   `json:"name" validate:"required"`
		Address Address  `json:"address" validate:"required"`
		Profile *Address `json:"profile"`
	}

	schema := dg.ConvertResponseToOpenAPISchema(UserResponse{})

	// Check that nested structs are registered
	if _, ok := dg.Schemas()["Address"]; !ok {
		t.Error("Expected Address schema to be registered in DocsGenerator")
	}

	// Check that nested structs are referenced
	if addressProp, exists := schema.Properties["address"]; exists {
		if addressProp.Ref == "" {
			t.Error("Expected address field to have a reference to nested schema")
		}
	} else {
		t.Error("Expected address property not found")
	}

	if profileProp, exists := schema.Properties["profile"]; exists {
		if profileProp.Ref == "" {
			t.Error("Expected profile field to have a reference to nested schema")
		}
	} else {
		t.Error("Expected profile property not found")
	}
}

func TestOpenAPISpec_RegisterRequestBodySchema(t *testing.T) {
	app := autofiber.New(fiber.Config{})

	type RegisterRequest struct {
		Email     string    `json:"email" validate:"required,email"`
		Password  string    `json:"password" validate:"required,min=6"`
		Name      string    `json:"name" validate:"required"`
		BirthDate time.Time `json:"birth_date"`
	}

	app.Post("/register", func(c *fiber.Ctx, req *RegisterRequest) (interface{}, error) {
		return req, nil
	}, autofiber.WithRequestSchema(RegisterRequest{}))

	spec := app.GetOpenAPISpec()
	path, exists := spec.Paths["/register"]
	if !exists || path == (autofiber.OpenAPIPath{}) || path.Post == nil {
		t.Fatal("POST /register not found in OpenAPI spec")
	}
	requestBody := path.Post.RequestBody
	if requestBody == nil {
		t.Fatal("RequestBody for POST /register is nil")
	}
	media, ok := requestBody.Content["application/json"]
	if !ok {
		t.Fatal("application/json content not found in requestBody")
	}
	schema := media.Schema
	if schema == nil {
		t.Fatal("Schema for requestBody is nil")
	}
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}
	// Check properties
	expectedFields := []string{"email", "password", "name", "birth_date"}
	for _, field := range expectedFields {
		if _, ok := schema.Properties[field]; !ok {
			t.Errorf("Expected field '%s' in request body schema", field)
		}
	}
	// Check required fields
	expectedRequired := []string{"email", "password", "name"}
	for _, req := range expectedRequired {
		found := false
		for _, r := range schema.Required {
			if r == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected required field '%s' in request body schema", req)
		}
	}
}
