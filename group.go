package autofiber

import (
	"github.com/gofiber/fiber/v2"
)

// Get registers a GET route with options in the group
func (ag *AutoFiberGroup) Get(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Get(path, handlers...)
	}

	return ag.Group.Get(path, autoHandler)
}

// Post registers a POST route with options in the group
func (ag *AutoFiberGroup) Post(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Post(path, handlers...)
	}

	return ag.Group.Post(path, autoHandler)
}

// Put registers a PUT route with options in the group
func (ag *AutoFiberGroup) Put(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Put(path, handlers...)
	}

	return ag.Group.Put(path, autoHandler)
}

// Delete registers a DELETE route with options in the group
func (ag *AutoFiberGroup) Delete(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Delete(path, handlers...)
	}

	return ag.Group.Delete(path, autoHandler)
}

// Patch registers a PATCH route with options in the group
func (ag *AutoFiberGroup) Patch(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Patch(path, handlers...)
	}

	return ag.Group.Patch(path, autoHandler)
}

// Head registers a HEAD route with options in the group
func (ag *AutoFiberGroup) Head(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Head(path, handlers...)
	}

	return ag.Group.Head(path, autoHandler)
}

// Options registers an OPTIONS route with options in the group
func (ag *AutoFiberGroup) Options(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.Options(path, handlers...)
	}

	return ag.Group.Options(path, autoHandler)
}

// All registers a route for all HTTP methods with options in the group
func (ag *AutoFiberGroup) All(path string, handler interface{}, options ...RouteOption) fiber.Router {
	opts := applyOptions(options)
	autoHandler := ag.app.createHandlerWithOptions(handler, opts)

	// Apply middleware
	if len(opts.Middleware) > 0 {
		handlers := append(opts.Middleware, autoHandler)
		return ag.Group.All(path, handlers...)
	}

	return ag.Group.All(path, autoHandler)
}

// Use adds middleware to the group
func (ag *AutoFiberGroup) Use(args ...interface{}) fiber.Router {
	return ag.Group.Use(args...)
}
