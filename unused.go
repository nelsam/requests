package requests

import (
	"fmt"
	"strings"
)

type UnusedFields struct {
	params  map[string]interface{}
	matched set
	missing []string
}

func (err *UnusedFields) HasMissing() bool {
	return err.NumMissing() > 0
}

func (err *UnusedFields) NumMissing() int {
	return len(err.params) - len(err.matched)
}

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

func (err *UnusedFields) Fields() []string {
	if err.missing == nil {
		err.parseMissing()
	}
	return err.missing
}

func (err *UnusedFields) Error() string {
	return fmt.Sprintf("Request fields found with no matching struct fields: %s",
		strings.Join(err.Fields(), ", "))
}
