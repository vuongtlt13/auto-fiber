package autofiber

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// parseFromMultipleSources parses request from multiple sources based on struct tags
func parseFromMultipleSources(c *fiber.Ctx, req interface{}) error {
	reqValue := reflect.ValueOf(req).Elem()
	reqType := reqValue.Type()

	// Parse body for POST/PUT/PATCH methods
	method := strings.ToUpper(c.Method())
	if method == "POST" || method == "PUT" || method == "PATCH" {
		if err := c.BodyParser(req); err != nil {
			return &ParseError{
				Field:   "body",
				Source:  "body",
				Message: "Invalid request body: " + err.Error(),
			}
		}
	}

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

	// If no parse tag, use auto parsing
	var key string
	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		jsonParts := strings.Split(jsonTag, ",")
		if jsonParts[0] != "" {
			key = jsonParts[0]
		} else {
			key = field.Name
		}
	} else {
		key = field.Name
	}

	required := strings.Contains(field.Tag.Get("validate"), "required")

	return &FieldInfo{
		Source:      Auto,
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
		// Smart parsing: try path first, then query
		if pathValue := c.Params(fieldInfo.Key); pathValue != "" {
			value = pathValue
		} else if queryValue := c.Query(fieldInfo.Key); queryValue != "" {
			value = queryValue
		} else {
			// Body will be handled by middleware BodyParser
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

// Helper functions for parsing
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
