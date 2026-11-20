package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
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
// folder, which is created, readable only by its owner
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

// moveFrom moves a file, or a folder, from where it was to where it is,
// when it's there, and there's none where it is, which isn't an error when
// it isn't there
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
	return os.Rename(from, to)
}
