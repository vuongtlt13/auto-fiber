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
	operationID := GenerateOperationID(method, path, handler)

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
		// Only allow requestBody for POST, PUT, PATCH
		methodUpper := strings.ToUpper(route.Method)
		if requestBody != nil && (methodUpper == "POST" || methodUpper == "PUT" || methodUpper == "PATCH") {
			operation.RequestBody = requestBody
		}
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

		// Get validation tags
		validateTag := field.Tag.Get("validate")
		isRequired := strings.Contains(validateTag, "required")

		handledFields[jsonName] = true

		// Parse the parseTag for source and key
		var source, key string
		if parseTag != "" {
			parts := strings.Split(parseTag, ",")
			sourcePart := parts[0]
			sourceKey := strings.SplitN(sourcePart, ":", 2)
			if len(sourceKey) == 2 {
				source = sourceKey[0]
				key = sourceKey[1]
			} else {
				source = sourceKey[0]
				key = jsonName
			}
		}

		switch source {
		case "path":
			for j, param := range parameters {
				if param.Name == key {
					fieldSchema := dg.convertFieldTypeToSchema(field.Type)
					parameters[j].Schema = &fieldSchema
					parameters[j].Description = field.Tag.Get("description")
					break
				}
			}
		case "query":
			fieldSchema := dg.convertFieldTypeToSchema(field.Type)
			param := OpenAPIParameter{
				Name:        key,
				In:          "query",
				Required:    isRequired,
				Description: field.Tag.Get("description"),
				Schema:      &fieldSchema,
			}
			parameters = append(parameters, param)
		case "header":
			if strings.ToLower(key) == "authorization" {
				needsBearer = true
				continue
			} else {
				fieldSchema := dg.convertFieldTypeToSchema(field.Type)
				param := OpenAPIParameter{
					Name:        key,
					In:          "header",
					Required:    isRequired,
					Description: field.Tag.Get("description"),
					Schema:      &fieldSchema,
				}
				parameters = append(parameters, param)
			}
		case "cookie":
			fieldSchema := dg.convertFieldTypeToSchema(field.Type)
			param := OpenAPIParameter{
				Name:        key,
				In:          "cookie",
				Required:    isRequired,
				Description: field.Tag.Get("description"),
				Schema:      &fieldSchema,
			}
			parameters = append(parameters, param)
		case "body":
			if bodySchema.Properties == nil {
				bodySchema = OpenAPISchema{
					Type:       "object",
					Properties: make(map[string]OpenAPISchema),
					Required:   []string{},
				}
			}
			bodySchema.Properties[key] = dg.convertFieldTypeToSchema(field.Type)
			if isRequired {
				bodySchema.Required = append(bodySchema.Required, key)
			}
			bodyFields = append(bodyFields, key)
		case "auto":
			if bodySchema.Properties == nil {
				bodySchema = OpenAPISchema{
					Type:       "object",
					Properties: make(map[string]OpenAPISchema),
					Required:   []string{},
				}
			}
			bodySchema.Properties[key] = dg.convertFieldTypeToSchema(field.Type)
			if isRequired {
				bodySchema.Required = append(bodySchema.Required, key)
			}
			bodyFields = append(bodyFields, key)
		}
	}

	// Create request body if there are body fields
	var requestBody *OpenAPIRequestBody
	if len(bodyFields) > 0 {
		// Register the schema as a component and use $ref
		dg.addSchema(schema)
		schemaName := GetSchemaName(schema)
		requestBody = &OpenAPIRequestBody{
			Required: true,
			Content: map[string]OpenAPIMediaType{
				"application/json": {
					Schema: &OpenAPISchema{
						Ref: "#/components/schemas/" + schemaName,
					},
				},
			},
		}
	} else {
		// Fallback: if POST/PUT/PATCH and has request schema, use ConvertRequestToOpenAPISchema
		method := strings.ToUpper(guessMethodFromPath(path))
		if (method == "POST" || method == "PUT" || method == "PATCH") && t.Kind() == reflect.Struct {
			schemaObj := dg.ConvertRequestToOpenAPISchema(schema)
			requestBody = &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMediaType{
					"application/json": {
						Schema: &schemaObj,
					},
				},
			}
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
		schemaName := GetSchemaName(route.Options.ResponseSchema)
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
	schemaName := GetSchemaName(schema)
	// Avoid duplicate registration
	if _, exists := dg.schemas[schemaName]; exists {
		return
	}
	openAPISchema := dg.ConvertRequestToOpenAPISchema(schema)
	dg.schemas[schemaName] = openAPISchema
}

// ConvertToOpenAPISchema converts a Go struct to OpenAPI schema.
// It analyzes struct fields, their types, tags, and validation rules to create a complete schema.
// This is a legacy function that defaults to request conversion behavior.
func (dg *DocsGenerator) ConvertToOpenAPISchema(schema interface{}) OpenAPISchema {
	return dg.ConvertRequestToOpenAPISchema(schema)
}

