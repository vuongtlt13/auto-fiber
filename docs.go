package autofiber

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// OpenAPISpec represents the OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers,omitempty"`
	Paths      map[string]OpenAPIPath `json:"paths"`
	Components OpenAPIComponents      `json:"components,omitempty"`
	Tags       []OpenAPITag           `json:"tags,omitempty"`
}

// OpenAPIInfo represents the API information
type OpenAPIInfo struct {
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Version     string          `json:"version"`
	Contact     *OpenAPIContact `json:"contact,omitempty"`
	License     *OpenAPILicense `json:"license,omitempty"`
}

// OpenAPIContact represents contact information
type OpenAPIContact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// OpenAPILicense represents license information
type OpenAPILicense struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// OpenAPIServer represents server information
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// OpenAPIPath represents a path in the API
type OpenAPIPath struct {
	Get     *OpenAPIOperation `json:"get,omitempty"`
	Post    *OpenAPIOperation `json:"post,omitempty"`
	Put     *OpenAPIOperation `json:"put,omitempty"`
	Delete  *OpenAPIOperation `json:"delete,omitempty"`
	Patch   *OpenAPIOperation `json:"patch,omitempty"`
	Head    *OpenAPIOperation `json:"head,omitempty"`
	Options *OpenAPIOperation `json:"options,omitempty"`
}

// OpenAPIOperation represents an API operation
type OpenAPIOperation struct {
	Tags        []string                   `json:"tags,omitempty"`
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	OperationID string                     `json:"operationId,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
	Security    []map[string][]string      `json:"security,omitempty"`
}

// OpenAPIParameter represents a parameter
type OpenAPIParameter struct {
	Name        string         `json:"name"`
	In          string         `json:"in"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Schema      *OpenAPISchema `json:"schema,omitempty"`
}

// OpenAPIRequestBody represents a request body
type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Required    bool                        `json:"required,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content"`
}

// OpenAPIMediaType represents media type content
type OpenAPIMediaType struct {
	Schema *OpenAPISchema `json:"schema,omitempty"`
}

// OpenAPIResponse represents a response
type OpenAPIResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty"`
}

// OpenAPISchema represents a schema
type OpenAPISchema struct {
	Type        string                   `json:"type,omitempty"`
	Format      string                   `json:"format,omitempty"`
	Description string                   `json:"description,omitempty"`
	Required    []string                 `json:"required,omitempty"`
	Properties  map[string]OpenAPISchema `json:"properties,omitempty"`
	Items       *OpenAPISchema           `json:"items,omitempty"`
	Ref         string                   `json:"$ref,omitempty"`
	Example     interface{}              `json:"example,omitempty"`
}

// OpenAPIComponents represents reusable components
type OpenAPIComponents struct {
	Schemas map[string]OpenAPISchema `json:"schemas,omitempty"`
}

// OpenAPITag represents a tag
type OpenAPITag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// RouteInfo stores information about a route for documentation
type RouteInfo struct {
	Path        string
	Method      string
	Handler     interface{}
	Options     *RouteOptions
	OperationID string
}

// DocsGenerator handles API documentation generation
type DocsGenerator struct {
	routes   []RouteInfo
	schemas  map[string]OpenAPISchema
	tags     map[string]OpenAPITag
	basePath string
}

// NewDocsGenerator creates a new documentation generator
func NewDocsGenerator(basePath string) *DocsGenerator {
	return &DocsGenerator{
		routes:   []RouteInfo{},
		schemas:  make(map[string]OpenAPISchema),
		tags:     make(map[string]OpenAPITag),
		basePath: basePath,
	}
}

// AddRoute adds a route to the documentation generator
func (dg *DocsGenerator) AddRoute(path, method string, handler interface{}, options *RouteOptions) {
	operationID := generateOperationID(method, path, handler)

	routeInfo := RouteInfo{
		Path:        path,
		Method:      method,
		Handler:     handler,
		Options:     options,
		OperationID: operationID,
	}

	dg.routes = append(dg.routes, routeInfo)

	// Add schemas if provided
	if options != nil {
		if options.RequestSchema != nil {
			dg.addSchema(options.RequestSchema)
		}
		if options.ResponseSchema != nil {
			dg.addSchema(options.ResponseSchema)
		}

		// Add tags
		for _, tag := range options.Tags {
			dg.tags[tag] = OpenAPITag{Name: tag}
		}
	}
}

