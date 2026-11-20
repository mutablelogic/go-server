package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// DataDir returns the application's folder of the user's data, which is
// kept, which is created, readable only by its owner
func (g *global) DataDir() (string, error) {
	return appDir(userDataDir, g.execName)
}

// CacheDir returns the application's folder of the user's cache, which can
// be cleared, which is created, readable only by its owner
func (g *global) CacheDir() (string, error) {
	return appDir(os.UserCacheDir, g.execName)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// appDir returns the folder of the application, of its name, in the base
// folder, which is created, readable only by its owner, which one which is
// there already is made, as MkdirAll doesn't change it, on Unix, as its
// permissions are what keeps what's in it private
func appDir(base func() (string, error), name string) (string, error) {
	if name == "" {
		return "", errors.New("the application has no name")
	}
	dir, err := base()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	} else if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0o700); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// userDataDir returns the user's folder of data, which is kept, which Go
// has no function of: ~/Library/Application Support on macOS, %AppData% on
// Windows, which are os.UserConfigDir's, and $XDG_DATA_HOME, when it's an
// absolute path, or ~/.local/share, elsewhere
func userDataDir() (string, error) {
	switch runtime.GOOS {
	case "darwin", "windows":
		return os.UserConfigDir()
	}
	if dir := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(dir) {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

// moveFrom moves a file from where it was to where it is, when it's there,
// and there's none where it is, which isn't an error when it isn't there.
// When they're on different filesystems, which it can't be renamed across,
// it's copied, and then removed, once the copy is where it is.
func moveFrom(from, to string) error {
	if _, err := os.Stat(from); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	} else if _, err := os.Stat(to); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.Rename(from, to); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return err
	}
	return copyFrom(from, to)
}

// copyFrom copies a file to where it is, of a temporary file beside it,
// which is linked there, so one which is there isn't replaced, and then
// removes it from where it was
func copyFrom(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()
	info, err := src.Stat()
	if err != nil {
		return err
	} else if !info.Mode().IsRegular() {
		return fmt.Errorf("%q isn't a file, which is copied", from)
	}
	temp, err := os.CreateTemp(filepath.Dir(to), "."+filepath.Base(to)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if _, err := io.Copy(temp, src); err != nil {
		temp.Close()
		return err
	} else if err := temp.Chmod(info.Mode().Perm()); err != nil {
		temp.Close()
		return err
	} else if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	} else if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Link(temp.Name(), to); errors.Is(err, fs.ErrExist) {
		// One which is there, such as another's, is kept
		return nil
	} else if err != nil {
		return err
	}
	return os.Remove(from)
}
