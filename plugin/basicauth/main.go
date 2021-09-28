package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	// Packages
	htpasswd "github.com/mutablelogic/go-server/pkg/htpasswd"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Passwords string `yaml:"passwords"`
	Groups    string `yaml:"groups"`
	Realm     string `yaml:"realm"`
	Admin     string `yaml:"admin"`
}

type basicauth struct {
	sync.Mutex
	Config
	*htpasswd.Htpasswd
	*htpasswd.Htgroups
	modtimep time.Time
	modtimeg time.Time
	dirty    bool
}

type handler struct {
	http.HandlerFunc
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultRealm   = "Authorization Required"
	defaultCharset = "UTF-8"
	defaultEncrypt = "bcrypt"
	deltaRead      = time.Second * 30
)

var (
	reBasicAuth = regexp.MustCompile(`^Basic\s+(.+)$`)
)

///////////////////////////////////////////////////////////////////////////////
// NEW

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	this := new(basicauth)

	// Load configuration
	if err := provider.GetConfig(ctx, &this.Config); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	}

	// Set absolute path
	if this.Config.Passwords != "" {
		if abspath, err := filepath.Abs(this.Config.Passwords); err != nil {
			provider.Print(ctx, "GetConfig: ", err)
			return nil
		} else {
			this.Config.Passwords = abspath
		}
	}

	// Set absolute path
	if this.Config.Groups != "" {
		if abspath, err := filepath.Abs(this.Config.Groups); err != nil {
			provider.Print(ctx, "GetConfig: ", err)
			return nil
		} else {
			this.Config.Groups = abspath
		}
	}

	// Set default realm
	if this.Realm == "" {
		this.Realm = defaultRealm
	}

	// Read in passwords
	if modtime, passwds, err := readpasswords(this.Config.Passwords, this.modtimep); err != nil {
		provider.Print(ctx, "Passwords: ", err)
		return nil
	} else {
		this.modtimep = modtime
		this.Htpasswd = passwds
	}

	// Read in groups
	if modtime, groups, err := readgroups(this.Config.Groups, this.modtimeg); err != nil {
		provider.Print(ctx, "Groups: ", err)
		return nil
	} else {
		this.modtimeg = modtime
		this.Htgroups = groups
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *basicauth) String() string {
	str := "<basicauth"
	if this.Realm != "" {
		str += fmt.Sprintf(" realm=%q", this.Realm)
	}
	if this.Htpasswd != nil {
		str += fmt.Sprint(" ", this.Htpasswd)
	}
	if this.Htgroups != nil {
		str += fmt.Sprint(" ", this.Htgroups)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "basicauth"
}

func (this *basicauth) Run(ctx context.Context, provider Provider) error {
	if err := this.addHandlers(ctx, provider); err != nil {
		return err
	}

	// Create ticker for re-reading password file
	ticker := time.NewTicker(deltaRead)
	defer ticker.Stop()

	// Run loop
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Write passwords and groups files
			if this.dirty {
				this.dirty = false
				if err := this.write(); err != nil {
					provider.Print(ctx, "Write: ", err)
				}
			}

			// Read passwords if path is set
			if this.Config.Passwords != "" {
				if modtime, passwds, err := readpasswords(this.Config.Passwords, this.modtimep); errors.Is(err, ErrNotModified) {
					// Nothing to do
				} else if err != nil {
					provider.Print(ctx, "Passwords: ", err)
				} else {
					provider.Printf(ctx, "Passwords: Changed: %q", this.Config.Passwords)
					this.Mutex.Lock()
					this.Htpasswd = passwds
					this.modtimep = modtime
					this.Mutex.Unlock()
				}
			}

			// Read groups if path is set
			if this.Config.Groups != "" {
				if modtime, groups, err := readgroups(this.Config.Groups, this.modtimeg); errors.Is(err, ErrNotModified) {
					// Nothing to do
				} else if err != nil {
					provider.Print(ctx, "Groups: ", err)
				} else {
					provider.Printf(ctx, "Groups: Changed: %q", this.Config.Groups)
					this.Mutex.Lock()
					this.Htgroups = groups
					this.modtimeg = modtime
					this.Mutex.Unlock()
				}
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func readpasswords(path string, modtime time.Time) (time.Time, *htpasswd.Htpasswd, error) {
	// Where we're using in-memory passwords, return a new one
	if path == "" {
		return time.Time{}, htpasswd.New(), nil
	}

	// Read or re-read passwords
	stat, err := os.Stat(path)
	if err != nil {
		return time.Time{}, nil, err
	} else if stat.Mode().IsRegular() == false {
		return time.Time{}, nil, ErrNotFound.With(path)
	} else if stat.ModTime() == modtime {
		return modtime, nil, ErrNotModified.With(path)
	}

	// Open file
	r, err := os.Open(path)
	if err != nil {
		return time.Time{}, nil, err
	}
	defer r.Close()

	// Read passwords
	if htpasswd, err := htpasswd.Read(r); err != nil {
		return time.Time{}, nil, err
	} else {
		return stat.ModTime(), htpasswd, nil
	}
}

func readgroups(path string, modtime time.Time) (time.Time, *htpasswd.Htgroups, error) {
	// Where we're using in-memory groups, return a new one
	if path == "" {
		return time.Time{}, htpasswd.NewGroups(), nil
	}

	// Read or re-read groups
	stat, err := os.Stat(path)
	if err != nil {
		return time.Time{}, nil, err
	} else if stat.Mode().IsRegular() == false {
		return time.Time{}, nil, ErrNotFound.With(path)
	} else if stat.ModTime() == modtime {
		return modtime, nil, ErrNotModified.With(path)
	}

	// Open file
	r, err := os.Open(path)
	if err != nil {
		return time.Time{}, nil, err
	}
	defer r.Close()

	// Read groups
	if htgroups, err := htpasswd.ReadGroups(r); err != nil {
		return time.Time{}, nil, err
	} else {
		return stat.ModTime(), htgroups, nil
	}
}

func (this *basicauth) write() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Where we're using in-memory passwords, don't create a new one
	if this.Config.Passwords != "" && this.Htpasswd != nil {
		var buf bytes.Buffer
		if err := this.Htpasswd.Write(&buf); err != nil {
			return err
		} else if err := ioutil.WriteFile(this.Config.Passwords, buf.Bytes(), 0600); err != nil {
			return err
		}
	}

	// Where we're using in-memory groups, don't create a new one
	if this.Config.Groups != "" && this.Htgroups != nil {
		var buf bytes.Buffer
		if err := this.Htgroups.Write(&buf); err != nil {
			return err
		} else if err := ioutil.WriteFile(this.Config.Groups, buf.Bytes(), 0600); err != nil {
			return err
		}
	}

	// Return success
	return nil
}
