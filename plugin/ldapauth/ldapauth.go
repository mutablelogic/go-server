package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	// Modules
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Addr        string        `yaml:"url"`
	User        string        `yaml:"user"`
	Password    string        `yaml:"password"`
	Expiry      time.Duration `yaml:"expiry"`
	GroupFilter string        `yaml:"groupfilter"`
	UserFilter  string        `yaml:"userfilter"`
	BaseDN      string        `yaml:"basedn"`
	Fields      []string      `yaml:"attrs"`
}

type plugin struct {
	Config
	*Credentials
	*JWT

	userfilter  string
	groupfilter string
	basedn      string
	fields      []string
	secret      []byte
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	TOKEN_NAME = "token"
	TOKEN_PATH = "/"
	KEY_SECRET = "LDAP_SECRET"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the plugin
func New(ctx context.Context, provider Provider) Plugin {
	this := new(plugin)

	// Load configuration
	cfg := Config{}
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, "GetConfig: ", err)
		return nil
	}

	// Create credentials
	if cred, err := NewCredentials(cfg.Addr, cfg.User, cfg.Password, cfg.Expiry); err != nil {
		provider.Print(ctx, "NewCredentials: ", err)
		return nil
	} else {
		this.Credentials = cred
	}

	// Set filter, basedn and fields
	this.userfilter = cfg.UserFilter
	this.groupfilter = cfg.GroupFilter
	this.basedn = cfg.BaseDN
	this.fields = cfg.Fields

	// Create JWT
	if jwt := NewJWT(this.getSecret); jwt == nil {
		provider.Print(ctx, "NewJWT: ", ErrInternalAppError)
		return nil
	} else {
		this.JWT = jwt
	}

	// Return success
	return this
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *plugin) String() string {
	str := "<ldapauth"
	if basedn := this.basedn; basedn != "" {
		str += fmt.Sprintf(" basedn=%q", this.basedn)
	}
	if len(this.fields) > 0 {
		str += fmt.Sprintf(" fields=%q", this.fields)
	}
	if filter := this.userfilter; filter != "" {
		str += fmt.Sprintf(" userfilter=%q", filter)
	}
	if filter := this.groupfilter; filter != "" {
		str += fmt.Sprintf(" groupfilter=%q", filter)
	}
	str += fmt.Sprint(" ", this.Credentials)
	str += fmt.Sprint(" ", this.JWT)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PLUGIN

func Name() string {
	return "ldapauth"
}

func Usage(output io.Writer) {
	fmt.Fprintf(output, "\nUsage of %q:\n", "ldapauth")
	fmt.Fprintf(output, "  url:\n")
	fmt.Fprintf(output, "\t  URL of LDAP server, should start with ldap:// or ldaps://\n")
	fmt.Fprintf(output, "  user:\n")
	fmt.Fprintf(output, "\t  LDAP administrator bind user (bind dn)\n")
	fmt.Fprintf(output, "  password:\n")
	fmt.Fprintf(output, "\t  LDAP administrator bind password\n")
	fmt.Fprintf(output, "  expiry:\n")
	fmt.Fprintf(output, "\t  Expiry time for authenticated user (duration)\n")
	fmt.Fprintf(output, "  userfilter:\n")
	fmt.Fprintf(output, "\t  Filter term used in search for user in LDAP (string)\n")
	fmt.Fprintf(output, "  groupfilter:\n")
	fmt.Fprintf(output, "\t  Filter term used in search for group in LDAP (string)\n")
	fmt.Fprintf(output, "  basedn:\n")
	fmt.Fprintf(output, "\t  Base DN used to locate user in LDAP (string)\n")
	fmt.Fprintf(output, "  attr:\n")
	fmt.Fprintf(output, "\t  Fields that should be returned by a search (array)\n")
}

func (this *plugin) Run(ctx context.Context, provider Provider) error {
	if err := this.addHandlers(ctx, provider); err != nil {
		return err
	}

	// Wait until done
	<-ctx.Done()

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *plugin) Ping() error {
	return this.Credentials.Ping()
}

func (this *plugin) Search(params url.Values, password string) (url.Values, error) {
	return this.Credentials.Search(this.userfilter, this.basedn, this.fields, params, password)
}

func (this *plugin) ListUsers(limit uint) ([]url.Values, error) {
	return this.Credentials.List(this.userfilter, this.basedn, this.fields, limit)
}

func (this *plugin) ListGroups(limit uint) ([]url.Values, error) {
	return this.Credentials.List(this.groupfilter, this.basedn, this.fields, limit)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *plugin) setToken(ctx context.Context, w http.ResponseWriter, attributes url.Values) error {
	// Create a JWT value
	token, _, err := this.JWT.Token(ctx, attributes)
	if err != nil {
		return err
	}

	// Set the token
	http.SetCookie(w, &http.Cookie{
		Name:    TOKEN_NAME,
		Path:    TOKEN_PATH,
		Value:   token,
		Expires: time.Now().Add(this.Credentials.Expiry()),
	})

	// Return success
	return nil
}

func (this *plugin) validate(w http.ResponseWriter, req *http.Request) (url.Values, time.Time) {
	// Obtain the session token
	token, err := req.Cookie(TOKEN_NAME)
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return nil, time.Time{}
	}

	// Parse the token
	attributes, expiry, err := this.JWT.Parse(req.Context(), token.Value)
	if err == ErrInvalidCredentials {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return nil, time.Time{}
	} else if err == ErrExpiredCredentials {
		if expiry.After(time.Now().Add(this.Credentials.Expiry())) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return nil, time.Time{}
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil, time.Time{}
	}

	// Return attributes
	return attributes, expiry
}

func (this *plugin) getSecret(ctx context.Context) ([]byte, error) {
	// TODO: Retrieve secret from provider or generate a new one
	secret := this.secret
	if isValidSecret(secret) == false {
		var err error
		if secret, err = this.rotateSecret(ctx); err != nil {
			return nil, err
		}
	}
	return secret, nil
}

func (this *plugin) rotateSecret(ctx context.Context) ([]byte, error) {
	secret := NewSecret()

	// TODO: Set the secret
	this.secret = secret

	// Success
	return secret, nil
}
