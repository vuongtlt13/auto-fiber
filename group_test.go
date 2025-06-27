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
	af := autofiber.New()
	group := af.Group("/api")
	return af, group
}

func TestGroup_Get(t *testing.T) {
	af, group := setupGroup()
	group.Get("/get", func(c *fiber.Ctx) error {
		return c.SendString("get ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/get", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Post(t *testing.T) {
	af, group := setupGroup()
	group.Post("/post", func(c *fiber.Ctx) error {
		return c.SendString("post ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/api/post", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Put(t *testing.T) {
	af, group := setupGroup()
	group.Put("/put", func(c *fiber.Ctx) error {
		return c.SendString("put ok")
	})

	req := httptest.NewRequest(http.MethodPut, "/api/put", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Delete(t *testing.T) {
	af, group := setupGroup()
	group.Delete("/delete", func(c *fiber.Ctx) error {
		return c.SendString("delete ok")
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/delete", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Patch(t *testing.T) {
	af, group := setupGroup()
	group.Patch("/patch", func(c *fiber.Ctx) error {
		return c.SendString("patch ok")
	})

	req := httptest.NewRequest(http.MethodPatch, "/api/patch", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Head(t *testing.T) {
	af, group := setupGroup()
	group.Head("/head", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodHead, "/api/head", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_Options(t *testing.T) {
	af, group := setupGroup()
	group.Options("/options", func(c *fiber.Ctx) error {
		return c.SendString("options ok")
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/options", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGroup_All(t *testing.T) {
	af, group := setupGroup()
	group.All("/all", func(c *fiber.Ctx) error {
		return c.SendString("all ok")
	})

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions}
	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/all", nil)
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
	group.Get("/use", func(c *fiber.Ctx) error {
		return c.SendString("use ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/use", nil)
	resp, err := af.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, called)
}
