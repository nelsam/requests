package requests

import (
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

	err = new(UnusedFields)
	err.params = map[string]interface{}{
		"name":  "test",
		"pass":  "test",
		"email": "test@test.com",
		"about": "test",
	}
	err.matched = set{"name", "pass", "email", "about"}
	assert.False(err.HasMissing())
}
