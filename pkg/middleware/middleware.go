package middleware

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

// ParseSource defines where a field should be parsed from
type ParseSource string

const (
	Body   ParseSource = "body"
	Query  ParseSource = "query"
	Path   ParseSource = "path"
	Header ParseSource = "header"
	Cookie ParseSource = "cookie"
	Form   ParseSource = "form"
	Auto   ParseSource = "auto" // Smart parsing based on HTTP method
)

// FieldInfo contains parsing information for a field
type FieldInfo struct {
	Source      ParseSource
	Key         string // Custom key name (e.g., "user_id" for "userId")
	Required    bool
	Default     interface{}
	Description string
}

// Simple wraps a simple handler
func Simple(handler func(*fiber.Ctx) error) fiber.Handler {
	return handler
}

// AutoParseRequest creates middleware for auto-parsing request from multiple sources
func AutoParseRequest(schema interface{}, customValidator *validator.Validate) fiber.Handler {
	if customValidator == nil {
		customValidator = validate
	}

	return func(c *fiber.Ctx) error {
		// Create a new instance of the schema
		schemaType := reflect.TypeOf(schema)
		if schemaType.Kind() == reflect.Ptr {
			schemaType = schemaType.Elem()
		}

		req := reflect.New(schemaType).Interface()

		// Parse from different sources based on struct tags and HTTP method
		if err := parseFromMultipleSources(c, req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Invalid request",
				"details": err.Error(),
			})
		}

		// Parse body separately for JSON fields
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
		}

		// Validate the request
		if err := customValidator.Struct(req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		}

		// Store parsed request in context
		c.Locals("parsed_request", req)
		return c.Next()
	}
}

// ValidateResponse creates middleware for response validation
func ValidateResponse(schema interface{}, customValidator *validator.Validate) fiber.Handler {
	if customValidator == nil {
		customValidator = validate
	}

	return func(c *fiber.Ctx) error {
		// Store validation info in context
		c.Locals("response_schema", schema)
		c.Locals("response_validator", customValidator)
		return c.Next()
	}
}

// ValidateAndJSON validates response data and returns JSON
func ValidateAndJSON(c *fiber.Ctx, data interface{}) error {
	schema := c.Locals("response_schema")
	validatorInstance := c.Locals("response_validator")

	// If no validation is set up, just return JSON
	if schema == nil || validatorInstance == nil {
		return c.JSON(data)
	}

	// Type assert validator
	if v, ok := validatorInstance.(*validator.Validate); ok {
		// Validate response data
		if err := validateResponseData(data, schema, v); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Response validation failed",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(data)
}

// validateResponseData validates response data against schema
func validateResponseData(data interface{}, schema interface{}, validator *validator.Validate) error {
	// If schema is nil, skip validation
	if schema == nil {
		return nil
	}

	// For simple types, validate directly
	if reflect.TypeOf(data) == reflect.TypeOf(schema) {
		return validator.Struct(data)
	}

	// For complex types, try to validate as struct
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	if dataValue.Kind() == reflect.Struct {
		return validator.Struct(data)
	}

	// For slices, validate each element
	if dataValue.Kind() == reflect.Slice {
		for i := 0; i < dataValue.Len(); i++ {
			element := dataValue.Index(i).Interface()
			if err := validateResponseData(element, schema, validator); err != nil {
				return fmt.Errorf("element %d: %w", i, err)
			}
		}
	}

	return nil
}

// parseFromMultipleSources parses request from multiple sources based on struct tags
func parseFromMultipleSources(c *fiber.Ctx, req interface{}) error {
	reqValue := reflect.ValueOf(req).Elem()
	reqType := reqValue.Type()

	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		fieldValue := reqValue.Field(i)

		// Get parsing information from struct tags
		fieldInfo := getFieldInfo(field, c.Method())
		if fieldInfo == nil {
			continue
		}

		// Parse based on source
		if err := parseFieldFromSource(c, fieldInfo, fieldValue); err != nil {
			return err
		}
	}

	return nil
}

// getFieldInfo extracts parsing information from struct tags with smart defaults
func getFieldInfo(field reflect.StructField, httpMethod string) *FieldInfo {
	// Check for parse tag first (highest priority)
	if parseTag := field.Tag.Get("parse"); parseTag != "" {
		return parseParseTag(parseTag, field)
	}

	// Check for individual source tags (legacy support)
	var source ParseSource
	var key string

	// Determine source from tags
	if field.Tag.Get("query") != "" {
		source = Query
		key = field.Tag.Get("query")
	} else if field.Tag.Get("path") != "" {
		source = Path
		key = field.Tag.Get("path")
	} else if field.Tag.Get("header") != "" {
		source = Header
		key = field.Tag.Get("header")
	} else if field.Tag.Get("cookie") != "" {
		source = Cookie
		key = field.Tag.Get("cookie")
	} else if field.Tag.Get("form") != "" {
		source = Form
		key = field.Tag.Get("form")
	} else if field.Tag.Get("json") != "" {
		// Smart parsing based on HTTP method
		source = getSmartSource(httpMethod)
		key = field.Tag.Get("json")
	} else {
		// Default smart parsing
		source = getSmartSource(httpMethod)
		key = field.Name
	}

	// Extract key name
	if key == "" {
		key = field.Name
	}

	// Check if required
	required := strings.Contains(field.Tag.Get("validate"), "required")

	return &FieldInfo{
		Source:      source,
		Key:         key,
		Required:    required,
		Description: field.Tag.Get("description"),
	}
}

