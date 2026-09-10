package outbound

// NotFoundError indicates the requested resource does not exist.
// Implementations must return it by value (NotFoundError{}), not by
// pointer — callers detect it with errors.As(err, &NotFoundError{}),
// which matches the value form only.
type NotFoundError struct{}

func (NotFoundError) Error() string {
	return "not found"
}

// CorruptedDataError indicates persisted data was retrieved successfully
// but could not be parsed as valid data. Implementations must return it
// by value (CorruptedDataError{}), not by pointer — callers detect it
// with errors.As(err, &CorruptedDataError{}), which matches the value
// form only.
type CorruptedDataError struct{}

func (CorruptedDataError) Error() string {
	return "corrupted data"
}
