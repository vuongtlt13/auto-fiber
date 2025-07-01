package autofiber

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type dummyAF struct{ AutoFiber }

func (af *dummyAF) validatorFunc() *dummyAF { return af }

func TestCreateHandlerWithOptions_UnhappyCases(t *testing.T) {
	af := &AutoFiber{}
	opts := &RouteOptions{}

	// 1. Handler is not a function
	assert.Panics(t, func() {
		af.createHandlerWithOptions(123, opts)
	})

	// 2. opts.RequestSchema == nil, handler wrong signature
	assert.Panics(t, func() {
		badHandler := func(c *fiber.Ctx, x int) (interface{}, error) { return nil, nil }
		af.createHandlerWithOptions(badHandler, opts)
	})

	// 3. opts.RequestSchema != nil, handler wrong signature
	opts.RequestSchema = struct{}{}
	assert.Panics(t, func() {
		badHandler := func(c *fiber.Ctx) (interface{}, error) { return nil, nil }
		af.createHandlerWithOptions(badHandler, opts)
	})
}
