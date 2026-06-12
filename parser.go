// Package autofiber provides request parsing utilities for extracting and validating data from multiple sources.
package autofiber

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// schemaMetaCache stores pre-computed field metadata per schema type, keyed by reflect.Type.
// Built once at registration time; reads are lock-free on the hot path.
var schemaMetaCache sync.Map // map[reflect.Type]*cachedSchemaMeta

// cachedSchemaMeta holds pre-computed metadata for a schema type.
type cachedSchemaMeta struct {
	hasBodyFields bool
	fields        []cachedField
}

// cachedField stores the index and pre-parsed FieldInfo for a single struct field.
// When embedded is non-nil the field is an anonymous embedded struct; info is nil.
type cachedField struct {
	index    int
	info     *FieldInfo   // nil when embedded != nil
	embedded reflect.Type // elem type of the embedded struct (non-nil = embedded field)
	embIsPtr bool         // true when the field is declared as a pointer (*EmbeddedType)
}

// getOrCacheSchemaMeta returns (and lazily builds) the cached metadata for t.
func getOrCacheSchemaMeta(t reflect.Type) *cachedSchemaMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if v, ok := schemaMetaCache.Load(t); ok {
		return v.(*cachedSchemaMeta)
	}
	meta := buildSchemaMeta(t)
	schemaMetaCache.Store(t, meta)
	return meta
}

// buildSchemaMeta computes the metadata for t. Panics on invalid parse tags so
// callers (AutoParseRequest) catch programmer errors at registration time.
func buildSchemaMeta(t reflect.Type) *cachedSchemaMeta {
	meta := &cachedSchemaMeta{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft := f.Type
		embIsPtr := false
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
			embIsPtr = true
		}

		// Embedded anonymous struct — recurse and propagate hasBodyFields.
		if f.Anonymous && ft.Kind() == reflect.Struct && ft != reflect.TypeOf(time.Time{}) {
			embMeta := getOrCacheSchemaMeta(ft)
			if embMeta.hasBodyFields {
				meta.hasBodyFields = true
			}
			meta.fields = append(meta.fields, cachedField{
				index:    i,
				embedded: ft,
				embIsPtr: embIsPtr,
			})
			continue
		}

		// Check for explicit body source before building FieldInfo.
		if pt := f.Tag.Get("parse"); pt != "" {
			parts := strings.Split(pt, ",")
			src := strings.SplitN(parts[0], ":", 2)[0]
			if src == string(Body) {
				meta.hasBodyFields = true
			}
		}

		meta.fields = append(meta.fields, cachedField{
			index: i,
			info:  computeFieldInfo(f),
		})
	}
	return meta
}

// parseFromMultipleSources parses request data from multiple sources (body, query, path, header, cookie, form)
// based on struct tags and HTTP method. It fills the req struct with parsed values and returns an error if parsing fails.
func parseFromMultipleSources(c *fiber.Ctx, req interface{}) error {
	reqValue := reflect.ValueOf(req).Elem()
	reqType := reqValue.Type()

	meta := getOrCacheSchemaMeta(reqType)

	// Parse body for POST/PUT/PATCH methods or when schema has explicit body fields.
	method := strings.ToUpper(c.Method())
	if method == "POST" || method == "PUT" || method == "PATCH" || (meta.hasBodyFields && len(c.Body()) > 0) {
		contentType := c.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			if len(c.Body()) == 0 {
				return &ParseError{
					Field:   "body",
					Source:  "body",
					Message: "Request body is required for JSON requests",
				}
			}
			if err := c.BodyParser(req); err != nil {
				return &ParseError{
					Field:   "body",
					Source:  "body",
					Message: "Invalid request body: " + err.Error(),
				}
			}
		} else if len(c.Body()) > 0 {
			if err := c.BodyParser(req); err != nil {
				return &ParseError{
					Field:   "body",
					Source:  "body",
					Message: "Invalid request body: " + err.Error(),
				}
			}
		}
	}

	for _, cf := range meta.fields {
		fieldValue := reqValue.Field(cf.index)

		// Embedded anonymous struct — recurse using its own cached metadata.
		if cf.embedded != nil {
			if cf.embIsPtr {
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(cf.embedded))
				}
				if err := parseFromMultipleSources(c, fieldValue.Interface()); err != nil {
					return fmt.Errorf("embedded field %s: %w", reqType.Field(cf.index).Name, err)
				}
			} else if fieldValue.CanAddr() {
				if err := parseFromMultipleSources(c, fieldValue.Addr().Interface()); err != nil {
					return fmt.Errorf("embedded field %s: %w", reqType.Field(cf.index).Name, err)
				}
			}
			continue
		}

		if cf.info == nil {
			continue
		}

		if err := parseFieldFromSource(c, cf.info, fieldValue); err != nil {
			return err
		}
	}

	return nil
}

