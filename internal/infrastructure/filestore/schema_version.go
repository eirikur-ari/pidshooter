package filestore

import "fmt"

// schemaVersion is the highest on-disk schema version a file-backed store
// can read.
type schemaVersion int

// validate reports an error naming fileType if version is negative or newer
// than v.
func (v schemaVersion) validate(fileType string, version int) error {
	switch {
	case version < 0:
		return fmt.Errorf("%s schema version %d is invalid", fileType, version)
	case version > int(v):
		return fmt.Errorf("%s schema version %d is newer than the %d this build supports", fileType, version, v)
	}
	return nil
}
