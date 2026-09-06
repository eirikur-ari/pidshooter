package fake

// Logger is a test double for outbound.Logger.
type Logger struct {
	Warnings []string
	Errors   []string
}

func (l *Logger) Warn(msg string) {
	l.Warnings = append(l.Warnings, msg)
}

func (l *Logger) Error(msg string) {
	l.Errors = append(l.Errors, msg)
}
