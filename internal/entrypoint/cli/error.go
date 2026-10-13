package cli

// ArgumentError indicates the given input was malformed.
type ArgumentError struct {
	Cause error
}

// Error returns the message of the Cause, or a generic message if there is none.
func (e ArgumentError) Error() string {
	if e.Cause == nil {
		return "invalid arguments"
	}
	return e.Cause.Error()
}

// Unwrap returns the Cause.
func (e ArgumentError) Unwrap() error { return e.Cause }
