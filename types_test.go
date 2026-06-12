package autofiber_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestAutoFiber_Struct(t *testing.T) {
	app := autofiber.New(fiber.Config{})
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

func TestValidateStruct_Valid(t *testing.T) {
	type User struct {
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=18"`
	}
	u := &User{Email: "test@example.com", Age: 20}
	err := autofiber.ValidateStruct(u)
	assert.NoError(t, err)
}

func TestValidateStruct_Invalid(t *testing.T) {
	type User struct {
		Email string `validate:"required,email"`
		Age   int    `validate:"gte=18"`
	}
	u := &User{Email: "not-an-email", Age: 10}
	err := autofiber.ValidateStruct(u)
	assert.Error(t, err)
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
