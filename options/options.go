// The options package includes some default option parsers for the
// requests package.
package options

import (
	"errors"
	"reflect"
)

// Required is an option func that ensures a non-nil value was passed
// along in the request.  It does not ensure that the value is
// non-empty.
func Required(orig, value interface{}, optionValue string) (interface{}, error) {
	if optionValue == "true" {
		if value == nil {
			return nil, errors.New("Required value has nil input")
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

// Immutable is an option func that ensures that a value is not
// modified after being set.  It will return an error if orig is
// non-empty and does not match the new value from the request.
func Immutable(orig, value interface{}, optionValue string) (interface{}, error) {
	immutableErr := errors.New("Value is immutable once set")
	if optionValue == "true" {
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
					return nil, immutableErr
				}
				return value, nil
			}
			return nil, immutableErr
		}
	}
	return value, nil
}
