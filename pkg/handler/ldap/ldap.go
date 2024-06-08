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
	schema "github.com/mutablelogic/go-server/pkg/handler/ldap/schema"
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
	schema         *schema.Schema
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

	// Set the schema
	self.schema = c.ObjectSchema()

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
			if code := ldapErrorCode(err); code == goldap.LDAPResultInvalidCredentials || code == goldap.LDAPResultUnwillingToPerform {
				return ErrNotAuthorized.With("Invalid credentials")
			} else {
				return err
			}
		} else {
			ldap.conn = conn
		}
	} else if _, err := ldap.conn.WhoAmI([]goldap.Control{}); err != nil {
		// TODO: ldap.ErrorNetwork, ldap.LDAPResultBusy, ldap.LDAPResultUnavailable:
		// would indicate that the connection is no longer valid
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
func (ldap *ldap) Get(objectClass string) ([]*schema.Object, error) {
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
	result := make([]*schema.Object, 0, len(sr.Entries))
	for _, entry := range sr.Entries {
		result = append(result, schema.NewObjectFromEntry(entry))
	}

	// Return success
	return result, nil
}

// Return all users
func (ldap *ldap) GetUsers() ([]*schema.Object, error) {
	return ldap.Get(ldap.schema.UserObjectClass[0])
}

// Return all groups
func (ldap *ldap) GetGroups() ([]*schema.Object, error) {
	return ldap.Get(ldap.schema.GroupObjectClass[0])
}

// Create a group with the given attributes
func (ldap *ldap) CreateGroup(group string, attrs ...schema.Attr) (*schema.Object, error) {
	ldap.Lock()
	defer ldap.Unlock()

	// Check connection
	if ldap.conn == nil {
		return nil, ErrOutOfOrder.With("Not connected")
	}

	// Create group
	o, err := ldap.schema.NewGroup(ldap.dn, group, attrs...)
	if err != nil {
		return nil, err
	}

	// If the gid is not set, then set it to the next available gid
	var nextGid int
	gid, err := ldap.SearchOne("(&(objectclass=device)(cn=lastgid))")
	if err != nil {
		return nil, err
	} else if gid == nil {
		return nil, ErrNotImplemented.With("lastgid not found")
	} else if gid_, err := strconv.ParseInt(gid.Get("serialNumber"), 10, 32); err != nil {
		return nil, ErrNotImplemented.With("lastgid not found")
	} else {
		nextGid = int(gid_) + 1
		if err := schema.OptGroupId(int(gid_))(o); err != nil {
			return nil, err
		}
	}

	// Create the request
	addReq := goldap.NewAddRequest(o.DN, []goldap.Control{})
	for name, values := range o.Values {
		addReq.Attribute(name, values)
	}

	// Request -> Response
	if err := ldap.conn.Add(addReq); err != nil {
		return nil, err
	}

	// Increment the gid
	if gid != nil && nextGid > 0 {
		modify := goldap.NewModifyRequest(gid.DN, []goldap.Control{})
		modify.Replace("serialNumber", []string{fmt.Sprint(nextGid)})
		if err := ldap.conn.Modify(modify); err != nil {
			return nil, err
		}
	}

	// TODO: Retrieve the group

	// Return success
	return o, nil
}

// Add a user to a group
func (ldap *ldap) AddGroupUser(user, group *schema.Object) error {
	// Use uniqueMember for groupOfUniqueNames,
	// use memberUid for posixGroup
	// use member for groupOfNames or if not posix
	return ErrNotImplemented
}

// Remove a user from a group
func (ldap *ldap) RemoveGroupUser(user, group *schema.Object) error {
	// Use uniqueMember for groupOfUniqueNames,
	// use memberUid for posixGroup
	// use member for groupOfNames or if not posix
	return ErrNotImplemented
}

// Change a passsord for a user. If the new password is empty, then the password is reset
// to a new random password. The old password is required for the change
// if the ldap connection is not bound to the admin user. The new password
// is returned if the change is successful
func (ldap *ldap) ChangePassword(o *schema.Object, old, new string) (string, error) {
	ldap.Lock()
	defer ldap.Unlock()

	// Check object
	if o == nil {
		return "", ErrBadParameter
	}

	// Check connection
	if ldap.conn == nil {
		return "", ErrOutOfOrder.With("Not connected")
	}

	// Modify the password
	modify := goldap.NewPasswordModifyRequest(o.DN, old, new)
	if result, err := ldap.conn.PasswordModify(modify); err != nil {
		return "", err
	} else {
		return result.GeneratedPassword, nil
	}
}

// Create a user in a specific group with the given attributes
func (ldap *ldap) CreateUser(name string, attrs ...schema.Attr) (*schema.Object, error) {
	ldap.Lock()
	defer ldap.Unlock()

	// Check connection
	if ldap.conn == nil {
		return nil, ErrOutOfOrder.With("Not connected")
	}

	// Create user object
	o, err := ldap.schema.NewUser(ldap.dn, name, attrs...)
	if err != nil {
		return nil, err
	}

	// If the uid is not set, then set it to the next available uid
	var nextId int
	uid, err := ldap.SearchOne("(&(objectclass=device)(cn=lastuid))")
	if err != nil {
		return nil, err
	} else if uid == nil {
		return nil, ErrNotImplemented.With("lastuid not found")
	} else if uid_, err := strconv.ParseInt(uid.Get("serialNumber"), 10, 32); err != nil {
		return nil, ErrNotImplemented.With("lastuid not found")
	} else {
		nextId = int(uid_) + 1
		if err := schema.OptUserId(int(uid_))(o); err != nil {
			return nil, err
		}
	}

	// Create the request
	addReq := goldap.NewAddRequest(o.DN, []goldap.Control{})
	for name, values := range o.Values {
		addReq.Attribute(name, values)
	}

	// Request -> Response
	if err := ldap.conn.Add(addReq); err != nil {
		return nil, err
	}

	// Increment the uid
	if uid != nil && nextId > 0 {
		modify := goldap.NewModifyRequest(uid.DN, []goldap.Control{})
		modify.Replace("serialNumber", []string{fmt.Sprint(nextId)})
		if err := ldap.conn.Modify(modify); err != nil {
			return nil, err
		}
	}

	// TODO: Add the user to a group

	// Return success
	return o, nil
}

// Delete an object
func (ldap *ldap) Delete(o *schema.Object) error {
	ldap.Lock()
	defer ldap.Unlock()

	// Check object
	if o == nil {
		return ErrBadParameter
	}

	// Check connection
	if ldap.conn == nil {
		return ErrOutOfOrder.With("Not connected")
	}

	// Delete the object
	return ldap.conn.Del(goldap.NewDelRequest(o.DN, []goldap.Control{}))
}

// Bind a user with password to check if they are authenticated
func (ldap *ldap) Bind(user *schema.Object, password string) error {
	ldap.Lock()
	defer ldap.Unlock()

	// Check connection
	if ldap.conn == nil {
		return ErrOutOfOrder.With("Not connected")
	}

	// Bind
	if err := ldap.conn.Bind(user.DN, password); err != nil {
		if ldapErrorCode(err) == goldap.LDAPResultInvalidCredentials {
			return ErrNotAuthorized.With("Invalid credentials")
		} else {
			return err
		}
	}

	// TODO: Rebind with the original user

	// Return success
	return nil
}

// Return one record
func (ldap *ldap) SearchOne(filter string) (*schema.Object, error) {
	// Check connection
	if ldap.conn == nil {
		return nil, ErrOutOfOrder.With("Not connected")
	}

	// Define the search request
	searchRequest := goldap.NewSearchRequest(
		ldap.dn, goldap.ScopeWholeSubtree, goldap.NeverDerefAliases, 1, 0, false,
		filter, // The filter to apply
		nil,    // The attributes to retrieve
		nil,
	)

	// Perform the search, return first result
	sr, err := ldap.conn.Search(searchRequest)
	if err != nil {
		return nil, err
	} else if len(sr.Entries) == 0 {
		return nil, nil
	} else {
		return schema.NewObjectFromEntry(sr.Entries[0]), nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Connect to the LDAP server
func ldapConnect(host string, port int, tls *tls.Config) (*goldap.Conn, error) {
	var url string
	if tls != nil {
		url = fmt.Sprintf("%s://%s:%d", defaultMethodSecure, host, port)
	} else {
		url = fmt.Sprintf("%s://%s:%d", defaultMethodPlain, host, port)
	}
	return goldap.DialURL(url, goldap.DialWithTLSConfig(tls))
}

// Disconnect from the LDAP server
func ldapDisconnect(conn *goldap.Conn) error {
	return conn.Close()
}

// Bind to the LDAP server with a user and password
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
