package game

// Config holds the gameplay parameters for a session.
type Config struct {
	Confirm   bool
	Speed     float64
	TimeLimit int
}

// defaultSpeed and defaultTimeLimit are the values DefaultConfig returns —
// this domain's baseline gameplay parameters.
const (
	defaultSpeed     = 2.0
	defaultTimeLimit = 30
)

// DefaultConfig returns this domain's default gameplay Config: confirm
// mode off, defaultSpeed, and defaultTimeLimit.
func DefaultConfig() Config {
	return Config{Confirm: false, Speed: defaultSpeed, TimeLimit: defaultTimeLimit}
}
