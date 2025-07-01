// Package autofiber provides route group functionality with automatic request parsing, validation, and documentation generation.
package autofiber

import (
	"github.com/gofiber/fiber/v2"
)

// AutoFiberGroup represents a group of routes with a common prefix and shared middleware.
type AutoFiberGroup struct {
	Group  *fiber.Group // Underlying Fiber group
	app    *AutoFiber   // Reference to the parent AutoFiber app
	Prefix string       // Prefix of the group
}

// Get registers a GET route with automatic request parsing, validation, and documentation generation in the group.
// The handler can be a simple function or a function that accepts a parsed request struct.
// Options can be provided to configure request/response schemas, middleware, and documentation.
func (ag *AutoFiberGroup) Get(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and full path
	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, "GET", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Get(path, handlers...)
	}

	return ag.Group.Get(path, autoHandler)
}

// Post registers a POST route with automatic request parsing, validation, and documentation generation in the group.
// The handler can be a simple function or a function that accepts a parsed request struct.
// Options can be provided to configure request/response schemas, middleware, and documentation.
func (ag *AutoFiberGroup) Post(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and full path
	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, "POST", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Post(path, handlers...)
	}

	return ag.Group.Post(path, autoHandler)
}

// Put registers a PUT route with automatic request parsing, validation, and documentation generation in the group.
// The handler can be a simple function or a function that accepts a parsed request struct.
// Options can be provided to configure request/response schemas, middleware, and documentation.
func (ag *AutoFiberGroup) Put(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and full path
	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, "PUT", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Put(path, handlers...)
	}

	return ag.Group.Put(path, autoHandler)
}

// Delete registers a DELETE route with automatic request parsing, validation, and documentation generation in the group.
// The handler can be a simple function or a function that accepts a parsed request struct.
// Options can be provided to configure request/response schemas, middleware, and documentation.
func (ag *AutoFiberGroup) Delete(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and full path
	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, "DELETE", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Delete(path, handlers...)
	}

	return ag.Group.Delete(path, autoHandler)
}

// Patch registers a PATCH route with automatic request parsing, validation, and documentation generation in the group.
// The handler can be a simple function or a function that accepts a parsed request struct.
// Options can be provided to configure request/response schemas, middleware, and documentation.
func (ag *AutoFiberGroup) Patch(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and full path
	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, "PATCH", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Patch(path, handlers...)
	}

	return ag.Group.Patch(path, autoHandler)
}

// Head registers a HEAD route with automatic request parsing, validation, and documentation generation in the group.
// The handler can be a simple function or a function that accepts a parsed request struct.
// Options can be provided to configure request/response schemas, middleware, and documentation.
func (ag *AutoFiberGroup) Head(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and full path
	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, "HEAD", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Head(path, handlers...)
	}

	return ag.Group.Head(path, autoHandler)
}

// Options registers an OPTIONS route with automatic request parsing, validation, and documentation generation in the group.
// The handler can be a simple function or a function that accepts a parsed request struct.
// Options can be provided to configure request/response schemas, middleware, and documentation.
func (ag *AutoFiberGroup) Options(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and full path
	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, "OPTIONS", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Options(path, handlers...)
	}

	return ag.Group.Options(path, autoHandler)
}

// All registers a route for all HTTP methods with automatic request parsing, validation, and documentation generation in the group.
// The handler can be a simple function or a function that accepts a parsed request struct.
// Options can be provided to configure request/response schemas, middleware, and documentation.
func (ag *AutoFiberGroup) All(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and full path
	fullPath := ag.Prefix + path
	ag.app.docsGenerator.AddRoute(fullPath, "ALL", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.All(path, handlers...)
	}

	return ag.Group.All(path, autoHandler)
}

// Use adds middleware to the group.
// This middleware will be applied to all routes registered in this group.
func (ag *AutoFiberGroup) Use(args ...interface{}) fiber.Router {
	return ag.Group.Use(args...)
}
