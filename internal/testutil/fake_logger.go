package testutil

// FakeLogger is a test double for outbound.Logger.
type FakeLogger struct {
	Warned  []string // messages passed to Warn, in call order
	Errored []string // messages passed to Error, in call order
}

func (l *FakeLogger) Warn(message string) {
	l.Warned = append(l.Warned, message)
}

func (l *FakeLogger) Error(message string) {
	l.Errored = append(l.Errored, message)
}
