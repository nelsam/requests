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

// Set executes errs[input] = err.  Returns true if err is non-nil,
// false otherwise.
func (errs InputErrors) Set(input string, err error) bool {
	errs[input] = err
	return err != nil
}

// Merge merges errs and newErrs, returning the resulting map. If errs
// is nil, nothing will be done and newErrs will be returned. If errs
// is non-nil, all keys and values in newErrs will be added to errs,
// and you can safely ignore the return value.
func (errs InputErrors) Merge(newErrs InputErrors) InputErrors {
	if errs == nil {
		return newErrs
	}
	for input, err := range newErrs {
		errs[input] = err
	}
	return errs
}

// HasErrors returns whether or not any of the errors in errs are
// non-nil.
func (errs InputErrors) HasErrors() bool {
	for _, err := range errs {
		if err != nil {
			return true
		}
	}
	return false
}

// Errors returns a clone of errs with all nil error indexes removed.
// If there are no non-nil errors, this method will return nil.
func (errs InputErrors) Errors() InputErrors {
	var errors InputErrors
	for input, err := range errs {
		if err != nil {
			if errors == nil {
				errors = make(InputErrors)
			}
			errors[input] = err
		}
	}
	return errors
}
