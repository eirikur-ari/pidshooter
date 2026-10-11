package filestore

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
)

// currentConfigSchemaVersion is the schema version config files are written
// with and the newest one accepted when loading.
const currentConfigSchemaVersion = 1

// maxConfigFileSize is the largest config file, in bytes, that Load accepts.
const maxConfigFileSize = 1 << 20 // 1 MiB

// configEncoder encodes the content of a config file.
type configEncoder interface {
	// encodeYAML returns the YAML encoding of value.
	encodeYAML(value any) ([]byte, error)
}

// configFile persists run config to a YAML file.
type configFile struct {
	file    file
	encoder configEncoder
}

// gameEntry is the on-disk YAML representation of the persisted game
// settings. A nil field is omitted from the written file and, on
// read, means the key was absent.
type gameEntry struct {
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
	Game    gameEntry    `yaml:"game"`
}

// NewConfigFile constructs an outbound.ConfigStore that persists to the
// default per-user config path. It fails if the user's home directory
// cannot be resolved.
func NewConfigFile() (outbound.ConfigStore, error) {
	f, err := newFile("config.yaml", maxConfigFileSize)
	if err != nil {
		return nil, err
	}

	return &configFile{file: f, encoder: fileEncoder{}}, nil
}

// newConfigFileAt constructs an outbound.ConfigStore that persists to path.
func newConfigFileAt(path string) outbound.ConfigStore {
	return &configFile{file: file{path: path, maxSize: maxConfigFileSize}, encoder: fileEncoder{}}
}

// Load returns the persisted config. A missing file is reported as
// outbound.NotFoundError; a file that fails to decode is reported as
// outbound.CorruptedDataError. Any other failure, including a schema
// version mismatch, is returned unwrapped.
func (c *configFile) Load() (outbound.Config, error) {
	data, err := c.file.read()
	if err != nil {
		return outbound.Config{}, err
	}

	var cd configContent
	if len(data) > 0 {
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		if err := dec.Decode(&cd); err != nil {
			return outbound.Config{}, outbound.CorruptedDataError{Message: err.Error()}
		}
	}

	if err := schemaVersion(currentConfigSchemaVersion).validate("config file", cd.Version); err != nil {
		return outbound.Config{}, err
	}

	return toConfig(cd), nil
}

// Save atomically replaces the persisted config with config.
func (c *configFile) Save(config outbound.Config) error {
	data, err := c.encoder.encodeYAML(toConfigContent(config))
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return c.file.write(data)
}
