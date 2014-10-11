package requests

// A Receiver is a type that receives a value from a request and
// performs its own logic to apply the input value to itself.
//
// Example:
//
//     type Password string
//
//     func (pass *Password) Receive(rawPassword interface{}) error {
//         *pass = hash(rawPassword.(string))
//     }
type Receiver interface {
	// Receive takes a value and attempts to read it in to the
	// underlying type.  It should return an error if the passed in
	// value cannot be parsed to the underlying type.
	Receive(interface{}) error
}

// A PreReceiver has an action to perform prior to receiving data from
// a user request.
type PreReceiver interface {
	// PreReceive performs initial tasks prior to receiving a value
	// from input.
	PreReceive() error
}

// A PostReceiver has an action to perform subsequent to receiving
// data from a user request.
type PostReceiver interface {
	// PostReceive performs final tasks subsequent to receiving a
	// value from input.
	PostReceive() error
}
