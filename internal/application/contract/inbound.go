package contract

// Config carries the user's intent for a game session.
type Config struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// GameService is the inbound port: the contract the delivery adapter calls.
type GameService interface {
	Play(cfg Config) error
}
