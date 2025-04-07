package schema

import (
	"encoding/json"
	"regexp"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

/////////////////////////////////////////////////////////////////////////////
// TYPES

type ACLItem struct {
	Role    string   `json:"role,omitempty" help:"Role name"`
	Priv    []string `json:"priv,omitempty" help:"Access privileges"`
	Grantor string   `json:"-" help:"Grantor"` // Ignore field for now
}

/////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultAclRole = "public"
)

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
)

var (
	// Map of privilege names to their string values
	// https://www.postgresql.org/docs/current/ddl-priv.html
	privs = map[rune]string{
		'r': privSelect,      // Large Object, Sequenct, Table, Column
		'a': privInsert,      // Table or Column
		'w': privUpdate,      // Large Object, Sequenct, Table, Column
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
)

var (
	reRoleName = `([^\n=]*)`
	rePriv     = `([^\n/]*)`
	reAclItem  = regexp.MustCompile(`^` + reRoleName + `=` + rePriv + `/` + reRoleName + `$`)
)

/////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

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
