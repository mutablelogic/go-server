package provider

import (
	"encoding/json"
	"fmt"
	"io"

	// Packages
	"github.com/mutablelogic/go-server"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

type JSONParser struct {
	plugin map[string]server.Plugin
}

func NewJSONParser(plugins ...server.Plugin) (*JSONParser, error) {
	parser := new(JSONParser)
	parser.plugin = make(map[string]server.Plugin, len(plugins))
	for _, plugin := range plugins {
		name := plugin.Name()
		if _, exists := parser.plugin[name]; exists {
			return nil, ErrDuplicateEntry.Withf("plugin %q already exists", plugin.Name())
		} else {
			parser.plugin[name] = plugin
		}
	}

	// Return success
	return parser, nil
}

func (p *JSONParser) Read(r io.Reader) error {
	dec := json.NewDecoder(r)
	path := []string{}
	state := -1
	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if t == nil {
			fmt.Printf("%v: nil %v\n", path, t)
			continue
		}
		switch t := t.(type) {
		case json.Delim:
			switch t {
			case '{':
				switch state {
				case -1: // We start processing
					state = 0
				case 1: // We are in a plugin
					state = 2
				case 3: // We are in a object
					// Plugin field value - object
					fmt.Printf("%v: object %v\n", path, t)
					state = 2
				default:
					return ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
				}
			case '}':
				switch state {
				case 2:
					path = path[:len(path)-1]
					// End of file or plugin
					if len(path) == 0 {
						state = 0
					} else {
						state = 2
					}
				case 0:
					// End of file
					state = -1
				default:
					return ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
				}
			case '[':
				switch state {
				case 3:
					// Plugin field value - array
					fmt.Printf("%v: array %v\n", path, t)
					state = 4
				default:
					return ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
				}
			case ']':
				switch state {
				case 4:
					// End of array
					fmt.Printf("%v: end array %v\n", path, t)
					state = 2
					path = path[:len(path)-1]
				default:
					return ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
				}
			}
		case string:
			switch state {
			case 0:
				// Plugin name and label
				fmt.Printf("%v: plugin %v\n", path, t)
				path = append(path, t)
				state = 1
			case 2:
				// Plugin field name
				fmt.Printf("%v: field name %v\n", path, t)
				path = append(path, t)
				state = 3
			case 3:
				// Plugin field value
				fmt.Printf("%v: field value %v\n", path, t)
				path = path[:len(path)-1]
				state = 2
			case 4:
				// Array value
				fmt.Printf("%v: array elem %v\n", path, t)
				state = 4
			default:
				return ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
			}
		case bool:
			switch state {
			case 3:
				// Plugin field value
				fmt.Printf("%v: field value %v\n", path, t)
				path = path[:len(path)-1]
				state = 2
			default:
				return ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
			}
		case float64:
			switch state {
			case 3:
				// Plugin field value
				fmt.Printf("%v: field value %v\n", path, t)
				path = path[:len(path)-1]
				state = 2
			default:
				return ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
			}
		default:
			return ErrBadParameter.Withf("Unexpected token: %v (%T)", t, t)
		}
	}

	// The end state should be -1
	if state != -1 {
		return ErrInternalAppError.With("Unexpected end state")
	}

	// Return success
	return nil
}
