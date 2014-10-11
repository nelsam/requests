// The options package includes some default option parsers for the
// requests package.
package options

import (
	"errors"
	"reflect"
)

var RequiredErr = errors.New("Required value is nil")

// Required is an option func that ensures a non-nil value was passed
// along in the request.  It does not ensure that the value is
// non-empty.
func Required(orig, value interface{}, optionValue string) (interface{}, error) {
	if optionValue == "true" {
		if value == nil {
			return nil, RequiredErr
		}
	}
	return value, nil
}

// Default is an option func that sets a default value for a field.
// If the value doesn't exist in the request (or is nil), the provided
// default will be used instead.
func Default(orig, value interface{}, optionValue string) (interface{}, error) {
	if value == nil {
		if optionValue != "" {
			// This is a string type, but we'll leave it up to the
			// unmarshal process (or the Receiver's Receive method) to
			// convert it.
			return optionValue, nil
		}
	}
	return value, nil
}

// Immutable is an option func that ensures that a value is not
// modified after being set.  It will return an error if orig is
// non-empty and does not match the new value from the request.
func Immutable(orig, value interface{}, optionValue string) (interface{}, error) {
	if orig != reflect.Zero(reflect.TypeOf(orig)).Interface() && orig != value {
		return nil, errors.New("Value is immutable once set")
	}
	return value, nil
}
