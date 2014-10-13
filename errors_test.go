package requests

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputErrors(t *testing.T) {
	assert := assert.New(t)
	errs := make(InputErrors)
	assert.Implements((*error)(nil), errs)
	assert.False(errs.HasErrors())
	assert.True(errs.Errors() == nil)
	emptyMessage := errs.Error()

	errs.Set("test", nil)
	assert.False(errs.HasErrors())
	assert.NotEqual(errs, errs.Errors())
	assert.Equal(emptyMessage, errs.Error())

	errs = make(InputErrors)
	errs.Set("test", errors.New("Test error"))
	assert.True(errs.HasErrors())
	assert.Equal(errs, errs.Errors())
	assert.NotEqual(emptyMessage, errs.Error())

	errs.Set("test2", errors.New("Second test error"))
	newErrs := InputErrors{
		"test":  errors.New("Overriding test error"),
		"test3": errors.New("New error"),
	}
	errs = errs.Merge(newErrs)
	assert.True(errs.HasErrors())
	assert.Equal(3, len(errs))
	assert.Equal("Overriding test error", errs["test"].Error())
	assert.NotEqual(emptyMessage, errs.Error())

	errs = nil
	assert.Equal(newErrs, errs.Merge(newErrs))
}
