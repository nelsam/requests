package requests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Fallback struct {
	First  string `db:"first_field" request:",foo,bar=baz,bacon"`
	Second string `response:"second_field"`
	Third  string `db:"-" response:"third_field,responseoption"`
	Fourth string `db:"fourth_field,dboption" response:"-"`
	Fifth  string `db:"-" response:"-" request:"fifth_field,requestoption"`
}

func fieldNames(structType reflect.Type) []string {
	names := make([]string, 0, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		names = append(names, name(structType.Field(i)))
	}
	return names
}

func fieldOptions(structType reflect.Type) [][]*tagOption {
	options := make([][]*tagOption, 0, structType.NumField())
	for i := 0; i < structType.NumField(); i++ {
		options = append(options, tagOptions(structType.Field(i)))
	}
	return options
}

func TestTags_Fallbacks(t *testing.T) {
	assert := assert.New(t)
	structType := reflect.TypeOf(Fallback{})

	noFallbackFields := fieldNames(structType)
	expected := []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth_field",
	}
	assert.Equal(expected, noFallbackFields)

	AddFallbackTag("db")
	dbFallbackFields := fieldNames(structType)
	expected[0] = "first_field"
	expected[2] = "-"
	expected[3] = "fourth_field"
	assert.Equal(expected, dbFallbackFields)

	AddFallbackTag("response")
	allFallbackFields := fieldNames(structType)
	expected[1] = "second_field"
	assert.Equal(expected, allFallbackFields)

	fallbackTags = nil
	AddFallbackTag("response")
	AddFallbackTag("db")
	reverseFallbackFields := fieldNames(structType)
	expected[2] = "third_field"
	expected[3] = "-"
	assert.Equal(expected, reverseFallbackFields)
}

func TestTags_Fallbacks_NoDuplicates(t *testing.T) {
	AddFallbackTag("db")
	expectedLength := len(fallbackTags)
	AddFallbackTag("db")
	assert.Equal(t, expectedLength, len(fallbackTags))
}

func TestTagOptions(t *testing.T) {
	assert := assert.New(t)
	options := fieldOptions(reflect.TypeOf(Fallback{}))
	// Options on the First field
	if assert.Equal(3, len(options[0])) {
		assert.Equal("foo", options[0][0].name)
		assert.Equal("true", options[0][0].value)
		assert.Equal("bar", options[0][1].name)
		assert.Equal("baz", options[0][1].value)
		assert.Equal("bacon", options[0][2].name)
		assert.Equal("true", options[0][2].value)
	}

	// Options on the Second field should be empty
	assert.Equal(0, len(options[1]))

	// Options on the Fifth field
	if assert.Equal(1, len(options[4])) {
		assert.Equal("requestoption", options[4][0].name)
		assert.Equal("true", options[4][0].value)
	}
}
