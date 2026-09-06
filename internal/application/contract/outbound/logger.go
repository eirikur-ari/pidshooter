package outbound

// Logger reports diagnostics at varying severity.
type Logger interface {
	// Warn reports anything that is not an error but still needs attention.
	Warn(msg string)
	// Error reports a system or application specific error that needs attention.
	Error(msg string)
}
