package filestore

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// configFile persists run config to a YAML file.
type configFile struct {
	file file
}

// configEntry is the on-disk YAML representation of the persisted game
// mode config. A nil field is omitted from the written file and, on
// read, means the key was absent.
type configEntry struct {
	ConfirmMode *bool    `yaml:"confirm_mode,omitempty"`
	Speed       *float64 `yaml:"speed,omitempty"`
	TimeLimit   *int     `yaml:"time_limit,omitempty"`
}

// processEntry is the on-disk YAML representation of the persisted
// process-discovery config. A nil field is omitted from the written file
// and, on read, means the key was absent.
type processEntry struct {
	IncludeRoot *bool `yaml:"include_root,omitempty"`
}

// configContent is the on-disk YAML representation of the persisted run
// config.
type configContent struct {
	// Version identifies configContent's shape. Bump currentConfigSchemaVersion
	// whenever the shape changes.
	Version int          `yaml:"version"`
	Mode    string       `yaml:"mode"`
	Process processEntry `yaml:"process"`
	Game    configEntry  `yaml:"game"`
}

// currentConfigSchemaVersion is the schema version this build of
// pidshooter reads and writes. See configContent.Version.
const currentConfigSchemaVersion = 1

// maxConfigFileSize bounds the config file size accepted before parsing.
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

// Load reads and decodes the persisted run config. A missing file is
// reported as outbound.NotFoundError; an empty file is treated the same as
// one with every field absent. A non-empty file that fails to decode —
// including one containing a key this build doesn't recognize — or that
// exceeds maxConfigFileSize, is reported as outbound.CorruptedDataError. A
// file written by a newer, unrecognized schema version (e.g. by a newer
// build, on a downgrade) is returned unwrapped rather than as
// outbound.CorruptedDataError. Any other read failure (e.g. a permission
// error) is also returned unwrapped.
func (c *configFile) Load() (outbound.ConfigStoreResult, error) {
	data, err := c.file.read()
	if err != nil {
		return outbound.ConfigStoreResult{}, err
	}

	var cd configContent
	if len(data) > 0 {
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		if err := dec.Decode(&cd); err != nil {
			return outbound.ConfigStoreResult{}, outbound.CorruptedDataError{Message: err.Error()}
		}
	}

	if cd.Version > currentConfigSchemaVersion {
		return outbound.ConfigStoreResult{}, fmt.Errorf("config file schema version %d is newer than the %d this build supports", cd.Version, currentConfigSchemaVersion)
	}

	return toConfigStoreResult(cd), nil
}

// Save encodes and persists the run config, replacing any previously
// persisted config atomically, so a crash or kill mid-write can never
// leave a truncated or partial file behind.
func (c *configFile) Save(result outbound.ConfigStoreResult) error {
	data, err := yaml.Marshal(toConfigContent(result))
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return c.file.write(data)
}
