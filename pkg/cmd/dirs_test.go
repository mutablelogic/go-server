package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func Test_Dirs_AppDir(t *testing.T) {
	base := t.TempDir()

	// The application's folder is created, readable only by its owner
	dir, err := appDir(func() (string, error) { return base, nil }, "filer")
	if err != nil {
		t.Fatal(err)
	} else if dir != filepath.Join(base, "filer") {
		t.Errorf("dir = %q", dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	} else if !info.IsDir() {
		t.Error("not a directory")
	} else if runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
		t.Errorf("mode = %v", info.Mode().Perm())
	}

	// And again, when it's there
	if _, err := appDir(func() (string, error) { return base, nil }, "filer"); err != nil {
		t.Error(err)
	}

	// An application without a name has no folder
	if _, err := appDir(func() (string, error) { return base, nil }, ""); err == nil {
		t.Error("expected an error")
	}
}

func Test_Dirs_UserDataDir(t *testing.T) {
	switch runtime.GOOS {
	case "darwin", "windows":
		want, err := os.UserConfigDir()
		if err != nil {
			t.Skip(err)
		}
		if got, err := userDataDir(); err != nil || got != want {
			t.Errorf("userDataDir() = %q, %v, want %q", got, err, want)
		}
	default:
		data := t.TempDir()
		t.Setenv("XDG_DATA_HOME", data)
		if got, err := userDataDir(); err != nil || got != data {
			t.Errorf("userDataDir() = %q, %v, want %q", got, err, data)
		}

		// One which isn't absolute is ignored, as XDG says
		home := t.TempDir()
		t.Setenv("XDG_DATA_HOME", "relative")
		t.Setenv("HOME", home)
		if got, err := userDataDir(); err != nil || got != filepath.Join(home, ".local", "share") {
			t.Errorf("userDataDir() = %q, %v", got, err)
		}
	}
}

func Test_Dirs_Global(t *testing.T) {
	g := &global{execName: "test-dirs"}
	data, err := g.DataDir()
	if err != nil {
		t.Skip(err)
	}
	defer os.Remove(data)
	cache, err := g.CacheDir()
	if err != nil {
		t.Skip(err)
	}
	defer os.Remove(cache)
	if filepath.Base(data) != "test-dirs" || filepath.Base(cache) != "test-dirs" {
		t.Errorf("data = %q, cache = %q", data, cache)
	} else if data == cache {
		t.Error("the data, and cache, folders are the same")
	}
}

func Test_Dirs_MoveFrom(t *testing.T) {
	dir := t.TempDir()
	from, to := filepath.Join(dir, "cache", "defaults.json"), filepath.Join(dir, "data", "defaults.json")
	if err := os.MkdirAll(filepath.Dir(from), 0o700); err != nil {
		t.Fatal(err)
	} else if err := os.MkdirAll(filepath.Dir(to), 0o700); err != nil {
		t.Fatal(err)
	}

	// Nothing to move isn't an error
	if err := moveFrom(from, to); err != nil {
		t.Error(err)
	}

	// It's moved, when there's none where it is
	if err := os.WriteFile(from, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	} else if err := moveFrom(from, to); err != nil {
		t.Fatal(err)
	} else if data, err := os.ReadFile(to); err != nil || string(data) != "old" {
		t.Errorf("moved = %q, %v", data, err)
	} else if _, err := os.Stat(from); !os.IsNotExist(err) {
		t.Error("it's still where it was")
	}

	// And isn't, when there's one where it is
	if err := os.WriteFile(from, []byte("older"), 0o600); err != nil {
		t.Fatal(err)
	} else if err := moveFrom(from, to); err != nil {
		t.Fatal(err)
	} else if data, _ := os.ReadFile(to); string(data) != "old" {
		t.Errorf("kept = %q", data)
	}
}
