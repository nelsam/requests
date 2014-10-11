package requests

// An Unmarshaller is a type that is capable of unmarshalling data,
// itself, rather than relying on generic behavior.
type Unmarshaller interface {
	Unmarshal(params map[string]interface{}) error
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
