package provider

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	// Packages
	"github.com/djthorpe/go-tablewriter/pkg/meta"
	"github.com/mutablelogic/go-server"
	"github.com/mutablelogic/go-server/pkg/types"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type PluginMeta struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Fields      []metafield  `json:"fields"`
	Type        reflect.Type `json:"-"`

	// Private fields
	fields map[string]*metafield
}

type metafield struct {
	// Field name
	Key string

	// Description
	Description string

	// Field type
	Type reflect.Type

	// Element type for maps, arrays and slices
	Elem reflect.Type

	// Field index
	Index []int

	// Field children, if any
	Children []metafield
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Field names from hcl or json tags
	tagNames       = "hcl,json"
	tagDescription = "description"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewPluginMeta returns metadata for a plugin
func NewPluginMeta(v server.Plugin) (*PluginMeta, error) {
	meta := &PluginMeta{
		Name:        v.Name(),
		Description: v.Description(),
		Type:        typeOf(v),
		fields:      make(map[string]*metafield),
	}

	// Get fields
	if fields, err := enumerate(nil, reflect.TypeOf(v)); err != nil {
		return nil, err
	} else {
		meta.Fields = fields
	}

	// Index the fields to ensure no duplicates
	if err := index("", meta.Fields, meta.fields); err != nil {
		return nil, err
	}

	// Return success
	return meta, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *PluginMeta) String() string {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (m *metafield) MarshalJSON() ([]byte, error) {
	type j struct {
		Key         string      `json:"key"`
		Description string      `json:"description,omitempty"`
		Type        string      `json:"type"`
		Elem        string      `json:"elem,omitempty"`
		Index       []int       `json:"index"`
		Children    []metafield `json:"children,omitempty"`
	}
	typeToString := func(t reflect.Type) string {
		if t == nil {
			return ""
		}
		return t.String()
	}
	return json.Marshal(j{
		Key:         m.Key,
		Description: m.Description,
		Type:        typeToString(m.Type),
		Elem:        typeToString(m.Elem),
		Index:       m.Index,
		Children:    m.Children,
	})
}

func (m *metafield) String() string {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *PluginMeta) Get(v server.Plugin, label string) (any, error) {
	if typeOf(v) != m.Type {
		return nil, ErrBadParameter.Withf("Expected %q, got %q", m.Type, typeOf(v))
	}

	// Dereference the target
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Simple case
	if field, exists := m.fields[label]; exists {
		return rv.FieldByIndex(field.Index).Interface(), nil
	}

	// Do complex type array, slice, map
	return nil, ErrNotImplemented
}

func (m *PluginMeta) Set(v server.Plugin, label string, value any) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return ErrBadParameter.Withf("Not addressable: %q", v.Name())
	}
	if typeOf(v) != m.Type {
		return ErrBadParameter.Withf("Expected %q, got %q", m.Type, typeOf(v))
	}

	// Dereference the target
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	// Simple case
	if field, exists := m.fields[label]; exists {
		return set(label, rv.FieldByIndex(field.Index), value)
	}

	// Do complex type array, slice, map
	return ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Determine the type of a value
func typeOf(v any) reflect.Type {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// Set a field value
func set(label string, dest reflect.Value, src any) error {
	if !dest.CanSet() {
		return ErrBadParameter.Withf("Cannot set %q", label)
	}
	if src == nil { // If value is nil then set to zero value
		dest.Set(reflect.Zero(dest.Type()))
		return nil
	}
	// Check type of source and destination
	if dest.Type() != reflect.TypeOf(src) {
		return ErrBadParameter.Withf("Cannot set %q, wrong type: %q (expected %q)", label, reflect.TypeOf(src), dest.Type())
	}

	// if the destination is a pointer, then create a new value
	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		dest = dest.Elem()
	}
	// Set the value
	dest.Set(reflect.ValueOf(src))
	return nil
}

func enumerate(index []int, rt reflect.Type) ([]metafield, error) {
	var result []metafield

	// Get metadata for a struct
	meta, err := meta.NewType(rt, strings.Split(tagNames, ",")...)
	if err != nil {
		return nil, err
	}
	for _, field := range meta.Fields() {
		fields, err := enumerate_field(index, field)
		if err != nil {
			return nil, err
		}
		result = append(result, fields)
	}

	// Return success
	return result, nil
}

func enumerate_field(index []int, field meta.Field) (metafield, error) {
	result := metafield{
		Key:         field.Name(),
		Description: field.Tag(tagDescription),
		Type:        field.Type(),
		Index:       append(index, field.Index()...),
	}

	switch result.Type.Kind() {
	case reflect.Struct:
		if fields, err := enumerate(result.Index, field.Type()); err != nil {
			return result, err
		} else {
			result.Children = fields
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		result.Elem = result.Type.Elem()
		// TODO
	}

	// Return appended fields
	return result, nil
}

func index(prefix string, fields []metafield, labels map[string]*metafield) error {
	var result error

	// Index the fields
	for _, field := range fields {
		label := prefix + types.LabelSeparator + field.Key
		if prefix == "" {
			label = field.Key
		}
		if !types.IsIdentifier(field.Key) {
			result = errors.Join(result, ErrBadParameter.Withf("%q", label))
			continue
		}
		if _, exists := labels[label]; exists {
			result = errors.Join(result, ErrDuplicateEntry.Withf("%q", label))
			continue
		} else {
			labels[label] = &field
		}
		if len(field.Children) > 0 {
			if err := index(label, field.Children, labels); err != nil {
				result = errors.Join(result, err)
			}
		}
	}

	// Return any errors
	return result
}
