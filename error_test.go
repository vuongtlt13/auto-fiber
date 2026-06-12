package autofiber_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestValidationRequestError_Error(t *testing.T) {
	err := &autofiber.ValidationRequestError{Message: "request failed"}
	assert.Equal(t, "request failed", err.Error())
}

func TestValidationResponseError_Error(t *testing.T) {
	err := &autofiber.ValidationResponseError{Message: "response failed"}
	assert.Equal(t, "response failed", err.Error())
}

func TestValidationRequestError_WithDetails(t *testing.T) {
	err := &autofiber.ValidationRequestError{
		Message: "Validation failed",
		Details: []autofiber.FieldErrorDetail{
			{Field: "email", Message: "email is required", Tag: "required"},
		},
	}
	assert.Equal(t, "Validation failed", err.Error())
	assert.Len(t, err.Details, 1)
	assert.Equal(t, "email", err.Details[0].Field)
}

func TestValidationResponseError_WithDetails(t *testing.T) {
	err := &autofiber.ValidationResponseError{
		Message: "Response validation failed",
		Details: []autofiber.FieldErrorDetail{
			{Field: "id", Message: "id is required", Tag: "required"},
		},
	}
	assert.Equal(t, "Response validation failed", err.Error())
	assert.Len(t, err.Details, 1)
}
