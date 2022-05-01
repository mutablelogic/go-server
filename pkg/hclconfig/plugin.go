package hclconfig

import (
	"fmt"
	"path/filepath"
	"plugin"
	"reflect"
	"strings"
	"time"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"

	// Modules

	"github.com/hashicorp/hcl/v2/hcldec"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Plugin struct {
	// Name of the plugin
	Name string

	// Configuration type
	Config reflect.Type

	// Configuration spec
	Spec hcldec.Spec
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	nameMethodName   = "Name"
	configMethodName = "GetConfig"
	hclTagName       = "hcl"
	hclTagAttr       = "attr"
	hclTagBlock      = "block"
	hclTagLabel      = "label"
	hclTagOptional   = "optional"
)

var (
	typeString     = reflect.TypeOf("")
	typeListString = reflect.TypeOf([]string{})
	typeDuration   = reflect.TypeOf(time.Second)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// OpenPlugin will attempt to retrieve details of a plugin
func (c *Config) OpenPlugin(path string) (*Plugin, error) {
	this := new(Plugin)

	// Open a plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	// Get the name of the plugin
	if namefn, err := p.Lookup(nameMethodName); err != nil {
		return nil, ErrInternalAppError.Withf("%q: no %s method", filepath.Base(path), nameMethodName)
	} else if namefn_, ok := namefn.(func() string); !ok {
		return nil, ErrInternalAppError.Withf(" %q: invalid %s method", filepath.Base(path), nameMethodName)
	} else {
		this.Name = namefn_()
	}

	// Get the configuration prototype
	var prototype interface{}
	if configfn, err := p.Lookup(configMethodName); err != nil {
		return nil, ErrInternalAppError.Withf("%q: no %s method", filepath.Base(path), configMethodName)
	} else if configfn_, ok := configfn.(func() interface{}); !ok {
		return nil, ErrInternalAppError.Withf("%q: invalid %s method", filepath.Base(path), configMethodName)
	} else if prototype = configfn_(); prototype == nil {
		return nil, ErrInternalAppError.Withf("%q: nil value returned from %s method", filepath.Base(path), configMethodName)
	}

	// Check prototype is a pointer to a struct
	v := reflect.ValueOf(prototype)
	if v.Kind() != reflect.Ptr {
		return nil, ErrInternalAppError.Withf("%q: invalid value returned from %s method", filepath.Base(path), configMethodName)
	} else {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, ErrInternalAppError.Withf("%q: invalid value returned from %s method", filepath.Base(path), configMethodName)
	}

	// Set configuration prototype for this plugin
	if spec, err := getBlock(v.Type(), this.Name, false); err != nil {
		return nil, err
	} else {
		this.Config = v.Type()
		this.Spec = &hcldec.BlockSetSpec{
			TypeName: spec.TypeName,
			Nested:   spec.Nested,
			MinItems: 0,
			MaxItems: 0,
		}
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *Plugin) String() string {
	str := "<plugin"
	str += fmt.Sprintf(" name=%q", p.Name)
	str += fmt.Sprintf(" config=%v", p.Config)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (p *Plugin) Append(tuple cty.Value) error {
	// Create a new object from a prototype using the cty.Value values
	// as field values
	data, _ := json.Marshal(tuple, tuple.Type())
	fmt.Printf("%s: append %s\n", p.Name, string(data))
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func getBlock(t reflect.Type, name string, required bool) (*hcldec.BlockSpec, error) {
	// Validate that 't' is a struct
	if t.Kind() != reflect.Struct {
		return nil, ErrBadParameter.Withf("unsupported hcl type on %q", name)
	}

	// Cycle through the fields to construct the specification
	n := t.NumField()
	label := 0
	result := make([]hcldec.Spec, 0, n)
	for i := 0; i < n; i++ {
		field := t.Field(i)
		tag := field.Tag.Get(hclTagName)
		if tag == "" {
			continue
		}

		comma := strings.Index(tag, ",")
		var name, kind string
		if comma != -1 {
			name = tag[:comma]
			kind = tag[comma+1:]
		} else {
			name = tag
			kind = hclTagAttr
		}

		switch kind {
		case hclTagAttr:
			if spec, err := getAttr(field, name, true); err != nil {
				return nil, err
			} else {
				result = append(result, spec)
			}
		case "optional":
			if spec, err := getAttr(field, name, false); err != nil {
				return nil, err
			} else {
				result = append(result, spec)
			}
		case hclTagBlock:
			required := true
			if field.Type.Kind() == reflect.Ptr {
				// Block should be optional if pointer to struct
				field.Type = field.Type.Elem()
				required = false
			}
			if spec, err := getBlock(field.Type, name, required); err != nil {
				return nil, err
			} else {
				result = append(result, spec)
			}
		case "label":
			if spec, err := getLabel(field, name, label); err != nil {
				return nil, err
			} else {
				result = append(result, spec)
				label++
			}
		default:
			return nil, ErrBadParameter.Withf("invalid hcl field tag %q on %q", kind, field.Name)
		}
	}

	// Return success
	return &hcldec.BlockSpec{
		TypeName: name,
		Nested:   hcldec.TupleSpec(result),
		Required: required,
	}, nil
}

func getAttr(field reflect.StructField, name string, required bool) (*hcldec.AttrSpec, error) {
	if t := getType(field.Type); t == cty.NilType {
		return nil, ErrBadParameter.Withf("unsupported go type %q on struct field %q", field.Type.String(), field.Name)
	} else {
		return &hcldec.AttrSpec{
			Name:     name,
			Type:     t,
			Required: required,
		}, nil
	}
}

func getLabel(field reflect.StructField, name string, index int) (*hcldec.BlockLabelSpec, error) {
	t := getType(field.Type)
	if t != cty.String {
		return nil, ErrBadParameter.Withf("unsupported go type %q on struct field %q", field.Type.String(), field.Name)
	} else {
		return &hcldec.BlockLabelSpec{
			Index: index,
			Name:  name,
		}, nil
	}
}

func getType(t reflect.Type) cty.Type {
	switch t {
	case typeString:
		return cty.String
	case typeListString:
		return cty.List(cty.String)
	case typeDuration:
		return cty.DynamicPseudoType
	}
	// By default, return NilType for unsupported types
	return cty.NilType
}
