package outbound

// Logger reports leveled diagnostic messages.
type Logger interface {
	// Warn reports msg as a warning.
	Warn(msg string)
	// Error reports msg as an error.
	Error(msg string)
}
