package filestore

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// fileEncoder encodes values in the formats that files are persisted in.
type fileEncoder struct{}

// encodeYAML returns the YAML encoding of value.
func (fileEncoder) encodeYAML(value any) ([]byte, error) {
	return yaml.Marshal(value)
}

// encodeJSON returns the JSON encoding of value, indented with two spaces.
func (fileEncoder) encodeJSON(value any) ([]byte, error) {
	return json.MarshalIndent(value, "", "  ")
}
