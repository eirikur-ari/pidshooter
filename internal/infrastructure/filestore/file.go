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

// file is a single persisted file with a limit on how large it may be.
type file struct {
	path    string
	maxSize int
}

// newFile returns a file named filename in the per-user config directory,
// limited to maxSize bytes. It fails if the user's home directory cannot be
// resolved.
func newFile(filename string, maxSize int) (file, error) {
	dir, err := fsutil.ConfigDir("pidshooter")
	if err != nil {
		return file{}, err
	}

	return file{path: filepath.Join(dir, filename), maxSize: maxSize}, nil
}

// read returns the file's content. A missing file is reported as
// outbound.NotFoundError and a file over maxSize as
// outbound.CorruptedDataError. Any other failure is returned unwrapped.
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

// write atomically replaces the file's content with data.
func (f file) write(data []byte) error {
	return fsutil.WriteFileAtomic(f.path, data, 0600)
}
