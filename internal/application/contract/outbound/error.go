package outbound

// NotFoundError indicates the requested resource does not exist.
// Implementations must return it by value (NotFoundError{}), not by
// pointer — callers detect it with errors.As(err, &NotFoundError{}),
// which matches the value form only.
type NotFoundError struct{}

func (NotFoundError) Error() string {
	return "not found"
}
