// Package autofiber provides route group functionality with automatic request parsing, validation, and documentation generation.
package autofiber

import (
	"github.com/gofiber/fiber/v2"
)

// AutoFiberGroup represents a group of routes with a common prefix and shared middleware.
type AutoFiberGroup struct {
	Group              *fiber.Group    // Underlying Fiber group
	app                *AutoFiber      // Reference to the parent AutoFiber app
	Prefix             string          // Prefix of the group
	groupMiddleware    []fiber.Handler // Middleware applied to every route in the group
	groupRequireJWTAuth bool           // When true, every route in the group requires JWT auth
}

// WithMiddleware adds middleware that will be prepended to every route registered in this group.
// Returns the group for chaining.
//
// Example:
//
//	api := app.Group("/api").WithMiddleware(rateLimitMiddleware, loggingMiddleware)
//	api.Get("/users", handler)
func (ag *AutoFiberGroup) WithMiddleware(middleware ...fiber.Handler) *AutoFiberGroup {
	ag.groupMiddleware = append(ag.groupMiddleware, middleware...)
	return ag
}

// WithJwtAuth marks every route registered in this group as requiring HTTP Bearer (JWT) auth.
// Returns the group for chaining.
//
// Example:
//
//	protected := app.Group("/admin").WithJwtAuth()
//	protected.Get("/dashboard", handler)
func (ag *AutoFiberGroup) WithJwtAuth() *AutoFiberGroup {
	ag.groupRequireJWTAuth = true
	return ag
}

// mergeOpts copies group-level settings into the per-route opts.
func (ag *AutoFiberGroup) mergeOpts(opts *RouteOptions) {
	if ag.groupRequireJWTAuth {
		opts.RequireJWTAuth = true
	}
	if len(ag.groupMiddleware) > 0 {
		// Group middleware runs before route-specific middleware.
		opts.Middleware = append(ag.groupMiddleware, opts.Middleware...)
	}
}

// registerRoute is the shared implementation used by all HTTP-method helpers.
func (ag *AutoFiberGroup) registerRoute(
	method, path string,
	handler interface{},
	options []RouteOption,
	register func(path string, handlers ...fiber.Handler) fiber.Router,
) fiber.Router {
	opts := applyOptions(options)
	ag.mergeOpts(opts)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, method, handler, opts)

	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return register(path, handlers...)
	}
	return register(path, autoHandler)
}

// Get registers a GET route in the group.
func (ag *AutoFiberGroup) Get(path string, handler interface{}, options ...RouteOption) fiber.Router {
	return ag.registerRoute("GET", path, handler, options, ag.Group.Get)
}

// Post registers a POST route in the group.
func (ag *AutoFiberGroup) Post(path string, handler interface{}, options ...RouteOption) fiber.Router {
	return ag.registerRoute("POST", path, handler, options, ag.Group.Post)
}

// Put registers a PUT route in the group.
func (ag *AutoFiberGroup) Put(path string, handler interface{}, options ...RouteOption) fiber.Router {
	return ag.registerRoute("PUT", path, handler, options, ag.Group.Put)
}

// Delete registers a DELETE route in the group.
func (ag *AutoFiberGroup) Delete(path string, handler interface{}, options ...RouteOption) fiber.Router {
	return ag.registerRoute("DELETE", path, handler, options, ag.Group.Delete)
}

// Patch registers a PATCH route in the group.
func (ag *AutoFiberGroup) Patch(path string, handler interface{}, options ...RouteOption) fiber.Router {
	return ag.registerRoute("PATCH", path, handler, options, ag.Group.Patch)
}

// Head registers a HEAD route in the group.
func (ag *AutoFiberGroup) Head(path string, handler interface{}, options ...RouteOption) fiber.Router {
	return ag.registerRoute("HEAD", path, handler, options, ag.Group.Head)
}

// Options registers an OPTIONS route in the group.
func (ag *AutoFiberGroup) Options(path string, handler interface{}, options ...RouteOption) fiber.Router {
	return ag.registerRoute("OPTIONS", path, handler, options, ag.Group.Options)
}

// All registers a route for all HTTP methods in the group.
func (ag *AutoFiberGroup) All(path string, handler interface{}, options ...RouteOption) fiber.Router {
	return ag.registerRoute("ALL", path, handler, options, ag.Group.All)
}

// Use adds middleware to the underlying fiber group (applies to all sub-routes).
func (ag *AutoFiberGroup) Use(args ...interface{}) fiber.Router {
	return ag.Group.Use(args...)
}
