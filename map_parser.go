package autofiber

import (
	"fmt"
	"reflect"
	"strings"
)

// ParseFromMap parses a struct from a map[string]interface{}
func ParseFromMap(data map[string]interface{}, schema interface{}) error {
	return parseFromMapInternal(data, schema)
}

// ParseFromInterface parses a struct from any interface{} (map, struct, etc.)
func ParseFromInterface(data interface{}, schema interface{}) error {
	return parseFromInterfaceInternal(data, schema)
}

// parseFromMapInternal parses a struct from a map[string]interface{}
func parseFromMapInternal(data map[string]interface{}, schema interface{}) error {
	reqValue := reflect.ValueOf(schema)
	if reqValue.Kind() != reflect.Ptr {
		return fmt.Errorf("schema must be a pointer")
	}
	reqValue = reqValue.Elem()

	reqType := reqValue.Type()

	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		fieldValue := reqValue.Field(i)

		// Get field key from json tag or field name
		key := getFieldKey(field)

		// Get value from map
		if value, exists := data[key]; exists {
			if err := setFieldValue(fieldValue, value); err != nil {
				return fmt.Errorf("field %s: %w", key, err)
			}
		}
	}

	return nil
}

// parseFromInterfaceInternal parses a struct from any interface{} (map, struct, etc.)
func parseFromInterfaceInternal(data interface{}, schema interface{}) error {
	// Handle map[string]interface{}
	if mapData, ok := data.(map[string]interface{}); ok {
		return parseFromMapInternal(mapData, schema)
	}

	// Handle map[string]string
	if mapData, ok := data.(map[string]string); ok {
		// Convert to map[string]interface{}
		interfaceMap := make(map[string]interface{})
		for k, v := range mapData {
			interfaceMap[k] = v
		}
		return parseFromMapInternal(interfaceMap, schema)
	}

	// Handle struct by converting to map
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	if dataValue.Kind() == reflect.Struct {
		return parseFromStruct(data, schema)
	}

	return fmt.Errorf("unsupported data type: %T", data)
}

// parseFromStruct parses from one struct to another
func parseFromStruct(data interface{}, schema interface{}) error {
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	reqValue := reflect.ValueOf(schema)
	if reqValue.Kind() != reflect.Ptr {
		return fmt.Errorf("schema must be a pointer")
	}
	reqValue = reqValue.Elem()

	dataType := dataValue.Type()

	// Create a map of field names to values from data struct
	dataMap := make(map[string]interface{})
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		key := getFieldKey(field)
		value := dataValue.Field(i).Interface()
		dataMap[key] = value
	}

	return parseFromMapInternal(dataMap, schema)
}

// getFieldKey gets the key name for a field from json tag or field name
func getFieldKey(field reflect.StructField) string {
	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		jsonParts := strings.Split(jsonTag, ",")
		if jsonParts[0] != "" {
			return jsonParts[0]
		}
	}
	return field.Name
}
