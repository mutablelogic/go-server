package ldap

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	// Packages
	goldap "github.com/go-ldap/ldap/v3"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// ldap instance
type ldap struct {
	sync.Mutex

	url            *url.URL
	tls            *tls.Config
	user, password string
	dn             string
	conn           *goldap.Conn
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New(c Config) (*ldap, error) {
	self := new(ldap)

	// Set the URL for the connection
	if c.URL == "" {
		return nil, ErrBadParameter.With("url")
	} else if url, err := url.Parse(c.URL); err != nil {
		return nil, err
	} else {
		self.url = url
	}

	// Check the scheme
	switch self.url.Scheme {
	case defaultMethodPlain:
		if self.url.Port() == "" {
			self.url.Host = fmt.Sprintf("%s:%d", self.url.Hostname(), defaultPortPlain)
		}
	case defaultMethodSecure:
		if self.url.Port() == "" {
			self.url.Host = fmt.Sprintf("%s:%d", self.url.Hostname(), defaultPortSecure)
		}
		self.tls = &tls.Config{
			InsecureSkipVerify: c.TLS.SkipVerify,
		}
	default:
		return nil, fmt.Errorf("scheme not supported: %q (expected: %q, %q)", self.url.Scheme, defaultMethodPlain, defaultMethodSecure)
	}

	// Extract the user
	if c.User != "" {
		self.user = c.User
	} else if self.url.User == nil {
		return nil, ErrBadParameter.With("missing user parameter")
	} else {
		self.user = self.url.User.Username()
	}

	// Extract the password
	if c.Password != "" {
		self.password = c.Password
	} else if self.url.User != nil {
		if password, ok := self.url.User.Password(); ok {
			self.password = password
		}
	}

	// Blank out the user and password in the URL
	self.url.User = nil

	// Set the Distinguished Name
	if c.DN == "" {
		return nil, ErrBadParameter.With("dn")
	} else {
		self.dn = c.DN
	}

	// Return success
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the port for the LDAP connection
func (ldap *ldap) Port() int {
	port, err := strconv.ParseUint(ldap.url.Port(), 10, 32)
	if err != nil {
		return 0
	} else {
		return int(port)
	}
}

// Return the host for the LDAP connection
func (ldap *ldap) Host() string {
	return ldap.url.Hostname()
}

// Return the user for the LDAP connection
func (ldap *ldap) User() string {
	if types.IsIdentifier(ldap.user) {
		return fmt.Sprint("cn=", ldap.user, ",", ldap.dn)
	} else {
		return ldap.user
	}
}

// Connect to the LDAP server, or ping the server if already connected
func (ldap *ldap) Connect() error {
	ldap.Lock()
	defer ldap.Unlock()

	if ldap.conn == nil {
		if conn, err := ldapConnect(ldap.Host(), ldap.Port(), ldap.tls); err != nil {
			return err
		} else if err := ldapBind(conn, ldap.User(), ldap.password); err != nil {
			if ldapErrorCode(err) == goldap.LDAPResultInvalidCredentials {
				return ErrNotAuthorized.With("Invalid credentials")
			} else {
				return err
			}
		} else {
			ldap.conn = conn
		}
	} else if _, err := ldap.conn.WhoAmI([]goldap.Control{}); err != nil {
		return errors.Join(err, ldap.Disconnect())
	}

	// Return success
	return nil
}

// Disconnect from the LDAP server
func (ldap *ldap) Disconnect() error {
	ldap.Lock()
	defer ldap.Unlock()

	// Disconnect from LDAP connection
	var result error
	if ldap.conn != nil {
		if err := ldapDisconnect(ldap.conn); err != nil {
			result = errors.Join(result, err)
		}
		ldap.conn = nil
	}

	// Return any errors
	return result
}

// Return the user who is currently authenticated
func (ldap *ldap) WhoAmI() (string, error) {
	ldap.Lock()
	defer ldap.Unlock()

	// Check connection
	if ldap.conn == nil {
		return "", ErrOutOfOrder.With("Not connected")
	}

	// Ping
	if whoami, err := ldap.conn.WhoAmI([]goldap.Control{}); err != nil {
		return "", err
	} else {
		return whoami.AuthzID, nil
	}
}

// Return the objects of a particular class, or use "*" to return all objects
func (ldap *ldap) Get(objectClass string) ([]*object, error) {
	ldap.Lock()
	defer ldap.Unlock()

	// Check parameters
	if !types.IsIdentifier(objectClass) && objectClass != "*" {
		return nil, ErrBadParameter.With("objectClass")
	}

	// Check connection
	if ldap.conn == nil {
		return nil, ErrOutOfOrder.With("Not connected")
	}

	// Define the search request
	searchRequest := goldap.NewSearchRequest(
		ldap.dn, goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprint("(&(objectClass=", objectClass, "))"), // The filter to apply
		nil, // The attributes to retrieve
		nil,
	)

	// Perform the search
	sr, err := ldap.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	// Print the results
	result := make([]*object, 0, len(sr.Entries))
	for _, entry := range sr.Entries {
		result = append(result, newObject(entry))
	}

	// Return success
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func ldapConnect(host string, port int, tls *tls.Config) (*goldap.Conn, error) {
	var url string
	if tls != nil {
		url = fmt.Sprintf("%s://%s:%d", defaultMethodSecure, host, port)
	} else {
		url = fmt.Sprintf("%s://%s:%d", defaultMethodPlain, host, port)
	}
	return goldap.DialURL(url, goldap.DialWithTLSConfig(tls))
}

func ldapDisconnect(conn *goldap.Conn) error {
	return conn.Close()
}

func ldapBind(conn *goldap.Conn, user, password string) error {
	if user == "" || password == "" {
		return conn.UnauthenticatedBind(user)
	} else {
		return conn.Bind(user, password)
	}
}

// Return the LDAP error code
func ldapErrorCode(err error) uint16 {
	if err, ok := err.(*goldap.Error); ok {
		return uint16(err.ResultCode)
	} else {
		return 0
	}
}
