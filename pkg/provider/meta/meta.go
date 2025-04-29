package meta

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
	"time"

	// Packages
	"github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	ast "github.com/mutablelogic/go-server/pkg/parser/ast"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Meta struct {
	Name        string
	Description string
	Type        reflect.Type
	Fields      []*Meta
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new metadata object from a structure
func New(v server.Plugin) (*Meta, error) {
	meta := new(Meta)
	if v == nil {
		return nil, httpresponse.ErrInternalError.Withf("nil value")
	}
	rt := reflect.TypeOf(v)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, httpresponse.ErrInternalError.Withf("expected struct, got %T", v)
	} else {
		meta.Name = v.Name()
		meta.Description = v.Description()
		meta.Type = rt
	}

	// Get visible fields
	fields := reflect.VisibleFields(rt)
	meta.Fields = make([]*Meta, 0, len(fields))
	for _, field := range fields {
		if field, err := newMetaField(field); err != nil {
			return nil, httpresponse.ErrInternalError.With(err.Error())
		} else if field != nil {
			meta.Fields = append(meta.Fields, field)
		}
	}

	// Return success
	return meta, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *Meta) String() string {
	var buf bytes.Buffer
	if err := m.Write(&buf); err != nil {
		return err.Error()
	}
	return buf.String()
}

func (m *Meta) Write(w io.Writer) error {
	var buf bytes.Buffer

	if m.Description != "" {
		buf.WriteString("// ")
		buf.WriteString(m.Description)
		buf.WriteString("\n")
	}
	buf.WriteString(m.Name)
	buf.WriteString(" \"label\" {\n")

	for _, field := range m.Fields {
		buf.WriteString("  ")
		buf.WriteString(field.Name)
		buf.WriteString(" = ")
		buf.WriteString("<" + typeName(field.Type) + ">")

		if field.Description != "" {
			buf.WriteString("  // ")
			buf.WriteString(field.Description)
		}

		buf.WriteString("\n")
	}

	buf.WriteString("}\n")
	_, err := w.Write(buf.Bytes())
	return err
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *Meta) Validate(values any) error {
	dict := values.(map[string]ast.Node)
	for _, field := range m.Fields {
		fmt.Println(field.Name, "=>", dict[field.Name])
	}
	return nil
}

func (m *Meta) New() server.Plugin {
	return reflect.New(m.Type).Interface().(server.Plugin)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newMetaField(rf reflect.StructField) (*Meta, error) {
	meta := new(Meta)

	//fmt.Println("newMetaField", rf.Name, rf.Type)

	// Name
	if name := nameForField(rf, "json", "yaml", "name"); name == "" {
		// Ignore this field
		return nil, nil
	} else {
		meta.Name = name
	}

	// Description
	meta.Description = rf.Tag.Get("help")

	// Type
	if t := typeName(rf.Type); t == "" {
		return nil, httpresponse.ErrInternalError.Withf("unsupported type: %s", rf.Type)
	} else {
		meta.Type = rf.Type
	}

	return meta, nil
}

var (
	timeType     = reflect.TypeOf(time.Time{})
	urlType      = reflect.TypeOf((*url.URL)(nil)).Elem()
	durationType = reflect.TypeOf(time.Duration(0))
)

func typeName(rt reflect.Type) string {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	switch rt.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rt == durationType {
			return "duration"
		}
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "string"
	case reflect.Slice:
		if subtype := typeName(rt.Elem()); subtype != "" {
			return "list(" + subtype + ")"
		}
	case reflect.Map:
		if subtype := typeName(rt.Elem()); subtype != "" && rt.Key().Kind() == reflect.String {
			return "map(" + subtype + ")"
		}
	case reflect.Struct:
		if rt == urlType {
			return "url"
		}
		if rt == timeType {
			return "datetime"
		}
	}
	return "ref"
}

func nameForField(rt reflect.StructField, tags ...string) string {
	for _, tag := range tags {
		tag, ok := rt.Tag.Lookup(tag)
		if !ok {
			continue
		}
		if tag == "-" || tag == "" {
			// Ignore
			return ""
		}
		name := strings.Split(tag, ",")
		if len(name) > 0 && name[0] != "" {
			return name[0]
		}
	}
	return rt.Name
}