// computeFieldInfo extracts parsing information from struct tags with smart defaults.
// Called once per field at schema-registration time, not per request.
func computeFieldInfo(field reflect.StructField) *FieldInfo {
	if parseTag := field.Tag.Get("parse"); parseTag != "" {
		return parseParseTag(parseTag, field)
	}

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

// getSmartSource determines the best source for a field based on HTTP method.
// For GET: path → query → body; for POST/PUT/PATCH: body → path → query; for DELETE: path → query.
func getSmartSource(httpMethod string) ParseSource {
	switch strings.ToUpper(httpMethod) {
	case "GET":
		return Path
	case "POST", "PUT", "PATCH":
		return Body
	case "DELETE":
		return Path
	default:
		return Body
	}
}

// parseParseTag parses the "parse" struct tag for complex parsing rules.
// The tag format is: parse:"source:key,required,default:value"
// Panics on unknown sources so programmer errors surface at registration time.
func parseParseTag(parseTag string, field reflect.StructField) *FieldInfo {
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

	// Validate source at registration time — panics catch typos immediately.
	switch source {
	case Body, Query, Path, Header, Cookie, Form, Auto:
		// valid
	default:
		panic(fmt.Sprintf(
			"autofiber: invalid parse source %q on field %q — must be one of: body, query, path, header, cookie, form, auto",
			source, field.Name,
		))
	}

	required := strings.Contains(parseTag, "required")

	var defaultValue interface{}
	for _, part := range parts {
		if strings.HasPrefix(part, "default:") {
			defaultStr := strings.TrimPrefix(part, "default:")
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

// convertDefaultValue converts a string default value to the appropriate Go type based on fieldType.
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

// parseFieldFromSource parses a single field from its specified source (query, path, header, etc.)
// and sets the value in the struct. Handles required and default values.
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
		// Smart parsing: try path first, then query.
		if pathValue := c.Params(fieldInfo.Key); pathValue != "" {
			value = pathValue
		} else if queryValue := c.Query(fieldInfo.Key); queryValue != "" {
			value = queryValue
		} else {
			// Body will be handled by BodyParser above.
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

// setFieldValue sets a struct field value with type conversion from string or interface{}.
func setFieldValue(field reflect.Value, value interface{}) error {
	switch field.Kind() {
	case reflect.String:
		if str, ok := value.(string); ok {
			field.SetString(str)
		} else {
			field.SetString(fmt.Sprintf("%v", value))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v := value.(type) {
		case string:
			if intVal, err := parseInt(v); err == nil {
				field.SetInt(int64(intVal))
			} else {
				return err
			}
		case int, int8, int16, int32, int64:
			// Use reflect to safely convert any integer type to int64.
			field.SetInt(reflect.ValueOf(v).Int())
		case float64:
			field.SetInt(int64(v))
		default:
			return fmt.Errorf("cannot convert %v to int", value)
		}
	case reflect.Bool:
		switch v := value.(type) {
		case string:
			field.SetBool(v == "true" || v == "1")
		case bool:
			field.SetBool(v)
		default:
			return fmt.Errorf("cannot convert %v to bool", value)
		}
	case reflect.Float32, reflect.Float64:
		switch v := value.(type) {
		case string:
			if floatVal, err := parseFloat(v); err == nil {
				field.SetFloat(floatVal)
			} else {
				return err
			}
		case float64:
			field.SetFloat(v)
		case int, int8, int16, int32, int64:
			// Use reflect to safely convert any integer type to float64.
			field.SetFloat(float64(reflect.ValueOf(v).Int()))
		default:
			return fmt.Errorf("cannot convert %v to float", value)
		}
	}
	return nil
}

// parseInt parses a string as an int.
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseFloat parses a string as a float64.
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
