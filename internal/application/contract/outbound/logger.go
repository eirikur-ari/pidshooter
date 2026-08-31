package outbound

// Logger reports non-fatal diagnostics.
type Logger interface {
	// Warn reports a failure the caller was able to continue past.
	Warn(msg string)
}
