package requests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Fallback struct {
	First  string `db:"first_field"`
	Second string `response:"second_field"`
	Third  string `db:"-" response:"third_field,responseoption"`
	Fourth string `db:"fourth_field,dboption" response:"-"`
	Fifth  string `db:"-" response:"-" request:"fifth_field,requestoption"`
}

func fallbackFields(structType reflect.Type) []string {
	names := make([]string, 0, 5)
	fieldNames := []string{
		"First",
		"Second",
		"Third",
		"Fourth",
		"Fifth",
	}
	for _, fieldName := range fieldNames {
		field, ok := structType.FieldByName(fieldName)
		if !ok {
			panic("Could not find field named " + fieldName + " on test struct")
		}
		names = append(names, name(field))
	}
	return names
}

func TestTags_Fallbacks(t *testing.T) {
	assert := assert.New(t)
	structType := reflect.TypeOf(Fallback{})

	noFallbackFields := fallbackFields(structType)
	expected := []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth_field",
	}
	assert.Equal(expected, noFallbackFields)

	AddFallbackTag("db")
	dbFallbackFields := fallbackFields(structType)
	expected[0] = "first_field"
	expected[2] = "-"
	expected[3] = "fourth_field"
	assert.Equal(expected, dbFallbackFields)

	AddFallbackTag("response")
	allFallbackFields := fallbackFields(structType)
	expected[1] = "second_field"
	assert.Equal(expected, allFallbackFields)

	fallbackTags = nil
	AddFallbackTag("response")
	AddFallbackTag("db")
	reverseFallbackFields := fallbackFields(structType)
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
