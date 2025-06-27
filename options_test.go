package autofiber_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestWithRequestSchema(t *testing.T) {
	type Req struct {
		Name string `json:"name"`
	}

	option := autofiber.WithRequestSchema(&Req{})
	opts := &autofiber.RouteOptions{}

	option(opts)

	assert.NotNil(t, opts.RequestSchema)
	assert.IsType(t, &Req{}, opts.RequestSchema)
}

func TestWithResponseSchema(t *testing.T) {
	type Resp struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	option := autofiber.WithResponseSchema(&Resp{})
	opts := &autofiber.RouteOptions{}

	option(opts)

	assert.NotNil(t, opts.ResponseSchema)
	assert.IsType(t, &Resp{}, opts.ResponseSchema)
}

func TestWithMiddleware(t *testing.T) {
	middleware1 := func(c *fiber.Ctx) error { return c.Next() }
	middleware2 := func(c *fiber.Ctx) error { return c.Next() }

	option := autofiber.WithMiddleware(middleware1, middleware2)
	opts := &autofiber.RouteOptions{}

	option(opts)

	assert.Len(t, opts.Middleware, 2)
}

func TestWithDescription(t *testing.T) {
	description := "Test route description"

	option := autofiber.WithDescription(description)
	opts := &autofiber.RouteOptions{}

	option(opts)

	assert.Equal(t, description, opts.Description)
}

func TestWithTags(t *testing.T) {
	tags := []string{"users", "api", "v1"}

	option := autofiber.WithTags(tags...)
	opts := &autofiber.RouteOptions{}

	option(opts)

	assert.Equal(t, tags, opts.Tags)
}

func TestWithTags_Single(t *testing.T) {
	tag := "users"

	option := autofiber.WithTags(tag)
	opts := &autofiber.RouteOptions{}

	option(opts)

	assert.Equal(t, []string{tag}, opts.Tags)
}
