package filestore

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// configFile persists run defaults to a YAML file.
type configFile struct {
	file file
}

// configEntry is the on-disk YAML representation of the persisted game
// mode defaults.
type configEntry struct {
	ConfirmMode bool    `yaml:"confirm_mode"`
	Speed       float64 `yaml:"speed"`
	TimeLimit   int     `yaml:"time_limit"`
	IncludeRoot bool    `yaml:"include_root"`
}

// configContent is the on-disk YAML representation of the persisted run
// defaults.
type configContent struct {
	// Version identifies configContent's shape. Bump
	// currentConfigSchemaVersion and add a migration step in Load whenever
	// a field is renamed or removed — otherwise yaml.v3 silently drops or
	// zero-fills old data on the next Save.
	Version int         `yaml:"version"`
	Mode    string      `yaml:"mode"`
	Game    configEntry `yaml:"game"`
}

// currentConfigSchemaVersion is the schema version this build of
// pidshooter reads and writes. See configContent.Version.
const currentConfigSchemaVersion = 1

// maxConfigFileSize bounds how large a config file Load will accept
// before parsing.
const maxConfigFileSize = 1 << 20 // 1 MiB

// NewConfigFile constructs an outbound.ConfigStore that persists to the
// default per-user config path. It fails if the user's home directory
// cannot be resolved.
func NewConfigFile() (outbound.ConfigStore, error) {
	f, err := newFile("config.yaml", maxConfigFileSize)
	if err != nil {
		return nil, err
	}

	return &configFile{file: f}, nil
}

// newConfigFileAt constructs an outbound.ConfigStore that persists to the
// given path, without touching the user's default config location.
func newConfigFileAt(path string) outbound.ConfigStore {
	return &configFile{file: file{path: path, maxSize: maxConfigFileSize}}
}

// Load reads and decodes the persisted run defaults. A missing file is
// reported as outbound.NotFoundError; a file that exists but fails to
// decode, or exceeds maxConfigFileSize, is reported as
// outbound.CorruptedDataError. A file written by a newer, unrecognized
// schema version (e.g. by a newer build, on a downgrade) is returned
// unwrapped rather than as outbound.CorruptedDataError. Any other read
// failure (e.g. a permission error) is also returned unwrapped.
func (c *configFile) Load() (outbound.DefaultConfig, error) {
	data, err := c.file.read()
	if err != nil {
		return outbound.DefaultConfig{}, err
	}

	var cd configContent
	if err := yaml.Unmarshal(data, &cd); err != nil {
		return outbound.DefaultConfig{}, outbound.CorruptedDataError{Message: err.Error()}
	}

	if cd.Version > currentConfigSchemaVersion {
		return outbound.DefaultConfig{}, fmt.Errorf("config file schema version %d is newer than the %d this build supports", cd.Version, currentConfigSchemaVersion)
	}

	return toDefaultConfig(cd), nil
}

// Save encodes and persists the run defaults, replacing any previously
// persisted defaults atomically, so a crash or kill mid-write can never
// leave a truncated or partial file behind.
func (c *configFile) Save(defaults outbound.DefaultConfig) error {
	data, err := yaml.Marshal(toConfigContent(defaults))
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return c.file.write(data)
}
