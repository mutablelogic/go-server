package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/scanner"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

type ACLItem struct {
	Role    string   `json:"role,omitempty" help:"Role name"`
	Priv    []string `json:"priv,omitempty" help:"Access privileges"`
	Grantor string   `json:"-" help:"Grantor"` // Ignore field for now
}

type ACLList []*ACLItem

/////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	privSelect      = "SELECT"
	privInsert      = "INSERT"
	privUpdate      = "UPDATE"
	privDelete      = "DELETE"
	privTruncate    = "TRUNCATE"
	privReferences  = "REFERENCES"
	privTrigger     = "TRIGGER"
	privCreate      = "CREATE"
	privConnect     = "CONNECT"
	privTemporary   = "TEMPORARY"
	privExecute     = "EXECUTE"
	privUsage       = "USAGE"
	privSet         = "SET"
	privAlterSystem = "ALTER SYSTEM"
	privMaintain    = "MAINTAIN"
	privWithGrant   = "WITH GRANT OPTION"
	privAll         = "ALL"
)

var (
	// Map of privilege names to their string values
	// https://www.postgresql.org/docs/current/ddl-priv.html
	privs = map[rune]string{
		'r': privSelect,      // Large Object, Sequence, Table, Column
		'a': privInsert,      // Table or Column
		'w': privUpdate,      // Large Object, Sequence, Table, Column
		'd': privDelete,      // Table
		'D': privTruncate,    // Table
		'x': privReferences,  // Table or Column
		't': privTrigger,     // Table
		'C': privCreate,      // Database, Schema, Tablespace
		'c': privConnect,     // Database
		'T': privTemporary,   // Database
		'X': privExecute,     // Function, Procedure
		'U': privUsage,       // Domain, Foreigh data wrapper, Foreign server, Language, Schema, Sequence, Type
		's': privSet,         // Parameter
		'A': privAlterSystem, // Parameter
		'm': privMaintain,    // Table
		'*': privWithGrant,   // Grant
	}
	privsIndex = make(map[string]rune, len(privs))
)

func init() {
	for k, v := range privs {
		if v != privWithGrant {
			privsIndex[v] = k
		}
	}
	privsIndex[privAll] = '$'
}

var (
	reRoleName = `([^\n=]*)`
	rePriv     = `([^\n/]*)`
	reAclItem  = regexp.MustCompile(`^` + reRoleName + `=` + rePriv + `/` + reRoleName + `$`)
)

/////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new ACLItem from a postgresql ACL string
func NewACLItem(v string) (*ACLItem, error) {
	tuples := reAclItem.FindStringSubmatch(v)
	if len(tuples) != 4 {
		return nil, httpresponse.ErrBadRequest.Withf("invalid ACL item: %q", v)
	}
	return &ACLItem{
		Role:    toRole(tuples[1]),
		Priv:    toPriv(tuples[2]),
		Grantor: toRole(tuples[3]),
	}, nil
}

// Parse an ACLItem from a command-line flag, like
// <role>:<priv>,<priv>,<priv>...
func ParseACLItem(v string) (*ACLItem, error) {
	item := new(ACLItem)
	if err := item.UnmarshalText([]byte(v)); err != nil {
		return nil, err
	} else {
		return item, nil
	}
}