// ConvertRequestToOpenAPISchema converts a Go struct to OpenAPI schema for request parsing.
// It prioritizes parse tags for field names, falls back to json tags, and handles validation.
func (dg *DocsGenerator) ConvertRequestToOpenAPISchema(schema interface{}) OpenAPISchema {
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

		parseTag := field.Tag.Get("parse")
		jsonTag := field.Tag.Get("json")

		var fieldName string
		include := false

		if parseTag != "" && parseTag != "-" {
			// parse tag format: body:data, query:name, etc.
			parts := strings.SplitN(parseTag, ",", 2)
			sourceKey := strings.SplitN(parts[0], ":", 2)
			if len(sourceKey) == 2 {
				source := sourceKey[0]
				key := sourceKey[1]
				if source == "body" {
					fieldName = key
					include = true
				}
			}
		} else if jsonTag != "" && jsonTag != "-" {
			// If there is no parse tag, only include if there is a valid json tag
			fieldName = strings.Split(jsonTag, ",")[0]
			if fieldName == "" {
				// If json tag is empty or only comma, skip this field
				continue
			}
			include = true
		}

		if !include {
			continue
		}

		validateTag := field.Tag.Get("validate")
		isRequired := strings.Contains(validateTag, "required")

		// Recursively register struct field types (except time.Time)
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			dg.addSchema(reflect.New(field.Type).Interface())
		}
		// Also register pointer to struct types
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && field.Type.Elem() != reflect.TypeOf(time.Time{}) {
			dg.addSchema(reflect.New(field.Type.Elem()).Interface())
		}

		fieldSchema := dg.convertFieldTypeToSchema(field.Type)

		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema.Description = desc
		}
		if example := field.Tag.Get("example"); example != "" {
			fieldSchema.Example = example
		}

		openAPISchema.Properties[fieldName] = fieldSchema

		if isRequired {
			openAPISchema.Required = append(openAPISchema.Required, fieldName)
		}
	}

	return openAPISchema
}

// ConvertResponseToOpenAPISchema converts a Go struct to OpenAPI schema for response serialization.
// It uses json tags for field names, falls back to camelCase field names, and handles validation.
func (dg *DocsGenerator) ConvertResponseToOpenAPISchema(schema interface{}) OpenAPISchema {
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

		// Get json tag
		jsonTag := field.Tag.Get("json")

		// Skip fields with json:"-" tag
		if jsonTag == "-" {
			continue
		}

		// Determine field name: json tag takes priority, then camelCase field name
		var fieldName string
		if jsonTag != "" && jsonTag != "-" {
			fieldName = strings.Split(jsonTag, ",")[0]
			if fieldName == "" {
				fieldName = toCamelCase(field.Name)
			}
		} else {
			// Convert field name to camelCase
			fieldName = toCamelCase(field.Name)
		}

		// Get validation tags
		validateTag := field.Tag.Get("validate")
		isRequired := strings.Contains(validateTag, "required")

		// Recursively register struct field types (except time.Time)
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			dg.addSchema(reflect.New(field.Type).Interface())
		}
		// Also register pointer to struct types
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && field.Type.Elem() != reflect.TypeOf(time.Time{}) {
			dg.addSchema(reflect.New(field.Type.Elem()).Interface())
		}

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

		openAPISchema.Properties[fieldName] = fieldSchema

		if isRequired {
			openAPISchema.Required = append(openAPISchema.Required, fieldName)
		}
	}

	return openAPISchema
}

// toCamelCase converts a field name to camelCase
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	// Handle common abbreviations and special cases
	if len(s) <= 2 {
		return strings.ToLower(s)
	}

	// Handle common abbreviations
	abbreviations := []string{"API", "HTTP", "JSON", "URL", "ID", "SQL", "XML", "HTML", "CSS", "JS"}
	for _, abbr := range abbreviations {
		if strings.HasPrefix(s, abbr) {
			// Convert abbreviation to lowercase and keep the rest as is
			return strings.ToLower(abbr) + s[len(abbr):]
		}
	}

	// Convert first character to lowercase
	result := strings.ToLower(s[:1]) + s[1:]
	return result
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
		// For generic instantiations, inline the schema
		schemaName := GetSchemaName(reflect.New(t).Interface())
		if strings.Contains(schemaName, "_") { // crude check for generic
			return dg.ConvertRequestToOpenAPISchema(reflect.New(t).Elem().Interface())
		}
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

// GetSchemaName gets the name of a schema from its Go type.
// It handles pointer types and returns the underlying type name.
// For generic structs, it generates a unique name including type parameters (e.g., APIResponse_User).
func GetSchemaName(schema interface{}) string {
	t := reflect.TypeOf(schema)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := t.Name()
	// If the type name contains a generic instantiation (e.g., APIResponse[...]), strip the generic part
	isGeneric := false
	if idx := strings.Index(name, "["); idx != -1 {
		name = name[:idx]
		isGeneric = true
	}
	// Only append type names for generic structs
	if isGeneric && t.NumField() > 0 {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldType := field.Type
			for fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct && fieldType != reflect.TypeOf(time.Time{}) {
				baseName := fieldType.Name()
				if baseName == "" {
					s := fieldType.String()
					if idx := strings.LastIndex(s, "."); idx != -1 {
						s = s[idx+1:]
					}
					baseName = s
				}
				name += "_" + baseName
			}
		}
	}
	// Replace any non-alphanumeric or underscore character with underscore
	var sanitized []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			sanitized = append(sanitized, r)
		} else {
			sanitized = append(sanitized, '_')
		}
	}
	return string(sanitized)
}

// GenerateOperationID generates a unique operation ID for the OpenAPI specification.
// It combines the HTTP method and path to create a unique identifier (no handler signature).
func GenerateOperationID(method, path string, handler interface{}) string {
	// Clean up the path for operation ID
	cleanPath := strings.ReplaceAll(path, "/", "_")
	cleanPath = strings.ReplaceAll(cleanPath, ":", "")
	cleanPath = strings.TrimPrefix(cleanPath, "_")

	return fmt.Sprintf("%s_%s", strings.ToLower(method), cleanPath)
}

// Helper: guess method from path (currently always returns POST as method info is not in path)
func guessMethodFromPath(path string) string {
	// Không có thông tin method trong path, trả về POST mặc định
	return "POST"
}

func (dg *DocsGenerator) Schemas() map[string]OpenAPISchema {
	return dg.schemas
}