// getSmartSource determines the best source based on HTTP method
func getSmartSource(httpMethod string) ParseSource {
	switch strings.ToUpper(httpMethod) {
	case "GET":
		// For GET requests, prioritize path → query → body
		return Path
	case "POST", "PUT", "PATCH":
		// For mutation requests, prioritize body → path → query
		return Body
	case "DELETE":
		// For DELETE requests, prioritize path → query
		return Path
	default:
		return Body
	}
}

// parseParseTag parses the "parse" tag for complex parsing rules
func parseParseTag(parseTag string, field reflect.StructField) *FieldInfo {
	// Format: parse:"source:key,required,default:value"
	parts := strings.Split(parseTag, ",")

	sourcePart := parts[0]
	sourceKey := strings.Split(sourcePart, ":")

	var source ParseSource
	var key string

	if len(sourceKey) == 2 {
		source = ParseSource(sourceKey[0])
		key = sourceKey[1]
	} else {
		source = ParseSource(sourceKey[0])
		key = field.Name
	}

	required := strings.Contains(parseTag, "required")

	// Extract default value if present
	var defaultValue interface{}
	for _, part := range parts {
		if strings.HasPrefix(part, "default:") {
			defaultStr := strings.TrimPrefix(part, "default:")
			// Convert default string to appropriate type based on field type
			defaultValue = convertDefaultValue(defaultStr, field.Type)
			break
		}
	}

	return &FieldInfo{
		Source:      source,
		Key:         key,
		Required:    required,
		Default:     defaultValue,
		Description: field.Tag.Get("description"),
	}
}

// convertDefaultValue converts string default value to appropriate type
func convertDefaultValue(defaultStr string, fieldType reflect.Type) interface{} {
	switch fieldType.Kind() {
	case reflect.String:
		return defaultStr
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.Atoi(defaultStr); err == nil {
			return val
		}
	case reflect.Bool:
		return defaultStr == "true" || defaultStr == "1"
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(defaultStr, 64); err == nil {
			return val
		}
	}
	return defaultStr
}

// parseFieldFromSource parses a single field from its specified source
func parseFieldFromSource(c *fiber.Ctx, fieldInfo *FieldInfo, fieldValue reflect.Value) error {
	var value interface{}

	switch fieldInfo.Source {
	case Body:
		// Body parsing is handled by the main BodyParser
		return nil // Skip, will be handled separately

	case Query:
		value = c.Query(fieldInfo.Key)

	case Path:
		value = c.Params(fieldInfo.Key)

	case Header:
		value = c.Get(fieldInfo.Key)

	case Cookie:
		value = c.Cookies(fieldInfo.Key)

	case Form:
		value = c.FormValue(fieldInfo.Key)

	case Auto:
		// Smart parsing: try path first, then query, then body
		if pathValue := c.Params(fieldInfo.Key); pathValue != "" {
			value = pathValue
		} else if queryValue := c.Query(fieldInfo.Key); queryValue != "" {
			value = queryValue
		} else {
			// Body will be handled separately
			return nil
		}

	default:
		return nil
	}

	// Handle required fields
	if fieldInfo.Required && (value == "" || value == nil) {
		return &ParseError{
			Field:   fieldInfo.Key,
			Source:  string(fieldInfo.Source),
			Message: "field is required",
		}
	}

	// Set default value if field is empty and has default
	if (value == "" || value == nil) && fieldInfo.Default != nil {
		value = fieldInfo.Default
	}

	// Convert and set the value
	if value != "" && value != nil {
		if err := setFieldValue(fieldValue, value); err != nil {
			return &ParseError{
				Field:   fieldInfo.Key,
				Source:  string(fieldInfo.Source),
				Message: err.Error(),
			}
		}
	}

	return nil
}

// setFieldValue sets a field value with type conversion
func setFieldValue(field reflect.Value, value interface{}) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value.(string))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if str, ok := value.(string); ok {
			// Parse string to int
			if intVal, err := parseInt(str); err == nil {
				field.SetInt(int64(intVal))
			} else {
				return err
			}
		}
	case reflect.Bool:
		if str, ok := value.(string); ok {
			field.SetBool(str == "true" || str == "1")
		}
	case reflect.Float32, reflect.Float64:
		if str, ok := value.(string); ok {
			if floatVal, err := parseFloat(str); err == nil {
				field.SetFloat(floatVal)
			} else {
				return err
			}
		}
	}
	return nil
}

// ParseError represents a parsing error
type ParseError struct {
	Field   string
	Source  string
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s (%s): %s", e.Field, e.Source, e.Message)
}

// Helper functions for parsing
func parseInt(s string) (int, error) {
	// Implementation for string to int conversion
	return strconv.Atoi(s)
}

func parseFloat(s string) (float64, error) {
	// Implementation for string to float conversion
	return strconv.ParseFloat(s, 64)
}

// GetParsedRequest retrieves the parsed request from context
func GetParsedRequest[T any](c *fiber.Ctx) *T {
	if req, ok := c.Locals("parsed_request").(*T); ok {
		return req
	}
	return nil
}
