package autofiber_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestParseError_Error(t *testing.T) {
	err := &autofiber.ParseError{
		Field:   "age",
		Source:  "query",
		Message: "invalid",
	}
	assert.Equal(t, "age (query): invalid", err.Error())
}

func TestParseSource_Constants(t *testing.T) {
	assert.Equal(t, autofiber.ParseSource("body"), autofiber.Body)
	assert.Equal(t, autofiber.ParseSource("query"), autofiber.Query)
	assert.Equal(t, autofiber.ParseSource("path"), autofiber.Path)
	assert.Equal(t, autofiber.ParseSource("header"), autofiber.Header)
	assert.Equal(t, autofiber.ParseSource("cookie"), autofiber.Cookie)
	assert.Equal(t, autofiber.ParseSource("form"), autofiber.Form)
	assert.Equal(t, autofiber.ParseSource("auto"), autofiber.Auto)
}
