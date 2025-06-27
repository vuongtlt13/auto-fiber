package autofiber_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	autofiber "github.com/vuongtlt13/auto-fiber"
)

func TestParseFromMap(t *testing.T) {
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	data := map[string]interface{}{
		"name":  "John",
		"age":   "25",
		"email": "john@example.com",
	}

	user := &User{}
	err := autofiber.ParseFromMap(data, user)

	assert.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 25, user.Age)
	assert.Equal(t, "john@example.com", user.Email)
}

func TestParseFromMap_WithJsonTag(t *testing.T) {
	type User struct {
		Name  string `json:"user_name"`
		Age   int    `json:"user_age"`
		Email string `json:"email"`
	}

	data := map[string]interface{}{
		"user_name": "John",
		"user_age":  "25",
		"email":     "john@example.com",
	}

	user := &User{}
	err := autofiber.ParseFromMap(data, user)

	assert.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 25, user.Age)
	assert.Equal(t, "john@example.com", user.Email)
}

func TestParseFromMap_NonPointerSchema(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}

	data := map[string]interface{}{
		"name": "John",
	}

	user := User{} // Not a pointer
	err := autofiber.ParseFromMap(data, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema must be a pointer")
}

func TestParseFromInterface_Map(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := map[string]interface{}{
		"name": "John",
		"age":  "25",
	}

	user := &User{}
	err := autofiber.ParseFromInterface(data, user)

	assert.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 25, user.Age)
}

func TestParseFromInterface_MapString(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := map[string]string{
		"name": "John",
		"age":  "25",
	}

	user := &User{}
	err := autofiber.ParseFromInterface(data, user)

	assert.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.Equal(t, 25, user.Age)
}

func TestParseFromInterface_Struct(t *testing.T) {
	type SourceUser struct {
		Name string `json:"name"`
		Age  string `json:"age"`
	}

	type TargetUser struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	source := &SourceUser{
		Name: "John",
		Age:  "25",
	}

	target := &TargetUser{}
	err := autofiber.ParseFromInterface(source, target)

	assert.NoError(t, err)
	assert.Equal(t, "John", target.Name)
	assert.Equal(t, 25, target.Age)
}

func TestParseFromInterface_UnsupportedType(t *testing.T) {
	type User struct {
		Name string `json:"name"`
	}

	data := "not a map or struct"
	user := &User{}
	err := autofiber.ParseFromInterface(data, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported data type")
}
