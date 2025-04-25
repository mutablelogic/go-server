package meta

import (
	"bytes"
	"io"
	"reflect"

	// Packages
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Meta struct {
	Name   string
	Type   reflect.Type
	Fields []*Meta
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new metadata object from a structure
func New(v any, name string) (*Meta, error) {
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
		meta.Name = name
		meta.Type = rt
	}

	// Get visible fields
	fields := reflect.VisibleFields(rt)
	meta.Fields = make([]*Meta, 0, len(fields))
	for _, field := range fields {
		if field, err := newMetaField(field); err != nil {
			return nil, httpresponse.ErrInternalError.With(err.Error())
		} else {
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

	buf.WriteString(m.Name)
	buf.WriteString(" \"label\" {\n")

	for _, field := range m.Fields {
		buf.WriteString("  ")
		buf.WriteString(field.Name)
		buf.WriteString(" ")
		buf.WriteString(typeName(field.Type))
		buf.WriteString("\n")
	}

	buf.WriteString("}\n")
	_, err := w.Write(buf.Bytes())
	return err
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newMetaField(rf reflect.StructField) (*Meta, error) {
	meta := new(Meta)
	meta.Name = rf.Name
	if t := typeName(rf.Type); t == "" {
		return nil, httpresponse.ErrInternalError.Withf("unsupported type: %s", rf.Type)
	} else {
		meta.Type = rf.Type
	}
	return meta, nil
}

func typeName(rt reflect.Type) string {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	switch rt.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
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
	}
	return ""
}
