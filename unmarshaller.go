package requests

// An Unmarshaller is a type that is capable of unmarshalling data,
// itself, rather than relying on generic behavior.  The calling
// function will make its best attempt at sending a map[string]interface{}
// to Unmarshal as the body parameter, but there are situations where
// it won't be.
//
// Primarily, the type will depend on the request's Content-Type header.
// Any parsable Content-Type will be parsed, to the best of this library's
// ability, into a map[string]interface{} as you would expect to see from
// Request.Params().  However, unrecognized Content-Type headers will cause
// the raw request.Body to be passed along instead.
//
// Most of the time, you should be able to assume that body is of type
// map[string]interface{} - the only time that body will be the raw
// request.Body *should* be when someone is performing a file upload using
// the file's raw bytes as the body of the request, such as the following
// example:
// https://developer.mozilla.org/en-US/docs/Using_files_from_web_applications#Example.3A_Uploading_a_user-selected_file
//
// Except on resource endpoints where you want to support that, you can safely
// run params := body.(map[string]interface{}) - panics from Unmarshal will be
// caught and handled as errors.
type Unmarshaller interface {
	Unmarshal(body interface{}) error
}

// A PreUnmarshaller is a type that performs certain actions prior to
// having data unmarshalled to it.
type PreUnmarshaller interface {
	PreUnmarshal() error
}

// A PostUnmarshaller is a type that performs certain actions
// subsequent to having data unmarshalled to it.
type PostUnmarshaller interface {
	PostUnmarshal() error
}
