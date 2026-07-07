package contract

// Config carries the user's intent for a game session.
type Config struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// GamePlay is the inbound port: the contract the entrypoint adapter calls.
type GamePlay interface {
	Play(cfg Config) error
}
