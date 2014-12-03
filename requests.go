// The requests package contains logic for loading and unmarshalling
// data contained within web requests.  The most common uses for this
// library are as follows:
//
//     params, err := requests.New(request).Params()
//
//     err := requests.New(request).Unmarshal(structPtr)
//
// Parameters will be loaded from the request body based on the
// request's Content-Type header.  Some attempts are made to unify
// data structure, to make it easier to treat all requests the same
// (regardless of Content-Type).
//
// For the Unmarshal process, the requests package uses a combination
// of reflection (to check field tags) and interfaces to figure out
// which values (from the above params) should be applied to which
// fields in the target struct.  Unmarshalling to non-struct types is
// not supported.
package requests

import "net/http"

// A Request is a type that stores data about an HTTP request and
// contains methods for reading that request's body.
type Request struct {
	httpRequest *http.Request
	body        interface{}
	params      map[string]interface{}
	queryParams map[string]interface{}
}

// New creates a new *Request based on the request parameter.
func New(request *http.Request) *Request {
	return &Request{httpRequest: request}
}
