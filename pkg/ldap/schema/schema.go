package schema

import (
	"maps"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
)

const (
	SchemaName   = "ldap"
	APIPrefix    = "/ldap/v1"
	MethodPlain  = "ldap"
	MethodSecure = "ldaps"
	PortPlain    = 389
	PortSecure   = 636

	// Time between connection retries
	MinRetryInterval = time.Second * 5
	MaxRetries       = 10

	// Maximum number of entries to return in a single request
	MaxListPaging = 500

	// Maximum list entries to return
	MaxListEntries = 1000

	// Attributes
	AttrObjectClasses  = "objectClasses"
	AttrObjectClass    = "objectClass"
	AttrAttributeTypes = "attributeTypes"
	AttrSubSchemaDN    = "subschemaSubentry"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ObjectType struct {
	// DN for the type
	DN *DN `json:"dn"`

	// Field name
	Field string `json:"field"`

	// Classes for the type
	ObjectClass []string `json:"objectclass"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewObjectType(dn, field string, classes ...string) (*ObjectType, error) {
	rdn, err := NewDN(dn)
	if err != nil {
		return nil, err
	}
	if !types.IsIdentifier(field) {
		return nil, httpresponse.ErrBadRequest.With("Field is not a valid identifier")
	}
	if len(classes) == 0 {
		return nil, httpresponse.ErrBadRequest.With("ObjectClass is empty")
	}
	return &ObjectType{
		DN:          rdn,
		Field:       field,
		ObjectClass: classes,
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (o *ObjectType) New(value string, attrs url.Values) (*Object, error) {
	object := new(Object)
	rdn, err := NewDN(o.Field + "=" + value)
	if err != nil {
		return nil, err
	} else {
		object.DN = rdn.Join(o.DN).String()
	}

	// Make a copy of the attributes
	object.Values = make(url.Values, len(attrs)+1)
	maps.Copy(object.Values, attrs)

	// Append the object classes
	objectClasses := object.GetAll(AttrObjectClass)
	for _, class := range o.ObjectClass {
		// Check case-insensitive
		if !slices.ContainsFunc(objectClasses, func(s string) bool {
			return strings.EqualFold(s, class)
		}) {
			objectClasses = append(objectClasses, class)
		}
	}
	object.Set(AttrObjectClass, objectClasses...)

	// Return success
	return object, nil
}