func (a ACLItem) WithPriv(priv ...string) *ACLItem {
	return &ACLItem{
		Role:    a.Role,
		Priv:    priv,
		Grantor: a.Grantor,
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Grant access privileges to a database
func (acl ACLItem) GrantDatabase(ctx context.Context, conn pg.Conn, name string) error {
	return acl.exec(ctx, conn.With("type", "DATABASE", "name", name, "granted_by", ""), acl.Role, aclGrant)
}

// Revoke access privileges to a database
func (acl ACLItem) RevokeDatabase(ctx context.Context, conn pg.Conn, name string) error {
	return acl.exec(ctx, conn.With("type", "DATABASE", "name", name, "granted_by", ""), acl.Role, aclRevoke)
}

// Grant access privileges to a schema
func (acl ACLItem) GrantSchema(ctx context.Context, conn pg.Conn, name string) error {
	return acl.exec(ctx, conn.With("type", "SCHEMA", "name", name, "granted_by", ""), acl.Role, aclGrant)
}

// Revoke access privileges to a schema
func (acl ACLItem) RevokeSchema(ctx context.Context, conn pg.Conn, name string) error {
	return acl.exec(ctx, conn.With("type", "SCHEMA", "name", name, "granted_by", ""), acl.Role, aclRevoke)
}

// Grant access privileges to a tablespace
func (acl ACLItem) GrantTablespace(ctx context.Context, conn pg.Conn, name string) error {
	return acl.exec(ctx, conn.With("type", "TABLESPACE", "name", name, "granted_by", ""), acl.Role, aclGrant)
}

// Revoke access privileges to a tablespace
func (acl ACLItem) RevokeTablespace(ctx context.Context, conn pg.Conn, name string) error {
	return acl.exec(ctx, conn.With("type", "TABLESPACE", "name", name, "granted_by", ""), acl.Role, aclRevoke)
}

func (acl ACLItem) exec(ctx context.Context, conn pg.Conn, role, sql string) error {
	// PUBLIC -> PUBLIC and role -> "role"
	if role == defaultAclRole {
		conn = conn.With("role", role)
	} else {
		conn = conn.With("role", types.DoubleQuote(role))
	}
	// Set the privileges
	for _, v := range acl.Priv {
		if err := conn.With("priv", v).Exec(ctx, sql); err != nil {
			return err
		}
	}
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// MARSHAL/UNMARSHAL

const (
	stInit = iota
	stSep
	stPriv
	stComma
)

func (acl *ACLItem) UnmarshalText(data []byte) error {
	var tokenizer scanner.Scanner
	state := stInit
	if acl == nil {
		return httpresponse.ErrBadRequest.With("nil ACLItem")
	}
	tokenizer.Init(bytes.NewReader(data))
	for tok := tokenizer.Scan(); tok != scanner.EOF; tok = tokenizer.Scan() {
		switch {
		case state == stInit && tok == scanner.String:
			if role, err := strconv.Unquote(tokenizer.TokenText()); err != nil {
				return httpresponse.ErrBadRequest.Withf("invalid role %q", tokenizer.TokenText())
			} else {
				acl.Role = role
			}
			state = stSep
		case state == stInit && tok == scanner.Ident:
			acl.Role = tokenizer.TokenText()
			state = stSep
		case state == stSep && tok == ':':
			state = stPriv
		case state == stPriv && tok == scanner.Ident:
			priv := strings.ToUpper(tokenizer.TokenText())
			if _, exists := privsIndex[priv]; !exists {
				return httpresponse.ErrBadRequest.Withf("invalid privilege %q", priv)
			} else if !slices.Contains(acl.Priv, priv) {
				acl.Priv = append(acl.Priv, priv)
			}
			state = stComma
		case state == stComma && tok == ',':
			state = stPriv
		default:
			return httpresponse.ErrBadRequest.Withf("parse error at %q", tokenizer.TokenText())
		}
	}
	if state == stInit || state == stPriv {
		return httpresponse.ErrBadRequest.Withf("parse error at %q", tokenizer.TokenText())
	}
	return nil
}

func (acl *ACLList) UnmarshalText(data []byte) error {
	if acl == nil {
		return httpresponse.ErrBadRequest.With("nil ACLList")
	}
	for _, v := range bytes.Fields(data) {
		item := new(ACLItem)
		if err := item.UnmarshalText(v); err != nil {
			return err
		}
		acl.Append(item)
	}
	return nil
}

func (acl *ACLList) UnmarshalJSON(data []byte) error {
	items := make([]*ACLItem, 0)
	if acl == nil {
		return httpresponse.ErrBadRequest.With("nil ACLList")
	}
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	for _, item := range items {
		acl.Append(item)
	}
	return nil
}

func (acl *ACLItem) UnmarshalJSON(data []byte) error {
	var fields map[string]any
	if acl == nil {
		return httpresponse.ErrBadRequest.With("nil ACLItem")
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	// Set role
	if role, ok := fields["role"].(string); !ok {
		return httpresponse.ErrBadRequest.With("missing role")
	} else if role = strings.TrimSpace(role); role == "" {
		return httpresponse.ErrBadRequest.With("missing role")
	} else if strings.ToUpper(role) == defaultAclRole {
		acl.Role = defaultAclRole
	} else {
		acl.Role = role
	}

	// Set priv
	priv, ok := fields["priv"].([]any)
	if !ok {
		return httpresponse.ErrBadRequest.With("missing priv")
	}
	for _, v := range priv {
		if s, ok := v.(string); !ok {
			return httpresponse.ErrBadRequest.Withf("invalid privilege %q", v)
		} else if s = strings.ToUpper(strings.TrimSpace(s)); s == "" {
			return httpresponse.ErrBadRequest.Withf("missing privilege")
		} else if _, exists := privsIndex[s]; !exists {
			return httpresponse.ErrBadRequest.Withf("invalid privilege %q", s)
		} else {
			acl.Priv = append(acl.Priv, s)
		}
	}

	// Return success
	return nil
}

/////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (a ACLItem) String() string {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

/////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Append an ACLItem to the list
func (acl *ACLList) Append(item *ACLItem) {
	if acl == nil {
		panic("nil ACLList")
	}

	if acl.Find(item.Role) != nil {
		fmt.Println("TODO: Merge two ACLItems for", item.Role)
	}

	*acl = append(*acl, item)
}

// Find an ACLItem in the list by role
func (acl *ACLList) Find(role string) *ACLItem {
	for _, item := range *acl {
		if item.Role == role {
			return item
		}
	}
	return nil
}

// Determine if has ALL privileges
func (acl ACLItem) IsAll() bool {
	return slices.Contains(acl.Priv, privAll)
}

/////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toRole(v string) string {
	if v == "" {
		return defaultAclRole
	}
	return v
}

func toPriv(v string) []string {
	priv := make([]string, 0, len(v))
	for i, r := range v {
		if p, ok := privs[r]; ok {
			if p == privWithGrant && i > 0 {
				priv[len(priv)-1] += " " + p
			} else {
				priv = append(priv, p)
			}
		}
	}

	// Return
	return priv
}

/////////////////////////////////////////////////////////////////////////////
// SQL

const (
	aclGrant  = `GRANT ${priv} ON ${type} ${"name"} TO ${role} ${granted_by}`
	aclRevoke = `REVOKE ${priv} ON ${type} ${"name"} FROM ${role} ${granted_by} CASCADE`
)
