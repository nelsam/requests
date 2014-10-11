// The requests package contains logic for loading and unmarshalling
// data contained within web requests.  The most common uses for this
// library are as follows:
//
//     params, err := requests.New(request).Params()
//
//     err := requests.New(request).Unmarshal(structPtr)
//
// For parameter parsing, the requests package uses the following
// logic:
//
//   1. A map[string]interface{} is created to store parameters.
//   2. If there are URL parameters, they are appended to the
//      map[string]interface{} using standard urlencode unmarshalling.
//   3. If the request body is non-empty:
//     1. Look up a codec matching the request's Content-Type header.
//     2. If no matching codec is found, fall back on urlencoded data.
//     3. Unmarshal values from the request body and append them to the
//        map[string]interface{}.
//
// The return value is the map[string]interface{} generated during
// that process.
//
// For the Unmarshal process, the requests package uses a combination
// of reflection (to check field tags) and interfaces to figure out
// which values (from the above params) should be applied to which
// fields in the target struct.  Unmarshalling to non-struct types is
// not supported.
package requests

import "net/http"

type Request struct {
	httpRequest *http.Request
	body        interface{}
	params      map[string]interface{}
}

func New(request *http.Request) *Request {
	return &Request{httpRequest: request}
}
