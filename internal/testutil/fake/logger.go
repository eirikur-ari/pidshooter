package fake

// Logger is a test double for outbound.Logger.
type Logger struct {
	Warned  []string // messages passed to Warn, in call order
	Errored []string // messages passed to Error, in call order
}

func (l *Logger) Warn(msg string) {
	l.Warned = append(l.Warned, msg)
}

func (l *Logger) Error(msg string) {
	l.Errored = append(l.Errored, msg)
}
