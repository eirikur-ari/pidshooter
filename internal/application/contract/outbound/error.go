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
// form only, regardless of Message. Message is optional and only adds
// detail to Error()'s text — it plays no part in matching.
type CorruptedDataError struct {
	Message string
}

func (e CorruptedDataError) Error() string {
	if e.Message == "" {
		return "corrupted data"
	}
	return e.Message + ": corrupted data"
}
