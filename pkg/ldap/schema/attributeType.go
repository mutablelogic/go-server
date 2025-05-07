package schema

import (
	"encoding/json"

	// Packages
	parser "github.com/yinyin/go-ldap-schema-parser"
)

//////////////////////////////////////////////////////////////////////////////////
// TYPES

type AttributeType struct {
	*parser.AttributeTypeSchema
}

//////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func ParseAttributeType(v string) (*AttributeType, error) {
	schema, err := parser.ParseAttributeTypeSchema(v)
	if err != nil {
		return nil, err
	}
	return &AttributeType{schema}, nil
}

//////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o *AttributeType) String() string {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
