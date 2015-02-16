// The options package includes some default option parsers for the
// requests package.
package options

import (
	"errors"
	"reflect"
)

var (
	ErrRequiredMissing = errors.New("Required value has nil input")
	ErrValueImmutable  = errors.New("Value is immutable once set")
)

// changeReceiver is just a clone of requests.ChangeReceiver, since we can't
// import requests in this package.
type changeReceiver interface {
	Receive(interface{}) (bool, error)
}

// receiver is a clone of requests.Receiver, similar to changeReceiver.
type receiver interface {
	Receive(interface{}) error
}

func isPtrOrInter(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface
}

// zeroOrEqual checks whether orig is either the zero value of its type or
// loosely equal to value.  The main utility here is that zeroOrEqual(4.0, 4)
// will return true, as will zeroOrEqual(new(int64), float32(0)).
func zeroOrEqual(orig, value interface{}) bool {
	origType := reflect.TypeOf(orig)
	if orig == reflect.Zero(origType).Interface() {
		return true
	}
	origVal := reflect.New(origType).Elem()
	origVal.Set(reflect.ValueOf(orig))
	compareValue := reflect.ValueOf(value)
	for !compareValue.Type().ConvertibleTo(origType) && isPtrOrInter(origType) {
		if origVal.IsNil() {
			if origVal.Kind() != reflect.Ptr {
				// Can't initialize
				return false
			}
			origVal.Set(reflect.New(origType.Elem()))
		}
		origVal = origVal.Elem()
		origType = origType.Elem()
	}
	if compareValue.Type().ConvertibleTo(origType) {
		compareValue = compareValue.Convert(origType)
	}
	return origVal.Interface() == compareValue.Interface()
}

// Required is an option func that ensures a non-nil value was passed
// along in the request.  It does not ensure that the value is
// non-empty.
func Required(orig, value interface{}, fromRequest bool, optionValue string) (interface{}, error) {
	if optionValue == "true" {
		if fromRequest {
			if value == nil || value == reflect.Zero(reflect.TypeOf(orig)).Interface() {
				return nil, ErrRequiredMissing
			}
		} else {
			if orig == reflect.Zero(reflect.TypeOf(orig)).Interface() {
				return nil, ErrRequiredMissing
			}
		}
	}
	return value, nil
}

// Default is an option func that sets a default value for a field.
// If its value is non-empty, it will use the following logic to
// decide what to return:
//
// If fromRequest is false, then the default will be returned if orig
// is equal to its zero value.
//
// If fromRequest is true, then the default will be returned if value
// is nil.
//
// In all other cases, value will be returned.
func Default(orig, value interface{}, fromRequest bool, optionValue string) (interface{}, error) {
	if optionValue != "" {
		// optionValue is always a string type, but we'll leave it up
		// to the calling code to perform conversion to orig's type.
		useDefault := false
		if fromRequest {
			useDefault = value == nil
		} else {
			useDefault = orig == reflect.Zero(reflect.TypeOf(orig)).Interface()
		}
		if useDefault {
			return optionValue, nil
		}
	}
	return value, nil
}

// Immutable is an option func that ensures that a value is not
// modified after being set.  It will return an error if orig is
// non-empty and does not match the new value from the request.
func Immutable(orig, value interface{}, fromRequest bool, optionValue string) (interface{}, error) {
	if optionValue == "true" && fromRequest {
		if _, ok := orig.(receiver); ok {
			return nil, errors.New("Receiver types cannot be immutable.  " +
				"See ChangeReceiver for a supported alternative.")
		}
		if !zeroOrEqual(orig, value) {
			if changeReceiver, ok := orig.(changeReceiver); ok {
				changed, err := changeReceiver.Receive(value)
				if err != nil {
					return nil, err
				}
				if changed {
					return nil, ErrValueImmutable
				}
				return value, nil
			}
			return nil, ErrValueImmutable
		}
	}
	return value, nil
}
