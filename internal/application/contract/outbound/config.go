package outbound

// Mode identifies which run mode pidshooter defaults to.
type Mode string

// The possible Mode values.
const (
	ModeGame  Mode = "game"
	ModeLucky Mode = "lucky"
	ModeList  Mode = "list"
)

// GameConfig is a user's saved game mode config. A nil field means that
// value was not present in the persisted data.
type GameConfig struct {
	ConfirmMode *bool
	Speed       *float64
	TimeLimit   *int
}

// ProcessConfig is a user's saved process-discovery config. A nil field
// means that value was not present in the persisted data.
type ProcessConfig struct {
	IncludeRoot *bool
}

// Config is a user's saved run config.
type Config struct {
	Mode    Mode
	Process ProcessConfig
	Game    GameConfig
}

// ConfigStore persists and retrieves run config.
type ConfigStore interface {
	// Load returns the persisted config. If none has been persisted yet,
	// it returns an empty Config and a NotFoundError. If the persisted
	// data exists but cannot be parsed, it returns an empty Config and a
	// CorruptedDataError.
	Load() (Config, error)
	// Save persists config, overwriting any previously persisted config.
	Save(config Config) error
}
