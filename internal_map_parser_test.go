package autofiber

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFromMap_ErrorCases(t *testing.T) {
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// schema is not a pointer
	data := map[string]interface{}{"id": 1, "name": "test"}
	err := ParseFromMap(data, TestStruct{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pointer")

	// map missing required key
	ptr := &TestStruct{}
	data = map[string]interface{}{"id": 1}
	err = ParseFromMap(data, ptr)
	assert.NoError(t, err)
	assert.Equal(t, 1, ptr.ID)
	assert.Equal(t, "", ptr.Name)

	// value wrong type
	data = map[string]interface{}{"id": "not-an-int", "name": "test"}
	ptr = &TestStruct{}
	err = ParseFromMap(data, ptr)
	assert.Error(t, err)
}

func TestParseFromInterface_ErrorCases(t *testing.T) {
	type TestStruct struct {
		ID int `json:"id"`
	}

	// map[string]string
	data := map[string]string{"id": "1"}
	ptr := &TestStruct{}
	err := ParseFromInterface(data, ptr)
	assert.NoError(t, err)
	assert.Equal(t, 1, ptr.ID)

	// struct
	src := TestStruct{ID: 42}
	ptr = &TestStruct{}
	err = ParseFromInterface(src, ptr)
	assert.NoError(t, err)
	assert.Equal(t, 42, ptr.ID)

	// unsupported type
	err = ParseFromInterface(123, ptr)
	assert.Error(t, err)
}

func TestParseFromStruct_ErrorCases(t *testing.T) {
	type Src struct {
		A string `json:"a"`
	}
	type Dst struct {
		B string `json:"b"`
	}

	src := Src{A: "foo"}
	ptr := &Dst{}
	err := parseFromStruct(src, ptr)
	assert.NoError(t, err)
	assert.Equal(t, "", ptr.B)

	// schema is not pointer
	err = parseFromStruct(src, Dst{})
	assert.Error(t, err)
}

func TestGetFieldKey_ErrorCases(t *testing.T) {
	field, _ := reflect.TypeOf(struct {
		A string `json:"-"`
		B string
	}{}).FieldByName("A")
	key := getFieldKey(field)
	assert.Equal(t, "A", key) // Should fallback to field name

	field, _ = reflect.TypeOf(struct {
		B string
	}{}).FieldByName("B")
	key = getFieldKey(field)
	assert.Equal(t, "B", key)
}

func TestParseFromMapInternal_EmbeddedPtrStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}
	type Person struct {
		Name    string   `json:"name"`
		Address *Address // embedded ptr
	}

	data := map[string]interface{}{
		"name":   "Alice",
		"street": "123 Main St",
		"city":   "Springfield",
	}
	p := &Person{}
	err := parseFromMapInternal(data, p)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", p.Name)
	// Address is not anonymous (not embedded), so street/city won't be set here
	// This test verifies it doesn't panic on ptr fields
}

func TestParseFromMapInternal_AnonymousEmbeddedPtr(t *testing.T) {
	type Base struct {
		Role string `json:"role"`
	}
	type User struct {
		*Base
		Name string `json:"name"`
	}

	data := map[string]interface{}{
		"name": "Bob",
		"role": "admin",
	}
	u := &User{}
	err := parseFromMapInternal(data, u)
	assert.NoError(t, err)
	assert.Equal(t, "Bob", u.Name)
	assert.NotNil(t, u.Base)
	assert.Equal(t, "admin", u.Role)
}

func TestSetFieldValue_ErrorCases(t *testing.T) {
	var s struct {
		I int
		S string
		B bool
		F float64
	}
	v := reflect.ValueOf(&s).Elem()

	// int field, set string (should error)
	err := setFieldValue(v.FieldByName("I"), "not-an-int")
	assert.Error(t, err)

	// string field, set int (should convert to string)
	err = setFieldValue(v.FieldByName("S"), 123)
	assert.NoError(t, err)
	assert.Equal(t, "123", v.FieldByName("S").String())

	// bool field, set int (should error)
	err = setFieldValue(v.FieldByName("B"), 1)
	assert.Error(t, err)

	// float field, set string (invalid float)
	err = setFieldValue(v.FieldByName("F"), "not-a-float")
	assert.Error(t, err)
}
