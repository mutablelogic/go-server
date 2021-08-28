package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	htpasswd "github.com/djthorpe/go-server/pkg/htpasswd"
	router "github.com/djthorpe/go-server/pkg/httprouter"
	"github.com/djthorpe/go-server/pkg/provider"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Path  string `yaml:"path"`
	Realm string `yaml:"realm"`
	Admin string `yaml:"admin"`
}

type basicauth struct {
	sync.Mutex
	Config
	*htpasswd.Htpasswd
	modtime time.Time
	dirty   bool
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
	if this.Config.Path != "" {
		if abspath, err := filepath.Abs(this.Config.Path); err != nil {
			provider.Print(ctx, "GetConfig: ", err)
			return nil
		} else {
			this.Config.Path = abspath
		}
	}

	// Set default realm
	if this.Realm == "" {
		this.Realm = defaultRealm
	}

	// Read in passwords
	if passwds, err := this.read(); err != nil {
		provider.Print(ctx, "ReadPasswords: ", err)
		return nil
	} else {
		this.Htpasswd = passwds
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
	if this.Path != "" {
		str += fmt.Sprintf(" path=%q", this.Path)
	}
	if this.Htpasswd != nil {
		str += fmt.Sprint(" ", this.Htpasswd)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *basicauth) read() (*htpasswd.Htpasswd, error) {
	// Where we're using in-memory passwords, don't create a new one
	if this.Config.Path == "" {
		if this.Htpasswd == nil {
			this.Htpasswd = htpasswd.New()
		}
		return this.Htpasswd, nil
	}

	// Read or re-read passwords
	if stat, err := os.Stat(this.Config.Path); err != nil {
		return nil, err
	} else if stat.Mode().IsRegular() == false {
		return nil, ErrNotFound.With(this.Config.Path)
	} else if stat.ModTime() == this.modtime && this.Htpasswd != nil {
		return this.Htpasswd, ErrNotModified.With(this.Config.Path)
	} else {
		this.modtime = stat.ModTime()
	}

	// Open file
	r, err := os.Open(this.Config.Path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Read passwords
	return htpasswd.Read(r)
}

func (this *basicauth) write() error {
	// Where we're using in-memory passwords, don't create a new one
	if this.Config.Path == "" || this.Htpasswd == nil {
		return nil
	}

	// Lock for writing
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Perform writing and close
	w, err := os.Create(this.Config.Path)
	if err != nil {
		return err
	}
	defer w.Close()

	return this.Htpasswd.Write(w)
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
			if this.dirty {
				this.dirty = false
				if err := this.write(); err != nil {
					provider.Print(ctx, "WritePasswords: ", err)
				} else if stat, err := os.Stat(this.Config.Path); err != nil {
					provider.Print(ctx, "WritePasswords: ", err)
				} else {
					this.modtime = stat.ModTime()
					provider.Printf(ctx, "WritePasswords: Passwords changed: %q", this.Config.Path)
				}
			}
			if passwds, err := this.read(); errors.Is(err, ErrNotModified) {
				// Nothing to do
			} else if err != nil {
				provider.Print(ctx, "ReadPasswords: ", err)
			} else {
				provider.Printf(ctx, "ReadPasswords: Passwords changed: %q", this.Config.Path)
				this.Mutex.Lock()
				this.Htpasswd = passwds
				this.Mutex.Unlock()
			}
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

func (this *basicauth) AddMiddlewareFunc(ctx context.Context, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var auth bool
		var user string
		if token := r.Header.Get("Authorization"); token != "" {
			if basic := reBasicAuth.FindStringSubmatch(token); basic != nil {
				basic[1] = strings.TrimRight(basic[1], "=")
				if credentials, err := base64.RawStdEncoding.DecodeString(basic[1]); err == nil {
					user, auth = this.authenticated(credentials)
				}
			}
		}
		// Return authentication error
		if !auth {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q, charset=%q", this.Realm, defaultCharset))
			router.ServeError(w, http.StatusUnauthorized)
			return
		}

		// Add user and realm to context
		ctx := r.Context()
		if user != "" {
			ctx = provider.ContextWithAuth(ctx, user, map[string]interface{}{"realm": this.Realm})
		}

		// Handle
		h(w, r.Clone(ctx))
	}
}

func (this *basicauth) AddMiddleware(ctx context.Context, h http.Handler) http.Handler {
	return &handler{
		this.AddMiddlewareFunc(ctx, h.ServeHTTP),
	}
}

func (this *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.HandlerFunc(w, r)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// authenticated the user and returns true if the given credentials are authenticated
func (this *basicauth) authenticated(credentials []byte) (string, bool) {
	// Lock for reading
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if creds := strings.SplitN(string(credentials), ":", 2); len(creds) != 2 {
		return "", false
	} else if this.Htpasswd == nil {
		return creds[0], false
	} else {
		return creds[0], this.Htpasswd.Verify(creds[0], creds[1])
	}
}
