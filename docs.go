// Package autofiber provides OpenAPI 3.0 specification generation for automatic API documentation.
package autofiber

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// OpenAPISpec represents the OpenAPI 3.0 specification structure.
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers,omitempty"`
	Paths      map[string]OpenAPIPath `json:"paths"`
	Components OpenAPIComponents      `json:"components,omitempty"`
	Tags       []OpenAPITag           `json:"tags,omitempty"`
}

// OpenAPIInfo represents the API information including title, description, version, and contact details.
type OpenAPIInfo struct {
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Version     string          `json:"version"`
	Contact     *OpenAPIContact `json:"contact,omitempty"`
	License     *OpenAPILicense `json:"license,omitempty"`
}

// OpenAPIContact represents contact information for the API.
type OpenAPIContact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// OpenAPILicense represents license information for the API.
type OpenAPILicense struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// OpenAPIServer represents server information for the API.
type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// OpenAPIPath represents a path in the API with all supported HTTP methods.
type OpenAPIPath struct {
	Get     *OpenAPIOperation `json:"get,omitempty"`
	Post    *OpenAPIOperation `json:"post,omitempty"`
	Put     *OpenAPIOperation `json:"put,omitempty"`
	Delete  *OpenAPIOperation `json:"delete,omitempty"`
	Patch   *OpenAPIOperation `json:"patch,omitempty"`
	Head    *OpenAPIOperation `json:"head,omitempty"`
	Options *OpenAPIOperation `json:"options,omitempty"`
}

// OpenAPIOperation represents an API operation with parameters, request body, and responses.
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

// OpenAPIParameter represents a parameter (query, path, header, cookie) for an API operation.
type OpenAPIParameter struct {
	Name        string         `json:"name"`
	In          string         `json:"in"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Schema      *OpenAPISchema `json:"schema,omitempty"`
}

// OpenAPIRequestBody represents a request body for an API operation.
type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Required    bool                        `json:"required,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content"`
}

// OpenAPIMediaType represents media type content (e.g., application/json).
type OpenAPIMediaType struct {
	Schema *OpenAPISchema `json:"schema,omitempty"`
}

// OpenAPIResponse represents a response for an API operation.
type OpenAPIResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty"`
}

// OpenAPISchema represents a JSON schema for request/response data structures.
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

// OpenAPIComponents represents reusable components like schemas and security schemes.
type OpenAPIComponents struct {
	Schemas         map[string]OpenAPISchema     `json:"schemas,omitempty"`
	SecuritySchemes map[string]map[string]string `json:"securitySchemes,omitempty"`
}

// OpenAPITag represents a tag for grouping API operations.
type OpenAPITag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// RouteInfo stores information about a route for documentation generation.
type RouteInfo struct {
	Path        string
	Method      string
	Handler     interface{}
	Options     *RouteOptions
	OperationID string
}

// DocsGenerator handles API documentation generation and OpenAPI specification creation.
type DocsGenerator struct {
	routes   []RouteInfo
	schemas  map[string]OpenAPISchema
	tags     map[string]OpenAPITag
	DocsInfo *OpenAPIInfo
}

// NewDocsGenerator creates a new documentation generator with the specified base path.
func NewDocsGenerator() *DocsGenerator {
	return &DocsGenerator{
		routes:  []RouteInfo{},
		schemas: make(map[string]OpenAPISchema),
		tags:    make(map[string]OpenAPITag),
	}
}

// AddRoute adds a route to the documentation generator with its metadata and options.
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

// convertPathToOpenAPIFormat converts Fiber path format (:param) to OpenAPI format ({param})
func convertPathToOpenAPIFormat(path string) string {
	// Replace :param with {param}
	result := path
	for {
		colonIndex := strings.Index(result, ":")
		if colonIndex == -1 {
			break
		}

		// Find the end of the parameter (next / or end of string)
		endIndex := len(result)
		for i := colonIndex + 1; i < len(result); i++ {
			if result[i] == '/' {
				endIndex = i
				break
			}
		}

		// Replace :param with {param}
		paramName := result[colonIndex+1 : endIndex]
		result = result[:colonIndex] + "{" + paramName + "}" + result[endIndex:]
	}

	return result
}

