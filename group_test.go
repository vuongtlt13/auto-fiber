package autofiber_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func setupGroup() (*autofiber.AutoFiber, *autofiber.AutoFiberGroup) {
	af := autofiber.New(fiber.Config{})
	group := af.Group("/api")
	return af, group
}

func TestGroup_Get(t *testing.T) {
	af, group := setupGroup()
	group.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "test", nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Post(t *testing.T) {
	af, group := setupGroup()
	group.Post("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "post ok", nil
	})

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Put(t *testing.T) {
	af, group := setupGroup()
	group.Put("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "put ok", nil
	})

	req := httptest.NewRequest(http.MethodPut, "/api/test", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Delete(t *testing.T) {
	af, group := setupGroup()
	group.Delete("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "delete ok", nil
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/test", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Patch(t *testing.T) {
	af, group := setupGroup()
	group.Patch("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "patch ok", nil
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/test", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Head(t *testing.T) {
	af, group := setupGroup()
	group.Head("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "head ok", nil
	})

	req := httptest.NewRequest(http.MethodHead, "/api/test", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Options(t *testing.T) {
	af, group := setupGroup()
	group.Options("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "options ok", nil
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_All(t *testing.T) {
	af, group := setupGroup()
	group.All("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "all ok", nil
	})

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/test", nil)
		resp, err := af.Test(req)
		assert.NoError(t, err)
		// HEAD returns 200 with empty body, others return 200 with body
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestGroup_Use(t *testing.T) {
	af, group := setupGroup()
	called := false
	group.Use(func(c *fiber.Ctx) error {
		called = true
		return c.Next()
	})
	group.Get("/test", func(c *fiber.Ctx) (interface{}, error) {
		return "use ok", nil
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}

func TestGroup_Docs_AddRoute(t *testing.T) {
	af, group := setupGroup()

	type Req struct {
		Name string `json:"name"`
	}
	type Resp struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	handler := func(c *fiber.Ctx, req *Req) (interface{}, error) {
		return &Resp{ID: 1, Name: req.Name}, nil
	}

	// Register a group route with schema and docs description
	group.Post("/docs-test", handler,
		autofiber.WithRequestSchema(Req{}),
		autofiber.WithResponseSchema(Resp{}),
		autofiber.WithDescription("Test group docs add route"),
		autofiber.WithTags("group", "docs"),
	)

	spec := af.GetOpenAPISpec()
	assert.NotNil(t, spec)
	_, exists := spec.Paths["/api/docs-test"]
	assert.True(t, exists, "Group route should be added to OpenAPI spec")
	if exists {
		assert.NotNil(t, spec.Paths["/api/docs-test"].Post)
		assert.Contains(t, spec.Paths["/api/docs-test"].Post.Tags, "group")
		assert.Equal(t, "Test group docs add route", spec.Paths["/api/docs-test"].Post.Description)
	}
}