// GenerateOpenAPISpec generates the OpenAPI specification
func (dg *DocsGenerator) GenerateOpenAPISpec(info OpenAPIInfo) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info:    info,
		Paths:   make(map[string]OpenAPIPath),
		Components: OpenAPIComponents{
			Schemas: dg.schemas,
		},
		Tags: dg.getTagsList(),
	}

	// Add default server if basePath is provided
	if dg.basePath != "" {
		spec.Servers = []OpenAPIServer{
			{URL: dg.basePath, Description: "Default server"},
		}
	}

	// Generate paths from routes
	for _, route := range dg.routes {
		path := dg.generatePath(route)
		spec.Paths[route.Path] = path
	}

	return spec
}

// GenerateJSON generates the OpenAPI specification as JSON
func (dg *DocsGenerator) GenerateJSON(info OpenAPIInfo) ([]byte, error) {
	spec := dg.GenerateOpenAPISpec(info)
	return json.MarshalIndent(spec, "", "  ")
}

// generatePath generates an OpenAPI path from route info
func (dg *DocsGenerator) generatePath(route RouteInfo) OpenAPIPath {
	operation := &OpenAPIOperation{
		Tags:        route.Options.Tags,
		Summary:     route.Options.Description,
		Description: route.Options.Description,
		OperationID: route.OperationID,
		Responses:   dg.generateResponses(route),
	}

	// Add request body if there's a request schema
	if route.Options != nil && route.Options.RequestSchema != nil {
		operation.RequestBody = dg.generateRequestBody(route.Options.RequestSchema)
	}

	// Add parameters for path variables
	operation.Parameters = dg.generateParameters(route.Path)

	path := OpenAPIPath{}
	switch strings.ToUpper(route.Method) {
	case "GET":
		path.Get = operation
	case "POST":
		path.Post = operation
	case "PUT":
		path.Put = operation
	case "DELETE":
		path.Delete = operation
	case "PATCH":
		path.Patch = operation
	case "HEAD":
		path.Head = operation
	case "OPTIONS":
		path.Options = operation
	}

	return path
}

// generateParameters generates parameters for path variables
func (dg *DocsGenerator) generateParameters(path string) []OpenAPIParameter {
	var params []OpenAPIParameter

	// Extract path parameters (e.g., /users/:id)
	segments := strings.Split(path, "/")
	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			paramName := strings.TrimPrefix(segment, ":")
			param := OpenAPIParameter{
				Name:        paramName,
				In:          "path",
				Required:    true,
				Description: fmt.Sprintf("Path parameter: %s", paramName),
				Schema: &OpenAPISchema{
					Type: "string",
				},
			}
			params = append(params, param)
		}
	}

	return params
}

// generateRequestBody generates request body from schema
func (dg *DocsGenerator) generateRequestBody(schema interface{}) *OpenAPIRequestBody {
	schemaName := getSchemaName(schema)

	return &OpenAPIRequestBody{
		Required: true,
		Content: map[string]OpenAPIMediaType{
			"application/json": {
				Schema: &OpenAPISchema{
					Ref: fmt.Sprintf("#/components/schemas/%s", schemaName),
				},
			},
		},
	}
}

// generateResponses generates responses for the operation
func (dg *DocsGenerator) generateResponses(route RouteInfo) map[string]OpenAPIResponse {
	responses := make(map[string]OpenAPIResponse)

	// Default success response
	successResponse := OpenAPIResponse{
		Description: "Successful operation",
	}

	// Add response schema if provided
	if route.Options != nil && route.Options.ResponseSchema != nil {
		schemaName := getSchemaName(route.Options.ResponseSchema)
		successResponse.Content = map[string]OpenAPIMediaType{
			"application/json": {
				Schema: &OpenAPISchema{
					Ref: fmt.Sprintf("#/components/schemas/%s", schemaName),
				},
			},
		}
	}

	responses["200"] = successResponse

	// Add common error responses
	responses["400"] = OpenAPIResponse{
		Description: "Bad Request",
		Content: map[string]OpenAPIMediaType{
			"application/json": {
				Schema: &OpenAPISchema{
					Type: "object",
					Properties: map[string]OpenAPISchema{
						"error":   {Type: "string"},
						"details": {Type: "object"},
					},
				},
			},
		},
	}

	responses["500"] = OpenAPIResponse{
		Description: "Internal Server Error",
		Content: map[string]OpenAPIMediaType{
			"application/json": {
				Schema: &OpenAPISchema{
					Type: "object",
					Properties: map[string]OpenAPISchema{
						"error": {Type: "string"},
					},
				},
			},
		},
	}

	return responses
}

