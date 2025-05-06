package ldap

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	// Packages
	ldap "github.com/go-ldap/ldap/v3"
	"github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
	ref "github.com/mutablelogic/go-server/pkg/ref"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// LDAP manager
type Manager struct {
	sync.Mutex
	url        *url.URL
	tls        *tls.Config
	user, pass string
	dn         *schema.DN
	conn       *ldap.Conn
	users      *schema.Group
	groups     *schema.Group
}

var _ server.LDAP = (*Manager)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager(opt ...Opt) (*Manager, error) {
	self := new(Manager)

	// Apply options
	o, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	}

	// Set the url for the connection
	if o.url == nil {
		return nil, httpresponse.ErrBadRequest.With("missing url parameter")
	} else {
		self.url = o.url
	}

	// Check the scheme
	switch self.url.Scheme {
	case schema.MethodPlain:
		if self.url.Port() == "" {
			self.url.Host = fmt.Sprintf("%s:%d", self.url.Hostname(), schema.PortPlain)
		}
	case schema.MethodSecure:
		if self.url.Port() == "" {
			self.url.Host = fmt.Sprintf("%s:%d", self.url.Hostname(), schema.PortSecure)
		}
		self.tls = &tls.Config{
			InsecureSkipVerify: o.skipverify,
		}
	default:
		return nil, fmt.Errorf("scheme not supported: %q", self.url.Scheme)
	}

	// Extract the user
	if o.user != "" {
		self.user = o.user
	} else if self.url.User == nil {
		return nil, httpresponse.ErrBadRequest.With("missing user parameter")
	} else {
		self.user = self.url.User.Username()
	}

	// Extract the password
	if o.pass != "" {
		self.pass = o.pass
	} else if self.url.User != nil {
		if password, ok := self.url.User.Password(); ok {
			self.pass = password
		}
	}

	// Blank out the user and password in the URL
	self.url.User = nil

	// Set the Distinguished Name
	if o.dn == nil {
		return nil, httpresponse.ErrBadRequest.With("missing dn parameter")
	} else {
		self.dn = o.dn
	}

	// Set the schemas for users, groups
	self.users = o.users
	self.groups = o.groups

	// Return success
	return self, nil
}

func (manager *Manager) Run(ctx context.Context) error {
	var retries uint

	// Connect after a short random delay
	ticker := time.NewTimer(time.Millisecond * time.Duration(rand.Intn(100)))
	defer ticker.Stop()

	// Continue to reconnect until cancelled
	for {
		select {
		case <-ctx.Done():
			if err := manager.Disconnect(); err != nil {
				return err
			}
			return nil
		case <-ticker.C:
			if err := manager.Connect(); err != nil {
				// Connection error
				logf(ctx, "LDAP connection error: %v", err)
				retries = min(retries+1, schema.MaxRetries)
				ticker.Reset(schema.MinRetryInterval * time.Duration(retries*retries))
			} else {
				// Connection successful
				if retries > 0 {
					logf(ctx, "LDAP connected")
				}
				retries = 0
				ticker.Reset(schema.MinRetryInterval * time.Duration(schema.MaxRetries))
			}
		}
	}
}