// GenerateOpenAPISpec generates the complete OpenAPI specification from collected route information.
func (dg *DocsGenerator) GenerateOpenAPISpec() *OpenAPISpec {
	// Use stored DocsInfo or default
	info := OpenAPIInfo{
		Title:   "AutoFiber API",
		Version: "1.0.0",
	}
	if dg.DocsInfo != nil {
		info = *dg.DocsInfo
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info:    info,
		Paths:   make(map[string]OpenAPIPath),
		Components: OpenAPIComponents{
			Schemas:         dg.schemas,
			SecuritySchemes: make(map[string]map[string]string),
		},
		Tags: dg.getTagsList(),
	}

	// Track if any route needs bearer auth
	needsBearerAuth := false

	// Generate paths from routes
	for _, route := range dg.routes {
		path, hasBearer := dg.generatePathWithSecurity(route)
		if hasBearer {
			needsBearerAuth = true
		}
		openAPIPath := convertPathToOpenAPIFormat(route.Path)
		spec.Paths[openAPIPath] = path
	}

	// Add bearerAuth security scheme if needed
	if needsBearerAuth {
		spec.Components.SecuritySchemes["bearerAuth"] = map[string]string{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
		}
	}

	return spec
}

// GenerateJSON generates the OpenAPI specification as JSON bytes.
// This is useful for serving the specification via HTTP or saving to a file.
func (dg *DocsGenerator) GenerateJSON() ([]byte, error) {
	spec := dg.GenerateOpenAPISpec()
	return json.MarshalIndent(spec, "", "  ")
}

// generatePathWithSecurity generates a path operation from route info with security considerations.
// It returns the OpenAPIPath and a boolean indicating if bearer authentication is required.
func (dg *DocsGenerator) generatePathWithSecurity(route RouteInfo) (OpenAPIPath, bool) {
	operation := &OpenAPIOperation{
		Tags:        route.Options.Tags,
		Summary:     route.Options.Description,
		Description: route.Options.Description,
		OperationID: route.OperationID,
		Responses:   dg.generateResponses(route),
	}

	hasBearer := false
	// Add parameters and request body based on parse tags
	if route.Options != nil && route.Options.RequestSchema != nil {
		parameters, requestBody, needsBearer := dg.generateParametersAndBodyWithSecurity(route.Options.RequestSchema, route.Path)
		operation.Parameters = parameters
		operation.RequestBody = requestBody
		if needsBearer {
			operation.Security = []map[string][]string{{"bearerAuth": {}}}
			hasBearer = true
		}
	} else {
		// Fallback to path-only parameters
		operation.Parameters = dg.generatePathParameters(route.Path)
	}

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

	return path, hasBearer
}

