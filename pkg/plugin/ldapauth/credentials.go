package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	ldap "github.com/go-ldap/ldap/v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Credentials struct {
	sync.Mutex
	conn           *ldap.Conn
	ctime          time.Time
	addr           *url.URL
	user, password string
	expiry         time.Duration
}

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Reconnect with the LDAP server every five minutes to prevent stale
	// connections, and reconnect if network error after a 200ms gap
	LDAP_RECONNECT_DELTA = 5 * time.Minute
	LDAP_BACKOFF_DELTA   = 200 * time.Millisecond
	LDAP_RETRY           = 3
)

var (
	ErrUserNotFound       = fmt.Errorf("User not found")
	ErrUserAmbiguity      = fmt.Errorf("More than one user found")
	ErrInvalidCredentials = fmt.Errorf("Invalid Credentials")
	ErrExpiredCredentials = fmt.Errorf("Expired Credentials")
	ErrTooManyRetries     = fmt.Errorf("Too many retries")
)

/////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCredentials(addr, user, password string, expiry time.Duration) (*Credentials, error) {
	this := new(Credentials)
	if u, err := url.Parse(addr); err != nil {
		return nil, err
	} else {
		this.addr = u
		this.user = user
		this.password = password
		this.expiry = expiry
	}

	// Return success
	return this, nil
}

/////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *Credentials) Expiry() time.Duration {
	return this.expiry
}

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Ping connects to the remote server
func (this *Credentials) Ping() error {
	this.Lock()
	defer this.Unlock()

	// Get a connection
	if _, err := this.getConn(); err != nil {
		return err
	} else {
		return nil
	}
}

// Search will lookup an entity, and can bind for authentication if password is provided.
// Returns ErrUserNotFound, ErrUserAmbiguity in cases where user credentials are incorrect.
// Uses a filter and basedn to search and fields for the attributes to return.
// TODO: Need to handle disabled users (should they be part of a group?)
func (this *Credentials) Search(filter, basedn string, fields []string, params url.Values, password string) (url.Values, error) {
	this.Lock()
	defer this.Unlock()

	// Call search now locked
	return this.search(filter, basedn, fields, params, password, 0)
}

func (this *Credentials) search(filter, basedn string, fields []string, params url.Values, password string, depth uint) (url.Values, error) {
	// Return if we have tried to search too many times
	if depth >= LDAP_RETRY {
		return nil, ErrTooManyRetries
	}
	// Get a connection
	conn, err := this.getConn()
	if err != nil {
		return nil, err
	}

	// Bind and if there is a network error then retry
	if err := conn.Bind(this.user, this.password); err != nil {
		if isBindError(err) {
			return nil, ErrInvalidCredentials
		} else if isNetworkError(err) {
			this.ctime = time.Time{}
			time.Sleep(time.Duration(depth) * LDAP_BACKOFF_DELTA)
			return this.search(filter, basedn, fields, params, password, depth+1)
		} else {
			return nil, err
		}
	}
	// Fill in details with the filter template, for example "${uid}" or "$uid" is replaced.
	filter, err = getFilter(filter, params)
	if err != nil {
		return nil, err
	}

	// Perform search, check for various errors
	request := ldap.NewSearchRequest(basedn, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, filter, fields, nil)
	response, err := conn.Search(request)
	if err != nil {
		return nil, err
	} else if len(response.Entries) == 0 {
		return nil, ErrUserNotFound
	} else if len(response.Entries) > 1 {
		return nil, ErrUserAmbiguity
	}

	// Fill in the attributes from the response
	attrs := url.Values(make(map[string][]string))
	for _, attr := range response.Entries[0].Attributes {
		attrs[attr.Name] = attr.Values
	}

	// If Password is given, then bind
	if password != "" {
		if err := conn.Bind(response.Entries[0].DN, password); isBindError(err) {
			return attrs, ErrInvalidCredentials
		} else if err != nil {
			return attrs, err
		}
	}

	// Success
	return attrs, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Credentials) String() string {
	str := "<ldapauth.credentials"
	str += fmt.Sprintf(" url=%q", this.addr)
	str += fmt.Sprintf(" user=%q", this.user)
	str += fmt.Sprintf(" expiry=%v", this.expiry)
	return str + ">"
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func getFilter(template string, params url.Values) (string, error) {
	var err error
	filter := os.Expand(template, func(key string) string {
		if value, exists := params[key]; exists && len(value) > 0 {
			return value[0]
		} else {
			err = fmt.Errorf("%w: %v", ErrBadParameter, strconv.Quote(key))
			return key
		}
	})
	return filter, err
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	} else if ldaperr, ok := err.(*ldap.Error); ok {
		switch ldaperr.ResultCode {
		case ldap.ErrorNetwork, ldap.LDAPResultBusy, ldap.LDAPResultUnavailable:
			return true
		}
	}
	// By default, return false
	return false
}

func isBindError(err error) bool {
	if err == nil {
		return false
	} else if ldaperr, ok := err.(*ldap.Error); ok {
		switch ldaperr.ResultCode {
		case ldap.LDAPResultInvalidCredentials:
			return true
		}
	}
	// By default, return false
	return false
}

func (this *Credentials) getConn() (*ldap.Conn, error) {
	// Close connection
	if this.ctime.IsZero() || time.Now().After(this.ctime) {
		if this.conn != nil {
			this.conn.Close()
			this.conn = nil
		}
	}

	// Open connection
	if this.conn == nil {
		if conn, err := ldap.DialURL(this.addr.String()); err != nil {
			return nil, err
		} else {
			this.conn = conn
			this.ctime = time.Now().Add(LDAP_RECONNECT_DELTA)
		}
	}

	// Return success
	return this.conn, nil
}
