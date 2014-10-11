package requests

import (
	"fmt"
	"strings"
)

// UnusedFields is an error type for input values that were not used
// in a request.
type UnusedFields struct {
	params  map[string]interface{}
	matched set
	missing []string
}

// HasMissing returns whether or not this error knows about any input
// values that were not used in a request.
func (err *UnusedFields) HasMissing() bool {
	return err.NumMissing() > 0
}

// NumMissing returns the number of input values that were not used in
// a request that this error knows about.
func (err *UnusedFields) NumMissing() int {
	return len(err.params) - len(err.matched)
}

// parseMissing is used to find the input values that had no matching
// fields in a request.
func (err *UnusedFields) parseMissing() {
	err.missing = make([]string, 0, err.NumMissing())
	for param := range err.params {
		hasMatch := false
		for _, found := range err.matched {
			if param == found {
				hasMatch = true
				break
			}
		}
		if !hasMatch {
			err.missing = append(err.missing, param)
		}
	}
}

// Fields returns the request names of the fields that had no
// corresponding struct fields in a request.
func (err *UnusedFields) Fields() []string {
	if err.missing == nil {
		err.parseMissing()
	}
	return err.missing
}

// Error returns an error message listing which fields could not be
// found in the target struct.
func (err *UnusedFields) Error() string {
	return fmt.Sprintf("Request fields found with no matching struct fields: %s",
		strings.Join(err.Fields(), ", "))
}
