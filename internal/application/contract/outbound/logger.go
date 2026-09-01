package outbound

// Logger reports diagnostics at varying severity.
type Logger interface {
	// Warn reports a failure the caller was able to continue past.
	Warn(msg string)
	// Error reports a more severe failure than Warn.
	Error(msg string)
}
