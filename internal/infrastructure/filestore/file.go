package filestore

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/outbound"
	"github.com/eirikur-ari/pidshooter/internal/infrastructure/fsutil"
)

// file handles the I/O shared by every file-backed store in this package:
// resolving the default per-user path, bounded reads mapped to outbound's
// not-found/corrupted-data errors, and atomic writes.
type file struct {
	path    string
	maxSize int
}

// newFile resolves filename under the user's default per-user config
// directory and returns a file bounded to maxSize bytes. It fails if the
// user's home directory cannot be resolved.
func newFile(filename string, maxSize int) (file, error) {
	dir, err := fsutil.ConfigDir("pidshooter")
	if err != nil {
		return file{}, err
	}

	return file{path: filepath.Join(dir, filename), maxSize: maxSize}, nil
}

// read returns the file's raw bytes. A missing file is reported as
// outbound.NotFoundError; a file over maxSize is reported as
// outbound.CorruptedDataError. Any other read failure (e.g. a permission
// error) is returned unwrapped.
func (f file) read() ([]byte, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, outbound.NotFoundError{}
		}
		return nil, err
	}
	if len(data) > f.maxSize {
		return nil, outbound.CorruptedDataError{
			Message: fmt.Sprintf("file is %d bytes, over the %d byte limit", len(data), f.maxSize),
		}
	}
	return data, nil
}

// write persists data to the file's path, replacing any previously
// persisted content atomically, so a crash or kill mid-write can never
// leave a truncated or partial file behind.
func (f file) write(data []byte) error {
	return fsutil.WriteFileAtomic(f.path, data, 0600)
}
