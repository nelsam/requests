package requests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnusedFields(t *testing.T) {
	assert := assert.New(t)

	err := new(UnusedFields)
	assert.False(err.HasMissing())
	assert.Equal(0, err.NumMissing())
	assert.Equal([]string{}, err.Fields())

	err = new(UnusedFields)
	err.params = map[string]interface{}{
		"name":  "test",
		"pass":  "test",
		"email": "test@test.com",
		"about": "test",
	}
	err.matched = set{"name", "pass"}
	assert.True(err.HasMissing())
	assert.Equal(2, err.NumMissing())
	unmatched := 0
	for _, field := range err.Fields() {
		switch field {
		case "email":
			fallthrough
		case "about":
			unmatched++
		default:
			t.Errorf("Unexpected unmatched field: %s", field)
		}
	}
	assert.Equal(2, unmatched)
	assert.True(strings.HasPrefix(err.Error(),
		"Request fields found with no matching struct fields"))
	assert.True(strings.Contains(err.Error(), "email"))
	assert.True(strings.Contains(err.Error(), "about"))

	err = new(UnusedFields)
	err.params = map[string]interface{}{
		"name":  "test",
		"pass":  "test",
		"email": "test@test.com",
		"about": "test",
	}
	err.matched = set{"name", "pass", "email", "about"}
	assert.False(err.HasMissing())
	assert.False(strings.Contains(err.Error(), "name"))
	assert.False(strings.Contains(err.Error(), "pass"))
	assert.False(strings.Contains(err.Error(), "email"))
	assert.False(strings.Contains(err.Error(), "about"))
}
