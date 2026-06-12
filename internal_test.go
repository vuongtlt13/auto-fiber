package autofiber

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGuessMethodFromPath(t *testing.T) {
	assert.Equal(t, "POST", guessMethodFromPath("/any/path"))
	assert.Equal(t, "POST", guessMethodFromPath(""))
}

func TestGetSchemaNameFromRef(t *testing.T) {
	assert.Equal(t, "User", GetSchemaNameFromRef("#/components/schemas/User"))
	assert.Equal(t, "", GetSchemaNameFromRef(""))
	assert.Equal(t, "SomeOtherRef", GetSchemaNameFromRef("SomeOtherRef"))
}

func TestParseParseTag_NoKey(t *testing.T) {
	// parse:"query" (no colon, no key) → key should default to field name
	type S struct {
		MyField string `parse:"query"`
	}
	f, _ := reflect.TypeOf(S{}).FieldByName("MyField")
	info := parseParseTag("query", f)
	assert.Equal(t, Query, info.Source)
	assert.Equal(t, "MyField", info.Key)
}

func TestParseParseTag_WithDefault(t *testing.T) {
	type S struct {
		Count int `parse:"query:count,default:10"`
	}
	f, _ := reflect.TypeOf(S{}).FieldByName("Count")
	info := parseParseTag("query:count,default:10", f)
	assert.Equal(t, Query, info.Source)
	assert.Equal(t, "count", info.Key)
	assert.Equal(t, 10, info.Default)
}

func TestParseParseTag_Required(t *testing.T) {
	type S struct {
		Token string `parse:"header:Authorization,required"`
	}
	f, _ := reflect.TypeOf(S{}).FieldByName("Token")
	info := parseParseTag("header:Authorization,required", f)
	assert.Equal(t, Header, info.Source)
	assert.Equal(t, "Authorization", info.Key)
	assert.True(t, info.Required)
}

func TestSetFieldValue_NumericConversions(t *testing.T) {
	var s struct {
		I  int
		I8 int8
		F  float64
		F32 float32
		B  bool
	}
	v := reflect.ValueOf(&s).Elem()

	// int from valid string
	err := setFieldValue(v.FieldByName("I"), "42")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), v.FieldByName("I").Int())

	// int from float64 (JSON number)
	err = setFieldValue(v.FieldByName("I"), float64(99))
	assert.NoError(t, err)
	assert.Equal(t, int64(99), v.FieldByName("I").Int())

	// int from int8 (typed integer)
	err = setFieldValue(v.FieldByName("I"), int8(7))
	assert.NoError(t, err)
	assert.Equal(t, int64(7), v.FieldByName("I").Int())

	// float64 from float64
	err = setFieldValue(v.FieldByName("F"), float64(3.14))
	assert.NoError(t, err)
	assert.InDelta(t, 3.14, v.FieldByName("F").Float(), 0.001)

	// float64 from valid string
	err = setFieldValue(v.FieldByName("F"), "2.71")
	assert.NoError(t, err)
	assert.InDelta(t, 2.71, v.FieldByName("F").Float(), 0.001)

	// float64 from int
	err = setFieldValue(v.FieldByName("F"), int(5))
	assert.NoError(t, err)
	assert.Equal(t, float64(5), v.FieldByName("F").Float())

	// float32 from invalid type → error
	err = setFieldValue(v.FieldByName("F32"), "bad-float-type-test")
	assert.Error(t, err)

	// bool from string "true"
	err = setFieldValue(v.FieldByName("B"), "true")
	assert.NoError(t, err)
	assert.True(t, v.FieldByName("B").Bool())

	// bool from bool false
	err = setFieldValue(v.FieldByName("B"), false)
	assert.NoError(t, err)
	assert.False(t, v.FieldByName("B").Bool())
}

func TestSchemaRequiresAuthHeader_NilAndNonStruct(t *testing.T) {
	// nil interface
	assert.False(t, schemaRequiresAuthHeader(nil))

	// non-struct (string)
	assert.False(t, schemaRequiresAuthHeader("not-a-struct"))

	// ptr to struct with auth header
	type WithAuth struct {
		Authorization string `parse:"header:Authorization,required"`
	}
	assert.True(t, schemaRequiresAuthHeader(&WithAuth{}))
}

func TestStructHasAuthHeader_EmbeddedRecurse(t *testing.T) {
	type Base struct {
		Authorization string `parse:"header:Authorization,required"`
	}
	type Derived struct {
		Base
		Name string `json:"name"`
	}
	assert.True(t, structHasAuthHeader(reflect.TypeOf(Derived{})))
}

func TestStructHasAuthHeader_EmbeddedWithTime(t *testing.T) {
	// time.Time embedded should not cause recursion panic
	type WithTime struct {
		CreatedAt time.Time
		Name      string `json:"name"`
	}
	assert.False(t, structHasAuthHeader(reflect.TypeOf(WithTime{})))
}

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