// utility logging function
func logf(ctx context.Context, format string, args ...any) {
	if log := ref.Log(ctx); log != nil {
		log.Printf(ctx, format, args...)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the port for the LDAP connection
func (ldap *Manager) Port() int {
	port, err := strconv.ParseUint(ldap.url.Port(), 10, 32)
	if err != nil {
		return 0
	} else {
		return int(port)
	}
}

// Return the host for the LDAP connection
func (ldap *Manager) Host() string {
	return ldap.url.Hostname()
}

// Return the user for the LDAP connection
func (ldap *Manager) User() string {
	if types.IsIdentifier(ldap.user) {
		// If it's an identifier, then append the DN
		return fmt.Sprint("cn=", ldap.user, ",", ldap.dn)
	} else {
		// Assume it's a DN
		return ldap.user
	}
}

// Connect to the LDAP server, or ping the server if already connected
func (manager *Manager) Connect() error {
	manager.Lock()
	defer manager.Unlock()

	if manager.conn == nil {
		if conn, err := ldapConnect(manager.Host(), manager.Port(), manager.tls); err != nil {
			return err
		} else if err := ldapBind(conn, manager.User(), manager.pass); err != nil {
			return ldaperr(err)
		} else {
			manager.conn = conn
		}
	} else if _, err := manager.conn.WhoAmI([]ldap.Control{}); err != nil {
		// TODO: ldap.ErrorNetwork, ldap.LDAPResultBusy, ldap.LDAPResultUnavailable:
		// would indicate that the connection is no longer valid
		var conn *ldap.Conn
		conn, manager.conn = manager.conn, nil
		return errors.Join(err, ldapDisconnect(conn))
	}

	// Return success
	return nil
}

// Disconnect from the LDAP server
func (ldap *Manager) Disconnect() error {
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
func (manager *Manager) WhoAmI() (string, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return "", httpresponse.ErrGatewayError.With("Not connected")
	}

	// Ping
	if whoami, err := manager.conn.WhoAmI([]ldap.Control{}); err != nil {
		return "", ldaperr(err)
	} else {
		return whoami.AuthzID, nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - OBJECTS

// Return the objects as a list
func (manager *Manager) List(ctx context.Context, request schema.ObjectListRequest) (*schema.ObjectList, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// Set the limit to be the minimum of user and schema limits
	limit := uint64(schema.MaxListEntries)
	if request.Limit != nil {
		limit = min(types.PtrUint64(request.Limit), limit)
	}

	// Set filter
	filter := "(objectclass=*)"
	if request.Filter != nil {
		filter = types.PtrString(request.Filter)
	}

	// Perform the search through paging, skipping the first N entries
	var list schema.ObjectList
	if err := manager.list(ctx, ldap.ScopeWholeSubtree, manager.dn.String(), filter, 0, func(entry *schema.Object) error {
		if list.Count >= request.Offset && list.Count < request.Offset+limit {
			list.Body = append(list.Body, entry)
		}
		list.Count = list.Count + 1
		return nil
	}, request.Attr...); err != nil {
		return nil, err
	}

	// Return success
	return &list, nil
}

// Return the objects as a list using paging, calling a function for each entry.
// When max is zero, paging is used to retrieve all entries. If max is greater than zero,
// then the maximum number of entries is returned.
func (manager *Manager) list(ctx context.Context, scope int, dn, filter string, max uint64, fn func(*schema.Object) error, attrs ...string) error {
	// Create the paging control
	var controls []ldap.Control
	paging := ldap.NewControlPaging(schema.MaxListPaging)
	if max == 0 {
		controls = []ldap.Control{paging}
	}

	// Create the search request
	req := ldap.NewSearchRequest(
		dn,
		scope,
		ldap.NeverDerefAliases,
		int(max), // Size Limit
		0,        // Time Limit
		false,    // Types Only
		filter,   // Filter
		attrs,    // Attributes
		controls, // Controls
	)

	// Perform the search through paging
	for {
		r := manager.conn.SearchAsync(ctx, req, 0)
		for r.Next() {
			entry := r.Entry()
			if entry == nil {
				continue
			}
			if err := fn(schema.NewObjectFromEntry(entry)); errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return err
			}
		}
		if err := r.Err(); err != nil {
			return ldaperr(err)
		}

		// Get response paging control, and copy the cookie over
		if resp, ok := ldap.FindControl(r.Controls(), ldap.ControlTypePaging).(*ldap.ControlPaging); !ok {
			break
		} else if len(resp.Cookie) == 0 {
			break
		} else {
			paging.SetCookie(resp.Cookie)
		}
	}

	// Return success
	return nil
}

// Get an object by DN
func (manager *Manager) Get(ctx context.Context, dn string) (*schema.Object, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// Make absolute DN
	absdn, err := manager.absdn(dn)
	if err != nil {
		return nil, err
	}

	// Get the object
	return manager.get(ctx, ldap.ScopeBaseObject, absdn.String(), "(objectclass=*)")
}

func (manager *Manager) get(ctx context.Context, scope int, dn, filter string, attrs ...string) (*schema.Object, error) {
	var result *schema.Object

	// Search for one object
	if err := manager.list(ctx, scope, dn, filter, 1, func(entry *schema.Object) error {
		result = entry
		return io.EOF
	}, attrs...); errors.Is(err, io.EOF) {
		// Do nothing
	} else if err != nil {
		return nil, err
	}

	// Return success
	return result, nil
}

// Create an object
func (manager *Manager) Create(ctx context.Context, dn string, attr url.Values) (*schema.Object, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// Make absolute DN
	absdn, err := manager.absdn(dn)
	if err != nil {
		return nil, err
	}

	// Create the request
	addReq := ldap.NewAddRequest(absdn.String(), []ldap.Control{})
	for key, values := range attr {
		if len(values) > 0 {
			addReq.Attribute(key, values)
		}
	}

	// Make the request
	if err := manager.conn.Add(addReq); err != nil {
		return nil, ldaperr(err)
	}

	// Return the new object
	return manager.get(ctx, ldap.ScopeBaseObject, addReq.DN, "(objectclass=*)")
}

// Delete an object by DN
func (manager *Manager) Delete(ctx context.Context, dn string) (*schema.Object, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// Make absolute DN
	absdn, err := manager.absdn(dn)
	if err != nil {
		return nil, err
	}

	// Get the object
	object, err := manager.get(ctx, ldap.ScopeBaseObject, absdn.String(), "(objectclass=*)")
	if err != nil {
		return nil, ldaperr(err)
	}

	// Delete the object
	if err := manager.conn.Del(ldap.NewDelRequest(object.DN, []ldap.Control{})); err != nil {
		return nil, ldaperr(err)
	}

	// Return success
	return object, nil
}

// Bind a user to check if they are authenticated, returns
// httpresponse.ErrNotAuthorized if the credentials are invalid
func (manager *Manager) Bind(ctx context.Context, dn, password string) (*schema.Object, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// Make absolute DN
	absdn, err := manager.absdn(dn)
	if err != nil {
		return nil, err
	}

	// Bind - which may result in invalid credentials
	var errs error
	if err := manager.conn.Bind(absdn.String(), password); ldapErrorCode(err) == ldap.LDAPResultInvalidCredentials {
		errs = ldaperr(err)
	} else if err != nil {
		return nil, err
	}

	// Rebind with this user
	if err := ldapBind(manager.conn, manager.User(), manager.pass); err != nil {
		return nil, errors.Join(errs, ldaperr(err))
	} else if errs != nil {
		return nil, errs
	}

	// Return the user
	return manager.get(ctx, ldap.ScopeBaseObject, absdn.String(), "(objectclass=*)")
}

// Change a password for a user. If the new password is empty, then the password is reset
// to a new random password and returned. The old password is required for the change if
// the ldap connection is not bound to the admin user.
func (manager *Manager) ChangePassword(ctx context.Context, dn, old string, new *string) (*schema.Object, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// New password is required
	if new == nil {
		return nil, httpresponse.ErrBadRequest.With("New password parameter is required")
	}

	// Make absolute DN
	absdn, err := manager.absdn(dn)
	if err != nil {
		return nil, err
	}

	// Modify the password
	if result, err := manager.conn.PasswordModify(ldap.NewPasswordModifyRequest(absdn.String(), old, types.PtrString(new))); err != nil {
		return nil, ldaperr(err)
	} else if new != nil {
		*new = result.GeneratedPassword
	}

	// Return the user
	return manager.get(ctx, ldap.ScopeBaseObject, absdn.String(), "(objectclass=*)")
}

// Update attributes for an object. It will replace the attributes where the values is not empty,
// and delete the attributes where the values is empty. The object is returned after the update.
func (manager *Manager) Update(ctx context.Context, dn string, attr url.Values) (*schema.Object, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// Make absolute DN
	absdn, err := manager.absdn(dn)
	if err != nil {
		return nil, err
	}

	// Create the request
	modifyReq := ldap.NewModifyRequest(absdn.String(), []ldap.Control{})
	for key, values := range attr {
		modifyReq.Replace(key, values)
	}

	// Make the request
	if err := manager.conn.Modify(modifyReq); err != nil {
		return nil, ldaperr(err)
	}

	// Return the new object
	return manager.get(ctx, ldap.ScopeBaseObject, absdn.String(), "(objectclass=*)")
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - OBEJCT CLASSES AND ATTRIBUTES

// Returns object classes
func (manager *Manager) ListObjectClasses(ctx context.Context) ([]*schema.ObjectClass, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// Get the subschema dn from rootDSE
	root, err := manager.get(ctx, ldap.ScopeBaseObject, "", "(objectclass=*)", schema.AttrSubSchemaDN)
	if err != nil {
		return nil, err
	}

	// Get the subschema dn attribute
	subschemadn := root.Get(schema.AttrSubSchemaDN)
	if subschemadn == nil {
		return nil, httpresponse.ErrNotFound.With(schema.AttrSubSchemaDN, " not found")
	}

	// List the object classes
	var result []*schema.ObjectClass
	if err := manager.list(ctx, ldap.ScopeBaseObject, types.PtrString(subschemadn), "(objectclass=subschema)", 1, func(entry *schema.Object) error {
		objectClasses := entry.GetAll(schema.AttrObjectClasses)
		if objectClasses == nil {
			return httpresponse.ErrInternalError.With(schema.AttrObjectClasses, " not found")
		}

		// Parse the object classes
		for _, objectClass := range objectClasses {
			if objectClass, err := schema.ParseObjectClass(objectClass); err == nil && objectClass != nil {
				result = append(result, objectClass)
			}
		}
		return nil
	}, schema.AttrObjectClasses); err != nil {
		return nil, err
	}

	// Return success
	return result, nil
}

// Returns attribute types
func (manager *Manager) ListAttributeTypes(ctx context.Context) ([]*schema.AttributeType, error) {
	manager.Lock()
	defer manager.Unlock()

	// Check connection
	if manager.conn == nil {
		return nil, httpresponse.ErrGatewayError.With("Not connected")
	}

	// Get the subschema dn from rootDSE
	root, err := manager.get(ctx, ldap.ScopeBaseObject, "", "(objectclass=*)", schema.AttrSubSchemaDN)
	if err != nil {
		return nil, err
	}

	// Get the subschema dn attribute
	subschemadn := root.Get(schema.AttrSubSchemaDN)
	if subschemadn == nil {
		return nil, httpresponse.ErrNotFound.With(schema.AttrSubSchemaDN, " not found")
	}

	// List the attribute types
	var result []*schema.AttributeType
	if err := manager.list(ctx, ldap.ScopeBaseObject, types.PtrString(subschemadn), "(objectclass=subschema)", 1, func(entry *schema.Object) error {
		attributeTypes := entry.GetAll(schema.AttrAttributeTypes)
		if attributeTypes == nil {
			return httpresponse.ErrInternalError.With(schema.AttrAttributeTypes, " not found")
		}

		// Parse the object classes
		for _, attributeType := range attributeTypes {
			if attributeType, err := schema.ParseAttributeType(attributeType); err == nil && attributeType != nil {
				result = append(result, attributeType)
			}
		}
		return nil
	}, schema.AttrAttributeTypes); err != nil {
		return nil, err
	}

	// Return success
	return result, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - USERS AND GROUPS

// Return all users
func (manager *Manager) ListUsers(ctx context.Context, request schema.ObjectListRequest) ([]*schema.ObjectList, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("ListUsers not implemented")
}

// Return all groups
func (manager *Manager) ListGroups(ctx context.Context, request schema.ObjectListRequest) ([]*schema.ObjectList, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("ListGroups not implemented")
}

// Get a user
func (manager *Manager) GetUser(ctx context.Context, dn string) (*schema.Object, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("GetUser not implemented")
}

// Get a group
func (manager *Manager) GetGroup(ctx context.Context, dn string) (*schema.Object, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("GetGroup not implemented")
}

// Create a user
func (manager *Manager) CreateUser(ctx context.Context, user string, attrs url.Values) (*schema.Object, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("CreateUser not implemented")
}

// Create a group
func (manager *Manager) CreateGroup(ctx context.Context, group string, attrs url.Values) (*schema.Object, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("CreateGroup not implemented")
}

// Delete a user
func (manager *Manager) DeleteUser(ctx context.Context, dn string) (*schema.Object, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("DeleteUser not implemented")
}

// Delete a group
func (manager *Manager) DeleteGroup(ctx context.Context, dn string) (*schema.Object, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("DeleteGroup not implemented")
}

// Add a user to a group, and return the group
func (manager *Manager) AddGroupUser(ctx context.Context, dn, user string) (*schema.Object, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("DeleteGroup not implemented")
}

// Remove a user from a group, and return the group
func (manager *Manager) RemoveGroupUser(ctx context.Context, dn, user string) (*schema.Object, error) {
	// TODO
	return nil, httpresponse.ErrNotImplemented.With("DeleteGroup not implemented")
}

/*
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
*/

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Connect to the LDAP server
func ldapConnect(host string, port int, tls *tls.Config) (*ldap.Conn, error) {
	var url string
	if tls == nil {
		url = fmt.Sprintf("%s://%s:%d", schema.MethodPlain, host, port)
	} else {
		url = fmt.Sprintf("%s://%s:%d", schema.MethodSecure, host, port)
	}
	return ldap.DialURL(url, ldap.DialWithTLSConfig(tls))
}

// Disconnect from the LDAP server
func ldapDisconnect(conn *ldap.Conn) error {
	return conn.Close()
}

// Bind to the LDAP server with a user and password
func ldapBind(conn *ldap.Conn, user, password string) error {
	if password == "" {
		return conn.UnauthenticatedBind(user)
	} else {
		return conn.Bind(user, password)
	}
}

// Return the LDAP error code
func ldapErrorCode(err error) uint16 {
	if err, ok := err.(*ldap.Error); ok {
		return uint16(err.ResultCode)
	} else {
		return 0
	}
}

// Translate LDAP error to HTTP error
func ldaperr(err error) error {
	if err == nil {
		return nil
	}
	code := ldapErrorCode(err)
	if code == 0 {
		return err
	}
	switch code {
	case ldap.LDAPResultInvalidCredentials:
		return httpresponse.ErrNotAuthorized.With("Invalid credentials")
	case ldap.LDAPResultNoSuchObject:
		return httpresponse.ErrNotFound.With("No such object")
	case ldap.LDAPResultEntryAlreadyExists:
		return httpresponse.ErrConflict.With(err.Error())
	case ldap.LDAPResultNoSuchAttribute:
		return httpresponse.ErrNotFound.With(err.Error())
	case ldap.LDAPResultConstraintViolation:
		return httpresponse.ErrConflict.With(err.Error())
	case ldap.LDAPResultUnwillingToPerform:
		return httpresponse.Err(http.StatusServiceUnavailable).With(err.Error())
	default:
		return httpresponse.ErrInternalError.With(err)
	}
}

// Make the DN absolute
func (manager *Manager) absdn(dn string) (*schema.DN, error) {
	rdn, err := schema.NewDN(dn)
	if err != nil {
		return nil, httpresponse.ErrBadRequest.Withf("Invalid DN: %v", err.Error())
	}
	if !manager.dn.AncestorOf(rdn) {
		return rdn.Join(manager.dn), nil
	}
	return rdn, nil
}
