package autofiber

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestValidateResponseData_SimpleTypes(t *testing.T) {
	type SimpleStruct struct {
		ID   int    `validate:"required"`
		Name string `validate:"required"`
	}

	validator := GetValidator()

	// Test with nil schema (should skip validation)
	err := validateResponseData(SimpleStruct{ID: 1, Name: "test"}, nil, validator)
	assert.NoError(t, err)

	// Test with same type (direct validation)
	data := SimpleStruct{ID: 1, Name: "test"}
	err = validateResponseData(data, SimpleStruct{}, validator)
	assert.NoError(t, err)

	// Test with invalid data (missing required fields)
	err = validateResponseData(SimpleStruct{ID: 1}, SimpleStruct{}, validator)
	assert.Error(t, err)
}

func TestValidateResponseData_WithSlices(t *testing.T) {
	type Item struct {
		ID   int    `validate:"required"`
		Name string `validate:"required"`
	}

	validator := GetValidator()

	// Test with valid slice
	data := []Item{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
	}
	err := validateResponseData(data, Item{}, validator)
	assert.NoError(t, err)

	// Test with invalid slice (one item missing required field)
	data = []Item{
		{ID: 1, Name: "Item 1"},
		{ID: 2}, // Missing Name
	}
	err = validateResponseData(data, Item{}, validator)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "element 1")
}

func TestValidateResponseData_WithUnsupportedMapType(t *testing.T) {
	type ValidStruct struct {
		ID   int    `validate:"required"`
		Name string `validate:"required"`
	}

	validator := GetValidator()

	// Test with unsupported map type (map[int]string instead of map[string]interface{})
	data := map[int]string{1: "test"}
	err := validateResponseData(data, ValidStruct{}, validator)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported map type")
}

func TestValidateResponseData_WithComplexMapTypes(t *testing.T) {
	type ValidStruct struct {
		ID   int    `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
	}

	validator := GetValidator()

	// Test with map[string]interface{} (should work)
	data := map[string]interface{}{
		"id":   1,
		"name": "test",
	}
	err := validateResponseData(data, ValidStruct{}, validator)
	assert.NoError(t, err)

	// Test with fiber.Map (should work)
	fiberMap := fiber.Map{
		"id":   1,
		"name": "test",
	}
	err = validateResponseData(fiberMap, ValidStruct{}, validator)
	assert.NoError(t, err)
}
