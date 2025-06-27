package autofiber_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestAutoFiber_Struct(t *testing.T) {
	app := autofiber.New()
	assert.NotNil(t, app)
	assert.NotNil(t, app.App)
}

func TestRouteOptions_Struct(t *testing.T) {
	options := &autofiber.RouteOptions{
		Description: "Test route",
		Tags:        []string{"test"},
	}

	assert.Equal(t, "Test route", options.Description)
	assert.Equal(t, []string{"test"}, options.Tags)
}

func TestParseSource_Constants_Types(t *testing.T) {
	assert.Equal(t, autofiber.ParseSource("body"), autofiber.Body)
	assert.Equal(t, autofiber.ParseSource("query"), autofiber.Query)
	assert.Equal(t, autofiber.ParseSource("path"), autofiber.Path)
	assert.Equal(t, autofiber.ParseSource("header"), autofiber.Header)
	assert.Equal(t, autofiber.ParseSource("cookie"), autofiber.Cookie)
	assert.Equal(t, autofiber.ParseSource("form"), autofiber.Form)
	assert.Equal(t, autofiber.ParseSource("auto"), autofiber.Auto)
}
