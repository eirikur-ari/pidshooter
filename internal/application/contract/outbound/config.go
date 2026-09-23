package outbound

// Mode identifies which run mode pidshooter defaults to.
type Mode string

// Mode's possible values.
const (
	ModeGame Mode = "game"
	ModeYolo Mode = "yolo"
	ModeList Mode = "list"
)

// GameConfig is the persistence representation of a user's saved game mode
// defaults.
type GameConfig struct {
	ConfirmMode bool
	Speed       float64
	TimeLimit   int
	IncludeRoot bool
}

// DefaultConfig is the persistence representation of a user's saved run
// defaults.
type DefaultConfig struct {
	Mode Mode
	Game GameConfig
}

// ConfigStore is the outbound port for persisting and retrieving run defaults.
type ConfigStore interface {
	// Load returns the persisted defaults. If none have been persisted yet,
	// it returns empty DefaultConfig and a NotFoundError. If the
	// persisted data exists but cannot be parsed, it returns empty
	// DefaultConfig and a CorruptedDataError.
	Load() (DefaultConfig, error)
	// Save persists defaults, overwriting any previously persisted defaults.
	Save(defaults DefaultConfig) error
}
