// Package autofiber provides response validation utilities for ensuring API responses match expected schemas.
package autofiber

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// validateResponseData validates response data against the provided schema using the given validator.
// It supports validating structs, pointers to structs, slices of structs, and maps. If the schema is nil, validation is skipped.
// Returns an error if validation fails, or nil if the data is valid.
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

	// For maps, convert to struct before validation
	if dataValue.Kind() == reflect.Map {
		// Create a new instance of the schema
		schemaType := reflect.TypeOf(schema)
		if schemaType.Kind() == reflect.Ptr {
			schemaType = schemaType.Elem()
		}

		// Convert any map type to map[string]interface{}
		var mapData map[string]interface{}

		// Try different map type assertions
		if m, ok := data.(map[string]interface{}); ok {
			mapData = m
		} else if fm, ok := data.(fiber.Map); ok {
			mapData = map[string]interface{}(fm)
		} else {
			// For any other map type, try to convert it
			dataType := dataValue.Type()
			if dataType.Key().Kind() == reflect.String {
				// Convert map to map[string]interface{}
				mapData = make(map[string]interface{})
				iter := dataValue.MapRange()
				for iter.Next() {
					key := iter.Key().String()
					value := iter.Value().Interface()
					mapData[key] = value
				}
			}
		}

		if mapData != nil {
			structData := reflect.New(schemaType).Interface()
			if err := ParseFromMap(mapData, structData); err != nil {
				return fmt.Errorf("failed to convert map to struct: %w", err)
			}
			return validator.Struct(structData)
		}

		// If we can't convert the map, return an error
		return fmt.Errorf("unsupported map type: %T", data)
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
