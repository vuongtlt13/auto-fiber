package autofiber

import (
	"github.com/gofiber/fiber/v2"
)

// Get registers a GET route with options
func (af *AutoFiber) Get(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := af.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and path
	af.docsGenerator.AddRoute(path, "GET", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return af.App.Get(path, handlers...)
	}

	return af.App.Get(path, autoHandler)
}

// Post registers a POST route with options
func (af *AutoFiber) Post(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := af.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and path
	af.docsGenerator.AddRoute(path, "POST", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return af.App.Post(path, handlers...)
	}

	return af.App.Post(path, autoHandler)
}

// Put registers a PUT route with options
func (af *AutoFiber) Put(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := af.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and path
	af.docsGenerator.AddRoute(path, "PUT", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return af.App.Put(path, handlers...)
	}

	return af.App.Put(path, autoHandler)
}

// Delete registers a DELETE route with options
func (af *AutoFiber) Delete(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := af.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and path
	af.docsGenerator.AddRoute(path, "DELETE", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return af.App.Delete(path, handlers...)
	}

	return af.App.Delete(path, autoHandler)
}

// Patch registers a PATCH route with options
func (af *AutoFiber) Patch(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := af.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and path
	af.docsGenerator.AddRoute(path, "PATCH", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return af.App.Patch(path, handlers...)
	}

	return af.App.Patch(path, autoHandler)
}

// Head registers a HEAD route with options
func (af *AutoFiber) Head(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := af.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and path
	af.docsGenerator.AddRoute(path, "HEAD", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return af.App.Head(path, handlers...)
	}

	return af.App.Head(path, autoHandler)
}

// Options registers an OPTIONS route with options
func (af *AutoFiber) Options(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := af.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and path
	af.docsGenerator.AddRoute(path, "OPTIONS", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return af.App.Options(path, handlers...)
	}

	return af.App.Options(path, autoHandler)
}

// All registers a route for all HTTP methods with options
func (af *AutoFiber) All(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := af.createHandlerWithOptions(handler, opts)

	// Add route to docs generator with correct method and path
	af.docsGenerator.AddRoute(path, "ALL", handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return af.App.All(path, handlers...)
	}

	return af.App.All(path, autoHandler)
}
