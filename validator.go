// Package autofiber provides response validation utilities for ensuring API responses match expected schemas.
package autofiber

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// validateResponseData validates response data against the provided schema using the given validator.
// It supports validating structs, pointers to structs, and slices of structs. If the schema is nil, validation is skipped.
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
