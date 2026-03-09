# jsonschema

Generates JSON Schemas from Go types, enriched with additional struct tag support. Supports structs, primitives, slices, and maps.

## Usage

```go
import "github.com/mutablelogic/go-server/pkg/jsonschema"

schema, err := jsonschema.For[MyStruct]()
```

`For[T]()` generates a schema for `T`, enriches it with the struct tags below, and caches the result — subsequent calls for the same type are free. `T` can be any Go type: a struct, `string`, `int`, `bool`, `[]string`, etc.

## Validation

```go
data := json.RawMessage(`{"name":"alice","age":30}`)
if err := schema.Validate(data); err != nil {
    // err describes the validation failure
}
```

`(*Schema).Validate(data json.RawMessage) error` unmarshals the JSON and validates it against the schema. Returns `nil` on success. Works for all JSON types — objects, strings, numbers, booleans, and arrays.

## Decode

```go
var cfg Config
if err := schema.Decode(data, &cfg); err != nil {
    // err is a validation or unmarshal failure
}
```

`(*Schema).Decode(data json.RawMessage, v any) error` combines unmarshalling, default application, and validation in one step. Works for all JSON types:

- **Objects**: applies schema defaults for any missing fields before validating, then unmarshals into `v`
- **Primitives and arrays**: validates then unmarshals into `v`

Fields with a `default:""` tag are automatically treated as optional so that defaults can be filled in. This can be overridden with an explicit `required:""` tag.

## Type mappings

| Go type | JSON Schema |
|---|---|
| `string` | `{"type":"string"}` |
| `bool` | `{"type":"boolean"}` |
| `int`, `int64`, etc. | `{"type":"integer"}` |
| `float32`, `float64` | `{"type":"number"}` |
| `[]T` | `{"type":"array"}` |
| `map[K]V` | `{"type":"object"}` |
| `struct` | `{"type":"object","properties":{…}}` |
| `time.Time` | `{"type":"string"}` |
| `time.Duration` | `{"type":"string","format":"duration"}` |

### time.Duration

`time.Duration` fields are represented as JSON strings using Go's standard duration syntax (e.g. `"5s"`, `"1h30m"`, `"250ms"`). `Decode` automatically parses the string and assigns the correct `time.Duration` value:

```go
type Config struct {
    Timeout time.Duration `json:"timeout" default:"30s"`
}

var cfg Config
schema.Decode(json.RawMessage(`{"timeout":"1m"}`), &cfg)
// cfg.Timeout == time.Minute
```

If `timeout` is omitted from the JSON, the default `"30s"` is applied before unmarshalling.

## Supported struct tags

| Tag | JSON Schema field | Applies to | Example |
|---|---|---|---|
| `jsonschema:"text"` | `description` | any | `jsonschema:"the user's name"` — takes priority over `help` |
| `help:"text"` | `description` | any | `help:"the user's name"` — fallback if `jsonschema` tag is absent |
| `enum:"a,b,c"` | `enum` | any | `enum:"red,green,blue"` — comma-separated, values are trimmed |
| `default:"value"` | `default` | any | `default:"42"` — type-aware: bools, ints, uints, and floats emit native JSON; `time.Duration` emits a duration string; all others emit a JSON string. Also removes the field from `required` so defaults can be applied. |
| `format:"value"` | `format` | any | `format:"date-time"` — any valid JSON Schema format string |
| `pattern:"regex"` | `pattern` | strings | `pattern:"^[a-z]+"` |
| `min:"N"` | `minimum` / `minLength` / `minItems` | numbers, strings, slices | `min:"1"` |
| `max:"N"` | `maximum` / `maxLength` / `maxItems` | numbers, strings, slices | `max:"100"` |
| `required:""` | parent `required` array | any | adds the field to the parent object's `required` list |
| `optional:""` | parent `required` array | any | removes the field from the parent object's `required` list |
| `readonly:""` | `readOnly` | any | marks the property as read-only |
| `deprecated:""` | `deprecated` | any | marks the property as deprecated |

### Tag precedence

- `optional` takes precedence over `required` if both are present on the same field.
- `jsonschema` takes precedence over `help` for descriptions.
- A field with `default:""` is implicitly optional; add `required:""` to override.

### Example

```go
type Config struct {
    Host    string        `json:"host"    required:""   help:"Hostname or IP address"`
    Port    int           `json:"port"    default:"8080" min:"1" max:"65535"`
    Scheme  string        `json:"scheme"  enum:"http,https" default:"https"`
    Timeout time.Duration `json:"timeout" default:"30s"  help:"Request timeout"`
    Tags    []string      `json:"tags"    min:"1" max:"10"`
    Slug    string        `json:"slug"    pattern:"^[a-z0-9-]+$"`
    Key     string        `json:"key"     deprecated:"" readonly:"" help:"Use scheme instead"`
}

schema, err := jsonschema.For[Config]()

// Validate raw JSON:
if err := schema.Validate(data); err != nil { ... }

// Unmarshal + apply defaults + validate in one step:
var cfg Config
if err := schema.Decode(data, &cfg); err != nil { ... }
```

Produces (in part):

```json
{
  "type": "object",
  "properties": {
    "host":    { "type": "string",  "description": "Hostname or IP address" },
    "port":    { "type": "integer", "default": 8080, "minimum": 1, "maximum": 65535 },
    "scheme":  { "type": "string",  "enum": ["http", "https"], "default": "https" },
    "timeout": { "type": "string",  "format": "duration", "default": "30s", "description": "Request timeout" },
    "tags":    { "type": "array",   "minItems": 1, "maxItems": 10 },
    "slug":    { "type": "string",  "pattern": "^[a-z0-9-]+$" },
    "key":     { "type": "string",  "description": "Use scheme instead", "readOnly": true, "deprecated": true }
  },
  "required": ["host"]
}
```
