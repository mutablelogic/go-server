package provider

import (
	"reflect"
	"strings"

	// Packages
	"github.com/djthorpe/go-tablewriter/pkg/meta"
	"github.com/mutablelogic/go-server"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type PluginMeta struct {
	Name        string
	Description string
	Fields      map[string]metafield
}

type metafield struct {
	Key   string
	Type  reflect.Type
	Index []int
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tagNames = "hcl,json"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewMeta returns metadata for a plugin
func NewMeta(v server.Plugin) (*PluginMeta, error) {
	meta := &PluginMeta{
		Name:        v.Name(),
		Description: v.Description(),
		Fields:      make(map[string]metafield),
	}

	// Get fields
	fields, err := enumerate_struct("", nil, reflect.TypeOf(v))
	if err != nil {
		return nil, err
	}

	// Field names must be unique
	for _, field := range fields {
		if _, exists := meta.Fields[field.Key]; exists {
			return nil, ErrDuplicateEntry.With(field.Key)
		} else {
			meta.Fields[field.Key] = field
		}
	}

	// Return success
	return meta, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *PluginMeta) Get(v server.Plugin, key string) any {
	// TODO
	// Need some special stuff for arrays and maps
	return nil
}

func (m *PluginMeta) Set(v server.Plugin, key string, value any) {
	// TODO
	// Need some special stuff for arrays and maps
	// Need to update dependencies between plugins
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func enumerate_struct(prefix string, index []int, rt reflect.Type) ([]metafield, error) {
	var result []metafield

	// Get metadata for a plugin
	meta, err := meta.NewType(rt, strings.Split(tagNames, ",")...)
	if err != nil {
		return nil, err
	}
	for _, field := range meta.Fields() {
		fields, err := enumerate_field(prefix, index, field)
		if err != nil {
			return nil, err
		}
		result = append(result, fields...)
	}

	// Return success
	return result, nil
}

func enumerate_field(prefix string, index []int, field meta.Field) ([]metafield, error) {
	key := field.Name()
	if prefix != "" {
		key = prefix + "." + key
	}
	return enumerate_type(key, append(index, field.Index()...), field.Type())
}

func enumerate_type(key string, index []int, rt reflect.Type) ([]metafield, error) {
	var result []metafield

	switch rt.Kind() {
	case reflect.Struct:
		if fields, err := enumerate_struct(key, index, rt); err != nil {
			return nil, err
		} else {
			result = append(result, fields...)
		}
	case reflect.Map:
		if k := rt.Key(); k.Kind() != reflect.String {
			return nil, ErrBadParameter.Withf("%s: Only maps with string keys are supported", key)
		} else {
			return enumerate_type(key+"[string]", index, rt.Elem())
		}
	case reflect.Slice, reflect.Array:
		return enumerate_type(key+"[number]", index, rt.Elem())
	default:
		result = append(result, metafield{Key: key, Index: index, Type: rt})
	}

	// Return appended fields
	return result, nil
}