// addSchema adds a schema to the components
func (dg *DocsGenerator) addSchema(schema interface{}) {
	schemaName := getSchemaName(schema)
	openAPISchema := dg.convertToOpenAPISchema(schema)
	dg.schemas[schemaName] = openAPISchema
}

// convertToOpenAPISchema converts a Go struct to OpenAPI schema
func (dg *DocsGenerator) convertToOpenAPISchema(schema interface{}) OpenAPISchema {
	t := reflect.TypeOf(schema)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return OpenAPISchema{Type: "object"}
	}

	openAPISchema := OpenAPISchema{
		Type:       "object",
		Properties: make(map[string]OpenAPISchema),
		Required:   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		// Get validation tags
		validateTag := field.Tag.Get("validate")
		isRequired := strings.Contains(validateTag, "required")

		// Convert field type to OpenAPI schema
		fieldSchema := dg.convertFieldTypeToSchema(field.Type)

		// Add description from struct tag
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema.Description = desc
		}

		// Add example from struct tag
		if example := field.Tag.Get("example"); example != "" {
			fieldSchema.Example = example
		}

		openAPISchema.Properties[jsonName] = fieldSchema

		if isRequired {
			openAPISchema.Required = append(openAPISchema.Required, jsonName)
		}
	}

	return openAPISchema
}

// convertFieldTypeToSchema converts a Go type to OpenAPI schema
func (dg *DocsGenerator) convertFieldTypeToSchema(t reflect.Type) OpenAPISchema {
	switch t.Kind() {
	case reflect.String:
		return OpenAPISchema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return OpenAPISchema{Type: "integer"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return OpenAPISchema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return OpenAPISchema{Type: "number"}
	case reflect.Bool:
		return OpenAPISchema{Type: "boolean"}
	case reflect.Struct:
		// Handle time.Time
		if t == reflect.TypeOf(time.Time{}) {
			return OpenAPISchema{Type: "string", Format: "date-time"}
		}
		// Handle other structs
		schemaName := getSchemaName(reflect.New(t).Interface())
		return OpenAPISchema{Ref: fmt.Sprintf("#/components/schemas/%s", schemaName)}
	case reflect.Slice, reflect.Array:
		itemSchema := dg.convertFieldTypeToSchema(t.Elem())
		return OpenAPISchema{
			Type:  "array",
			Items: &itemSchema,
		}
	case reflect.Ptr:
		return dg.convertFieldTypeToSchema(t.Elem())
	default:
		return OpenAPISchema{Type: "string"}
	}
}

// getTagsList returns the list of tags
func (dg *DocsGenerator) getTagsList() []OpenAPITag {
	var tags []OpenAPITag
	for _, tag := range dg.tags {
		tags = append(tags, tag)
	}
	return tags
}

// generateOperationID generates a unique operation ID
func generateOperationID(method, path string, handler interface{}) string {
	handlerName := reflect.TypeOf(handler).String()
	handlerName = strings.TrimPrefix(handlerName, "func(")
	handlerName = strings.Split(handlerName, "(")[0]

	// Clean up the path for operation ID
	cleanPath := strings.ReplaceAll(path, "/", "_")
	cleanPath = strings.ReplaceAll(cleanPath, ":", "")
	cleanPath = strings.TrimPrefix(cleanPath, "_")

	return fmt.Sprintf("%s_%s_%s", strings.ToLower(method), cleanPath, handlerName)
}

// getSchemaName gets the name of a schema
func getSchemaName(schema interface{}) string {
	t := reflect.TypeOf(schema)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// WithDocsInfo sets documentation information for the API
func WithDocsInfo(info OpenAPIInfo) RouteOption {
	return func(opts *RouteOptions) {
		// This is a global option, not per-route
		// We'll handle this in the main AutoFiber struct
	}
}

// WithDocsServer adds a server to the documentation
func WithDocsServer(server OpenAPIServer) RouteOption {
	return func(opts *RouteOptions) {
		// This is a global option, not per-route
		// We'll handle this in the main AutoFiber struct
	}
}
