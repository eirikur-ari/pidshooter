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
// config. A nil field means that value was not present in the persisted
// data.
type GameConfig struct {
	ConfirmMode *bool
	Speed       *float64
	TimeLimit   *int
}

// ProcessConfig is the persistence representation of a user's saved
// process-discovery config. A nil field means that value was not present
// in the persisted data.
type ProcessConfig struct {
	IncludeRoot *bool
}

// ConfigStoreResult is the persistence representation of a user's saved run
// config.
type ConfigStoreResult struct {
	Mode    Mode
	Process ProcessConfig
	Game    GameConfig
}

// ConfigStore is the outbound port for persisting and retrieving run config.
type ConfigStore interface {
	// Load returns the persisted config. If none has been persisted yet,
	// it returns an empty ConfigStoreResult and a NotFoundError. If the
	// persisted data exists but cannot be parsed, it returns an empty
	// ConfigStoreResult and a CorruptedDataError.
	Load() (ConfigStoreResult, error)
	// Save persists result, overwriting any previously persisted config.
	Save(result ConfigStoreResult) error
}
