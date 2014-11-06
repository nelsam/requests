package requests

// A ReceiveTyper has methods for returning the type that its
// Receive method expects.  It *must* implement Receiver as well,
// otherwise its ReceiveType method will be useless.  If it does,
// then request data destined for the ReceiveTyper will be converted
// (if possible) to the same type as the return value of its
// ReceiveType method.
type ReceiveTyper interface {
	Receiver

	// ReceiveType should return a value (preferably empty) of the same
	// type as the ReceiveTyper's Receive method expects.  Any value
	// in a request destined to be an argument for the ReceiveTyper's
	// Receive method will first be converted to the same type as the
	// value returned by ReceiveType.
	ReceiveType() interface{}
}

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

// A ChangeReceiver is a receiver that, in addition to performing its
// own logic for receiving input values, also returns whether or not the
// passed in value was different from the existing value.
//
// This is used primarily for immutable option checking.  Struct fields
// of type Receiver cannot support the "immutable" option, so types
// which are used in struct fields that need the "immutable" option should
// implement ChangeReceiver, instead.
//
// Note that this will *not* be used if the current value of the field is
// equal to the empty value of the field - only if the field is set to
// a non-empty value will the immutable option care about the ChangeReceiver
// interface.
type ChangeReceiver interface {
	// Receive takes a value and performs the same logic as
	// Receiver.Receive, but returns whether or not the new value is
	// different from the old value, as well as any errors encountered.
	Receive(interface{}) (valueChanged bool, err error)
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
