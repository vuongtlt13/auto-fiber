package autofiber

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSmartSource(t *testing.T) {
	// Test GET method
	source := getSmartSource("GET")
	assert.Equal(t, Path, source)

	// Test POST method
	source = getSmartSource("POST")
	assert.Equal(t, Body, source)

	// Test PUT method
	source = getSmartSource("PUT")
	assert.Equal(t, Body, source)

	// Test PATCH method
	source = getSmartSource("PATCH")
	assert.Equal(t, Body, source)

	// Test DELETE method
	source = getSmartSource("DELETE")
	assert.Equal(t, Path, source)

	// Test unknown method (default case)
	source = getSmartSource("UNKNOWN")
	assert.Equal(t, Body, source)

	// Test case insensitive
	source = getSmartSource("get")
	assert.Equal(t, Path, source)

	source = getSmartSource("post")
	assert.Equal(t, Body, source)
}

func TestConvertDefaultValue(t *testing.T) {
	// Test string type
	result := convertDefaultValue("test", reflect.TypeOf(""))
	assert.Equal(t, "test", result)

	// Test int type
	result = convertDefaultValue("42", reflect.TypeOf(0))
	assert.Equal(t, 42, result)

	// Test invalid int (should return original string)
	result = convertDefaultValue("invalid", reflect.TypeOf(0))
	assert.Equal(t, "invalid", result)

	// Test bool type
	result = convertDefaultValue("true", reflect.TypeOf(true))
	assert.Equal(t, true, result)

	result = convertDefaultValue("1", reflect.TypeOf(true))
	assert.Equal(t, true, result)

	result = convertDefaultValue("false", reflect.TypeOf(true))
	assert.Equal(t, false, result)

	// Test float type
	result = convertDefaultValue("3.14", reflect.TypeOf(0.0))
	assert.Equal(t, 3.14, result)

	// Test invalid float (should return original string)
	result = convertDefaultValue("invalid", reflect.TypeOf(0.0))
	assert.Equal(t, "invalid", result)
}
