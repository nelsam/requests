package requests

// Defaulter is a type that returns its own default value, for when it
// is not included in the request.  This can be used as an alternative
// to the "default=something" tag option.
type Defaulter interface {
	// DefaultValue should return the default value of this type.
	DefaultValue() interface{}
}
