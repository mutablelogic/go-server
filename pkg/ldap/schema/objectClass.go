package schema

import (
	"encoding/json"

	// Packages
	parser "github.com/yinyin/go-ldap-schema-parser"
)

//////////////////////////////////////////////////////////////////////////////////
// TYPES

type ObjectClass struct {
	*parser.ObjectClassSchema
}

//////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func ParseObjectClass(v string) (*ObjectClass, error) {
	schema, err := parser.ParseObjectClassSchema(v)
	if err != nil {
		return nil, err
	}
	return &ObjectClass{schema}, nil
}

//////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o *ObjectClass) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
