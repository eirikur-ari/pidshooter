package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// ConfigDir returns the per-user configuration directory for appName,
// resolved as $XDG_CONFIG_HOME/<appName> when $XDG_CONFIG_HOME is set to
// an absolute path, falling back to $HOME/.config/<appName> otherwise.
// It returns an error rather than falling back to another location if
// the user's home directory cannot be resolved.
func ConfigDir(appName string) (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(xdg) {
		return filepath.Join(xdg, appName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user home directory: %w", err)
	}

	return filepath.Join(home, ".config", appName), nil
}

// WriteFileAtomic writes data to path with permissions perm, atomically
// replacing any existing file so a crash or kill during the write can
// never leave path holding truncated or partial content. It creates
// path's parent directory (and any missing ancestors) if it does not
// already exist, and sets it to mode 0700 either way — including
// tightening it down if it already existed with looser permissions.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("could not create directory %s: %w", dir, err)
	}

	if err := os.Chmod(dir, 0700); err != nil {
		return fmt.Errorf("could not set permissions on %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("could not create temp file in %s: %w", dir, err)
	}

	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("could not write temp file %s: %w", tmp.Name(), err)
	}

	if err = tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("could not set permissions on %s: %w", tmp.Name(), err)
	}

	preserveOwner(tmp.Name(), path)

	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("could not sync temp file %s: %w", tmp.Name(), err)
	}

	if err = tmp.Close(); err != nil {
		return fmt.Errorf("could not close temp file %s: %w", tmp.Name(), err)
	}

	if err = os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("could not rename %s to %s: %w", tmp.Name(), path, err)
	}

	// Best-effort: fsync the directory entry too, so the rename itself is
	// durable across a crash, not just the file content. Not all platforms
	// support syncing a directory handle (notably Windows), so a failure
	// here is not reported — the rename already happened, and Rename's
	// atomicity still guarantees a reader never sees a torn file.
	if dirFile, dirErr := os.Open(dir); dirErr == nil {
		_ = dirFile.Sync()
		_ = dirFile.Close()
	}

	return nil
}

// preserveOwner best-effort chowns the replacement file to the invoking
// user — preferring sudo's own record of who that is over the file being
// replaced — so an elevated write doesn't leave it owned by the writing
// process instead.
func preserveOwner(tmpPath, path string) {
	uid, gid, ok := sudoOwner()
	if !ok {
		uid, gid, ok = existingOwner(path)
	}
	if !ok {
		return
	}
	_ = os.Chown(tmpPath, uid, gid)
}

// sudoOwner reports the user sudo recorded as having invoked the program,
// if SUDO_UID and SUDO_GID are both set and valid.
func sudoOwner() (uid, gid int, ok bool) {
	uid, uidErr := strconv.Atoi(os.Getenv("SUDO_UID"))
	gid, gidErr := strconv.Atoi(os.Getenv("SUDO_GID"))
	return uid, gid, uidErr == nil && gidErr == nil
}

// existingOwner reports a file's current owner, if it exists.
func existingOwner(path string) (uid, gid int, ok bool) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, false
	}
	return int(stat.Uid), int(stat.Gid), true
}
