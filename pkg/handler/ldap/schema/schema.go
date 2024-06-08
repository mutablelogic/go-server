package schema

import (
	"fmt"
	"slices"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

/////////////////////////////////////////////////////////////////////
// TYPES

// Schema is the schema for the LDAP server
type Schema struct {
	UserOU           string   `hcl:"user-ou,optional" description:"User Organisational Unit"`
	GroupOU          string   `hcl:"group-ou,optional" description:"Group Organisational Unit"`
	UserObjectClass  []string `hcl:"user-object-class,optional" description:"User object class"`
	GroupObjectClass []string `hcl:"group-object-class,optional" description:"Group object class"`
}

/////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultPosixUserObjectClass = "posixAccount"
)

const (
	// https://learn.microsoft.com/en-us/windows/win32/adschema/a-instancetype
	INSTANCE_TYPE_WRITABLE = 0x00000004
)

const (
	// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/11972272-09ec-4a42-bf5e-3e99b321cf55
	GROUP_TYPE_BUILTIN_LOCAL_GROUP = 0x00000001
	GROUP_TYPE_ACCOUNT_GROUP       = 0x00000002
	GROUP_TYPE_RESOURCE_GROUP      = 0x00000004
	GROUP_TYPE_UNIVERSAL_GROUP     = 0x00000008
	GROUP_TYPE_APP_BASIC_GROUP     = 0x00000010
	GROUP_TYPE_APP_QUERY_GROUP     = 0x00000020
	GROUP_TYPE_SECURITY_ENABLED    = 0x80000000
)

const (
	// https://learn.microsoft.com/en-us/troubleshoot/windows-server/active-directory/useraccountcontrol-manipulate-account-properties
	USER_SCRIPT                         = 0x00000001
	USER_ACCOUNTDISABLE                 = 0x00000002
	USER_HOMEDIR_REQUIRED               = 0x00000008
	USER_LOCKOUT                        = 0x00000010
	USER_PASSWD_NOTREQD                 = 0x00000020
	USER_PASSWD_CANT_CHANGE             = 0x00000040
	USER_ENCRYPTED_TEXT_PWD_ALLOWED     = 0x00000080
	USER_TEMP_DUPLICATE_ACCOUNT         = 0x00000100
	USER_NORMAL_ACCOUNT                 = 0x00000200
	USER_INTERDOMAIN_TRUST_ACCOUNT      = 0x00000800
	USER_WORKSTATION_TRUST_ACCOUNT      = 0x00001000
	USER_SERVER_TRUST_ACCOUNT           = 0x00002000
	USER_DONT_EXPIRE_PASSWORD           = 0x00010000
	USER_MNS_LOGON_ACCOUNT              = 0x00020000
	USER_SMARTCARD_REQUIRED             = 0x00040000
	USER_TRUSTED_FOR_DELEGATION         = 0x00080000
	USER_NOT_DELEGATED                  = 0x00100000
	USER_USE_DES_KEY_ONLY               = 0x00200000
	USER_DONT_REQ_PREAUTH               = 0x00400000
	USER_PASSWORD_EXPIRED               = 0x00800000
	USER_TRUSTED_TO_AUTH_FOR_DELEGATION = 0x01000000
	USER_PARTIAL_SECRETS_ACCOUNT        = 0x04000000
)

/////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Returns a new group object
func (s Schema) NewGroup(dn, name string, attrs ...Attr) (*Object, error) {
	// Check parameters
	if !types.IsIdentifier(name) || s.GroupOU == "" {
		return nil, ErrBadParameter.With("name")
	}

	// Set attributes
	group := NewObject(s.groupDN(dn, name))
	group.Set("objectClass", s.GroupObjectClass...)
	group.Set("cn", name)
	if !s.isPosix() {
		// Active Directory Supported
		group.Set("sAMAccountName", name)
		group.Set("instanceType", fmt.Sprintf("0x%08X", INSTANCE_TYPE_WRITABLE))
		group.Set("groupType", fmt.Sprintf("0x%08X", GROUP_TYPE_RESOURCE_GROUP|GROUP_TYPE_SECURITY_ENABLED))
	}

	// Add additional attributes
	for _, attr := range attrs {
		if err := attr(group); err != nil {
			return nil, err
		}
	}

	// groupOfUniqueNames support
	if slices.Contains(s.GroupObjectClass, "groupOfUniqueNames") && !group.Has("uniqueMember") {
		group.Set("uniqueMember", group.DN)
	}

	// Return success
	return group, nil
}

// Returns a new user object
func (s Schema) NewUser(dn, name string, attrs ...Attr) (*Object, error) {
	// Check parameters
	if !types.IsIdentifier(name) || s.UserOU == "" {
		return nil, ErrBadParameter.With("name")
	}

	// Set attributes
	user := NewObject(s.userDN(dn, name))
	user.Set("objectClass", s.UserObjectClass...)
	user.Set("cn", name)
	if !s.isPosix() {
		// Active Directory Supported
		user.Set("name", name)
		user.Set("sAMAccountName", name)
		user.Set("userAccountControl", fmt.Sprintf("0x%08X", 0))
		user.Set("instanceType", fmt.Sprintf("0x%08X", INSTANCE_TYPE_WRITABLE))
	}

	// Add additional attributes
	for _, attr := range attrs {
		if err := attr(user); err != nil {
			return nil, err
		}
	}

	// set userPrincipalName
	if !s.isPosix() && user.Has("mail") {
		user.Set("userPrincipalName", user.Get("mail"))
	}

	// Return success
	return user, nil
}

/////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns true if the schema is posix (objectClass includes posixAccount)
func (s Schema) isPosix() bool {
	return slices.Contains(s.UserObjectClass, defaultPosixUserObjectClass)
}

// Returns the group DN for a given name
func (s Schema) groupDN(dn, name string) string {
	return fmt.Sprintf("cn=%s,ou=%s,%s", name, s.GroupOU, dn)
}

// Returns the user DN for a given name
func (s Schema) userDN(dn, name string) string {
	if s.isPosix() {
		return fmt.Sprintf("uid=%s,ou=%s,%s", name, s.UserOU, dn)
	} else {
		return fmt.Sprintf("cn=%s,ou=%s,%s", name, s.UserOU, dn)
	}
}
