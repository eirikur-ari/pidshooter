package testutil

type FakeLogger struct {
	Warned  []string
	Errored []string
}

func (l *FakeLogger) Warn(message string) {
	l.Warned = append(l.Warned, message)
}

func (l *FakeLogger) Error(message string) {
	l.Errored = append(l.Errored, message)
}
