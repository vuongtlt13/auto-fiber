// Package autofiber provides OpenAPI/Swagger documentation configuration and serving utilities.
package autofiber

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// GetOpenAPISpec returns the complete OpenAPI specification as a struct.
// If no documentation info is set, it uses default values.
func (af *AutoFiber) GetOpenAPISpec() *OpenAPISpec {
	return af.docsGenerator.GenerateOpenAPISpec()
}

// GetOpenAPIJSON returns the OpenAPI specification as JSON bytes.
// This is useful for serving the specification via HTTP or saving to a file.
func (af *AutoFiber) GetOpenAPIJSON() ([]byte, error) {
	return af.docsGenerator.GenerateJSON()
}

// ServeDocs serves the OpenAPI specification as JSON at the specified path.
// This creates a GET route that returns the OpenAPI specification.
func (af *AutoFiber) ServeDocs(path string) {
	af.App.Get(path, func(c *fiber.Ctx) error {
		jsonData, err := af.GetOpenAPIJSON()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate documentation"})
		}
		c.Set("Content-Type", "application/json")
		return c.Send(jsonData)
	})
}

// ServeSwaggerUI serves Swagger UI for the OpenAPI documentation.
// This creates a GET route that serves an HTML page with Swagger UI interface.
// swaggerPath is the URL path where Swagger UI will be served.
// docsPath is the URL path where the OpenAPI JSON specification is served.
func (af *AutoFiber) ServeSwaggerUI(swaggerPath, docsPath string) {
	swaggerHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="description" content="SwaggerUI" />
    <title>SwaggerUI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-bundle.js" crossorigin></script>
    <script>
        window.onload = () => {
            window.ui = SwaggerUIBundle({
                url: '%s',
                dom_id: '#swagger-ui',
            });
        };
    </script>
</body>
</html>`, docsPath)

	af.App.Get(swaggerPath, func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})
}
