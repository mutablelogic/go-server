package folders

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type File struct {
	// Prefix
	Prefix string

	// Path to the file
	Path string

	// Hash of the file contents
	Hash string

	// Infomation about the file
	info fs.FileInfo
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	R_OK = 4
	W_OK = 2
	X_OK = 1
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFile(prefix, path string) (*File, error) {
	f := new(File)

	// Stat the file
	abspath := filepath.Join(prefix, path)
	info, err := os.Stat(abspath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound.Withf("not found: %q", filepath.Base(path))
	} else if err != nil {
		return nil, err
	} else if !info.Mode().IsRegular() {
		return nil, ErrBadParameter.Withf("Not a regular file: %q", info.Name())
	} else {
		f.Prefix = prefix
		f.Path = path
		f.info = info
	}

	// Obtain a hash of the file contents. Returns an empty string on error
	result := sha256.New()
	data, err := os.ReadFile(abspath)
	if err != nil {
		return nil, err
	}
	result.Write(data)

	// Set the file hash
	f.Hash = hex.EncodeToString(result.Sum(nil))

	// Return success
	return f, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f File) String() string {
	data, _ := json.MarshalIndent(f, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the body
func (f *File) Render() ([]byte, error) {
	return os.ReadFile(filepath.Join(f.Prefix, f.Path))
}

// Remove a file from a source folder
func (f *File) Remove() error {
	absfile := filepath.Join(f.Prefix, f.Path)
	return os.Remove(absfile)
}

// Copy a file from a source folder to a destination folder, making directories
// as needed
func (f *File) Copy(dest string, perm fs.FileMode) error {
	// Make destination directories
	absfile := filepath.Join(dest, f.Path)
	absdir := filepath.Dir(absfile)
	if err := os.MkdirAll(absdir, perm); err != nil {
		return err
	}

	// Open the source file
	r, err := os.Open(filepath.Join(f.Prefix, f.Path))
	if err != nil {
		return err
	}
	defer r.Close()

	// Create the destination file
	w, err := os.Create(absfile)
	if err != nil {
		return err
	}
	defer w.Close()

	// Set the mode on the destination file
	if err := os.Chmod(absfile, f.info.Mode()); err != nil {
		if err_ := os.Remove(absfile); err != nil {
			return errors.Join(err, err_)
		}
		return err
	}

	// Copy the file
	if _, err := io.Copy(w, r); err != nil {
		if err_ := os.Remove(absfile); err != nil {
			return errors.Join(err, err_)
		}
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns boolean value which indicates if a file is readable by current
// user
func isReadableFileAtPath(path string) error {
	return syscall.Access(path, R_OK)
}

// Returns boolean value which indicates if a file is writable by current
// user
func isWritableFileAtPath(path string) error {
	return syscall.Access(path, W_OK)
}

// Returns boolean value which indicates if a file is executable by current
// user
func isExecutableFileAtPath(path string) error {
	return syscall.Access(path, X_OK)
}
