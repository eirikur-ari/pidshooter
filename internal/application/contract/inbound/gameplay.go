package inbound

// GamePlayConfig carries the user's intent for a game session.
type GamePlayConfig struct {
	Patterns    []string
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
}

// GamePlay is the inbound port: the contract the delivery adapter calls.
type GamePlay interface {
	Play(cfg GamePlayConfig) error
}
