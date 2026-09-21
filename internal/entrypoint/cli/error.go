package cli

// ArgumentError indicates the given input was malformed. Callers detect
// it with errors.As(err, &ArgumentError{}), which matches by type only,
// regardless of Cause.
type ArgumentError struct {
	Cause error
}

func (e ArgumentError) Error() string { return e.Cause.Error() }

func (e ArgumentError) Unwrap() error { return e.Cause }