// generateParametersAndBodyWithSecurity generates parameters and request body from parse tags with security handling.
// It analyzes struct fields and their parse tags to determine parameter sources (query, path, header, cookie, body).
// Returns parameters, request body, and a boolean indicating if bearer authentication is required.
func (dg *DocsGenerator) generateParametersAndBodyWithSecurity(schema interface{}, path string) ([]OpenAPIParameter, *OpenAPIRequestBody, bool) {
	var parameters []OpenAPIParameter
	var bodyFields []string
	var bodySchema OpenAPISchema
	needsBearer := false

	t := reflect.TypeOf(schema)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return parameters, nil, false
	}

	// Add path parameters from URL
	parameters = append(parameters, dg.generatePathParameters(path)...)

	// Track which fields are handled by parse tags
	handledFields := make(map[string]bool)

	// Process fields based on parse tags
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		parseTag := field.Tag.Get("parse")

		// Get field name for JSON
		jsonTag := field.Tag.Get("json")
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		// Convert field type to OpenAPI schema
		fieldSchema := dg.convertFieldTypeToSchema(field.Type)

		// Add description and example
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema.Description = desc
		}
		if example := field.Tag.Get("example"); example != "" {
			fieldSchema.Example = example
		}

		if parseTag == "" {
			// No parse tag - assume it's a body field (matching middleware behavior)
			if bodySchema.Properties == nil {
				bodySchema = OpenAPISchema{
					Type:       "object",
					Properties: make(map[string]OpenAPISchema),
					Required:   []string{},
				}
			}
			bodySchema.Properties[jsonName] = fieldSchema
			bodyFields = append(bodyFields, jsonName)
			handledFields[jsonName] = true
			continue
		}

		// Parse the parse tag
		parts := strings.Split(parseTag, ",")
		sourcePart := parts[0]
		sourceKey := strings.Split(sourcePart, ":")

		if len(sourceKey) != 2 {
			continue
		}

		source := sourceKey[0]
		key := sourceKey[1]

		// Check if required
		required := strings.Contains(parseTag, "required")

		handledFields[jsonName] = true

		switch source {
		case "path":
			// Path parameters are already handled by generatePathParameters
			// Just update existing ones with field info
			for i, param := range parameters {
				if param.Name == key {
					parameters[i].Schema = &fieldSchema
					parameters[i].Description = fieldSchema.Description
					break
				}
			}
		case "query":
			param := OpenAPIParameter{
				Name:        key,
				In:          "query",
				Required:    required,
				Description: fieldSchema.Description,
				Schema:      &fieldSchema,
			}
			parameters = append(parameters, param)
		case "header":
			// Special case: Authorization header -> use security scheme
			if strings.ToLower(key) == "authorization" {
				needsBearer = true
				// Do not add as parameter, will be handled by security
				continue
			} else {
				param := OpenAPIParameter{
					Name:        key,
					In:          "header",
					Required:    required,
					Description: fieldSchema.Description,
					Schema:      &fieldSchema,
				}
				parameters = append(parameters, param)
			}
		case "cookie":
			param := OpenAPIParameter{
				Name:        key,
				In:          "cookie",
				Required:    required,
				Description: fieldSchema.Description,
				Schema:      &fieldSchema,
			}
			parameters = append(parameters, param)
		case "body":
			// Add to body schema
			if bodySchema.Properties == nil {
				bodySchema = OpenAPISchema{
					Type:       "object",
					Properties: make(map[string]OpenAPISchema),
					Required:   []string{},
				}
			}
			bodySchema.Properties[jsonName] = fieldSchema
			if required {
				bodySchema.Required = append(bodySchema.Required, jsonName)
			}
			bodyFields = append(bodyFields, jsonName)
		case "auto":
			// For auto, we need to determine based on HTTP method
			// This is complex, so we'll add to body for now
			if bodySchema.Properties == nil {
				bodySchema = OpenAPISchema{
					Type:       "object",
					Properties: make(map[string]OpenAPISchema),
					Required:   []string{},
				}
			}
			bodySchema.Properties[jsonName] = fieldSchema
			if required {
				bodySchema.Required = append(bodySchema.Required, jsonName)
			}
			bodyFields = append(bodyFields, jsonName)
		}
	}

	// Create request body if there are body fields
	var requestBody *OpenAPIRequestBody
	if len(bodyFields) > 0 {
		requestBody = &OpenAPIRequestBody{
			Required: true,
			Content: map[string]OpenAPIMediaType{
				"application/json": {
					Schema: &bodySchema,
				},
			},
		}
	}

	return parameters, requestBody, needsBearer
}

// generatePathParameters generates parameters for path variables from the URL path.
// It extracts parameters like /users/:id and creates OpenAPI parameter definitions.
func (dg *DocsGenerator) generatePathParameters(path string) []OpenAPIParameter {
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

// generateRequestBody generates request body from schema (legacy method).
// This method creates a request body that references a schema component.
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

// generateResponses generates responses for the operation including success and error responses.
// It creates standard 200, 400, and 500 responses with appropriate schemas.
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

// addSchema adds a schema to the components section of the OpenAPI specification.
// It converts the Go struct to an OpenAPI schema and stores it for reference.
func (dg *DocsGenerator) addSchema(schema interface{}) {
	schemaName := getSchemaName(schema)
	openAPISchema := dg.convertToOpenAPISchema(schema)
	dg.schemas[schemaName] = openAPISchema
}

// convertToOpenAPISchema converts a Go struct to OpenAPI schema.
// It analyzes struct fields, their types, tags, and validation rules to create a complete schema.
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

// convertFieldTypeToSchema converts a Go type to OpenAPI schema.
// It handles basic types, structs, slices, arrays, and pointers with appropriate OpenAPI types.
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

// getTagsList returns the list of tags as a slice for the OpenAPI specification.
func (dg *DocsGenerator) getTagsList() []OpenAPITag {
	var tags []OpenAPITag
	for _, tag := range dg.tags {
		tags = append(tags, tag)
	}
	return tags
}

// generateOperationID generates a unique operation ID for the OpenAPI specification.
// It combines the HTTP method, path, and handler information to create a unique identifier.
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

// getSchemaName gets the name of a schema from its Go type.
// It handles pointer types and returns the underlying type name.
func getSchemaName(schema interface{}) string {
	t := reflect.TypeOf(schema)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
