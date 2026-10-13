package outbound

// NotFoundError indicates the requested resource does not exist.
type NotFoundError struct{}

func (NotFoundError) Error() string {
	return "not found"
}

// CorruptedDataError indicates persisted data was retrieved successfully
// but could not be parsed as valid data. Message is optional and only
// adds detail to the error text.
type CorruptedDataError struct {
	Message string
}

func (e CorruptedDataError) Error() string {
	if e.Message == "" {
		return "corrupted data"
	}
	return e.Message + ": corrupted data"
}
