package driving

// Config carries the user's intent for a game session.
type Config struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// GameService is the driving port: the contract the CLI adapter calls.
type GameService interface {
	Play(cfg Config) error
}
