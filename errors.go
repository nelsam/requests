package requests

import "bytes"

// InputErrors is an error type that maps input names to errors
// encountered while parsing their value.  A nil error will be stored
// for any input names that were parsed successfully.
type InputErrors map[string]error

// Error returns the InputError's full error string.
func (errs InputErrors) Error() string {
	buff := bytes.NewBufferString("Input errors:\n\n")
	for name, err := range errs {
		if err != nil {
			buff.WriteString(" * ")
			buff.WriteString(name)
			buff.WriteString(": ")
			buff.WriteString(err.Error())
			buff.WriteString("\n")
		}
	}
	return buff.String()
}

// Set takes an input and an error, and sets the error to the input if
// the error is non-nil.  The return value will be true if err is
// non-nil, false otherwise.
func (errs InputErrors) Set(input string, err error) bool {
	errs[input] = err
	return err == nil
}

func (errs InputErrors) Merge(newErrs InputErrors) InputErrors {
	for input, err := range newErrs {
		errs[input] = err
	}
	return errs
}

func (errs InputErrors) HasErrors() bool {
	for _, err := range errs {
		if err != nil {
			return true
		}
	}
	return false
}
