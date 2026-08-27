package inbound

// Config holds run parameters, sourced from persisted defaults and
// overridden by explicit input.
type Config struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}