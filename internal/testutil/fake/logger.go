package fake

// Logger is a test double for outbound.Logger.
type Logger struct {
	Warnings []string
}

func (l *Logger) Warn(msg string) {
	l.Warnings = append(l.Warnings, msg)
}
