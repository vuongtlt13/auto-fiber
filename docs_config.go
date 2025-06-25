package autofiber

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// WithDocsInfo sets the API documentation information
func (af *AutoFiber) WithDocsInfo(info OpenAPIInfo) *AutoFiber {
	af.docsInfo = &info
	return af
}

// WithDocsServer adds a server to the documentation
func (af *AutoFiber) WithDocsServer(server OpenAPIServer) *AutoFiber {
	af.docsServers = append(af.docsServers, server)
	return af
}

// WithDocsBasePath sets the base path for documentation
func (af *AutoFiber) WithDocsBasePath(basePath string) *AutoFiber {
	af.docsGenerator = NewDocsGenerator(basePath)
	return af
}

// GetOpenAPISpec returns the OpenAPI specification
func (af *AutoFiber) GetOpenAPISpec() *OpenAPISpec {
	if af.docsInfo == nil {
		af.docsInfo = &OpenAPIInfo{
			Title:   "AutoFiber API",
			Version: "1.0.0",
		}
	}

	// Add servers if provided
	if len(af.docsServers) > 0 {
		af.docsGenerator.basePath = af.docsServers[0].URL
	}

	return af.docsGenerator.GenerateOpenAPISpec(*af.docsInfo)
}

// GetOpenAPIJSON returns the OpenAPI specification as JSON
func (af *AutoFiber) GetOpenAPIJSON() ([]byte, error) {
	if af.docsInfo == nil {
		af.docsInfo = &OpenAPIInfo{
			Title:   "AutoFiber API",
			Version: "1.0.0",
		}
	}

	// Add servers if provided
	if len(af.docsServers) > 0 {
		af.docsGenerator.basePath = af.docsServers[0].URL
	}

	return af.docsGenerator.GenerateJSON(*af.docsInfo)
}

// ServeDocs serves the OpenAPI documentation at the specified path
func (af *AutoFiber) ServeDocs(path string) {
	af.Get(path, func(c *fiber.Ctx) error {
		jsonData, err := af.GetOpenAPIJSON()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate documentation"})
		}
		c.Set("Content-Type", "application/json")
		return c.Send(jsonData)
	})
}

// ServeSwaggerUI serves Swagger UI for the OpenAPI documentation
func (af *AutoFiber) ServeSwaggerUI(swaggerPath, docsPath string) {
	// Serve Swagger UI HTML
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

	af.Get(swaggerPath, func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerHTML)
	})
}
