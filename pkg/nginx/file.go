package nginx

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type File struct {
	// Path to the file
	Path string

	// Info for the file
	Info fs.FileInfo

	// Enabled indicates the configuration file is enabled
	Enabled bool
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f *File) String() string {
	str := "<file"
	if enabled := f.Enabled; enabled {
		str += " enabled"
	}
	if hash := f.Hash(); hash != "" {
		str += fmt.Sprintf(" hash=%q", hash)
	}
	if ext := f.Ext(); ext != "" {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	if name := f.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if enabled_base := f.EnabledBase(); enabled_base != "" {
		str += fmt.Sprintf(" enabled_base=%q", enabled_base)
	}
	if available_base := f.AvailableBase(); available_base != "" {
		str += fmt.Sprintf(" available_base=%q", available_base)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the hash of the filename
func (f *File) Hash() string {
	if f.Path == "" {
		return ""
	} else {
		sum := md5.Sum([]byte(f.Path))
		return hex.EncodeToString(sum[:])
	}
}

// Return the file extension
func (f *File) Ext() string {
	if f.Info != nil {
		return filepath.Ext(f.Info.Name())
	} else if f.Path != "" {
		return filepath.Ext(f.Path)
	} else {
		return ""
	}
}

// Return the last element of the path by name
func (f *File) Name() string {
	var filename string
	if f.Info != nil {
		filename = f.Info.Name()
	} else if f.Path != "" {
		filename = filepath.Base(f.Path)
	} else {
		return ""
	}
	return strings.TrimSuffix(filename, f.Ext())
}

// Return the "enabled" name of the file
func (f *File) AvailableBase() string {
	if name, ext := f.Name(), f.Ext(); name == "" {
		return ""
	} else {
		return name + ext
	}
}

// Return the "enabled" name of the file
func (f *File) EnabledBase() string {
	if name, hash, ext := f.Name(), f.Hash(), f.Ext(); name == "" || hash == "" {
		return ""
	} else {
		return strings.Trim(name, "-") + "-" + hash + ext
	}
}
