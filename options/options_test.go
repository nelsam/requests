package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequired(t *testing.T) {
	var (
		orig        string
		value       interface{}
		optionValue = "false"
	)
	_, err := Required(orig, value, optionValue)
	assert.NoError(t, err)

	optionValue = "true"
	_, err = Required(orig, value, optionValue)
	assert.Error(t, err)

	orig = "test"
	_, err = Required(orig, value, optionValue)
	assert.Error(t, err)

	value = "test value"
	v, err := Required(orig, value, optionValue)
	assert.NoError(t, err)
	assert.Equal(t, value, v)
}

func TestDefault(t *testing.T) {
	var (
		orig         string
		value        interface{}
		defaultValue string
	)
	v, err := Default(orig, value, defaultValue)
	assert.NoError(t, err)
	assert.Equal(t, value, v)

	defaultValue = "test"
	v, err = Default(orig, value, defaultValue)
	assert.NoError(t, err)
	assert.Equal(t, defaultValue, v)

	orig = "test orig"
	v, err = Default(orig, value, defaultValue)
	assert.NoError(t, err)
	assert.Equal(t, defaultValue, v)

	value = "test input"
	v, err = Default(orig, value, defaultValue)
	assert.NoError(t, err)
	assert.Equal(t, value, v)
}

func TestImmutable(t *testing.T) {
	var (
		orig        string      = "orig test"
		value       interface{} = "value test"
		optionValue             = "false"
	)

	v, err := Immutable(orig, value, optionValue)
	assert.NoError(t, err)
	assert.Equal(t, value, v)

	optionValue = "true"
	_, err = Immutable(orig, value, optionValue)
	assert.Error(t, err)

	orig = ""
	v, err = Immutable(orig, value, optionValue)
	assert.NoError(t, err)
	assert.Equal(t, value, v)

	orig = "test value"
	value = "test value"
	v, err = Immutable(orig, value, optionValue)
	assert.NoError(t, err)
	assert.Equal(t, orig, v)
}
