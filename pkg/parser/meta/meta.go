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
	return meta, nil
}
