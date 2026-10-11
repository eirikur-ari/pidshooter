package outbound

// Logger logs warning and error messages.
type Logger interface {
	// Warn logs message as a warning.
	Warn(message string)
	// Error logs message as an error.
	Error(message string)
}
