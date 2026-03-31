package jsonschema

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

///////////////////////////////////////////////////////////////////////////////
// TEST STRUCTS

type simpleStruct struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type enumStruct struct {
	Color string `json:"color" enum:"red,green,blue"`
	Size  string `json:"size"  enum:"small, medium, large"`
}

type noJSONTagStruct struct {
	Title string
	Body  string `json:"-"`
}

type nestedStruct struct {
	Meta  innerStruct `json:"meta"`
	Label string      `json:"label"`
}

type innerStruct struct {
	Kind string `json:"kind" enum:"foo,bar"`
}

type ptrStruct struct {
	Role *string `json:"role" enum:"admin,user,guest"`
}

type omitemptyStruct struct {
	Tag string `json:"tag,omitempty" enum:"x,y,z"`
}

// json:",omitempty" — name before comma is empty, so field name is used
type emptyJSONNameStruct struct {
	Value string `json:",omitempty" enum:"p,q"`
}

// unexported field — upstream won't create a property, exercises prop == nil
type withUnexportedField struct {
	Name   string `json:"name"`
	secret string //nolint:unused
}

type formatStruct struct {
	CreatedAt string `json:"created_at" format:"date-time"`
	Count     int    `json:"count"      format:"int32"`
}

type helpStruct struct {
	Name  string `json:"name"  jsonschema:"the name"  help:"fallback name"`
	Email string `json:"email" help:"email address"`
	Bio   string `json:"bio"`
}

type defaultStruct struct {
	Name    string  `json:"name"    default:"alice"`
	Count   int     `json:"count"   default:"42"`
	Enabled bool    `json:"enabled" default:"true"`
	Score   float64 `json:"score"   default:"3.14"`
}

type requiredStruct struct {
	Name  string `json:"name"  required:""`
	Email string `json:"email" required:""`
	Bio   string `json:"bio"`
}

type optionalStruct struct {
	Name string `json:"name" optional:""`
}

type minMaxStruct struct {
	Name  string   `json:"name"   min:"2"  max:"50"`
	Age   int      `json:"age"    min:"0"  max:"150"`
	Score float64  `json:"score"  min:"0.0" max:"10.0"`
	Tags  []string `json:"tags"  min:"1"  max:"5"`
}

type deprecatedStruct struct {
	OldField string `json:"old_field" deprecated:""`
	NewField string `json:"new_field"`
}

type patternStruct struct {
	Slug string `json:"slug" pattern:"^[a-z0-9-]+$"`
}

type readonlyStruct struct {
	ID   string `json:"id"   readonly:""`
	Name string `json:"name"`
}

type uuidStruct struct {
	ID string `json:"id" format:"uuid"`
}

type uuidNativeStruct struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type uuidRequiredStruct struct {
	ID   string `json:"id"   format:"uuid" required:""`
	Name string `json:"name"`
}

type uuidDefaultStruct struct {
	ID string `json:"id" format:"uuid" default:"00000000-0000-0000-0000-000000000000"`
}

type uuidPointerStruct struct {
	ID *string `json:"id" format:"uuid" optional:""`
}

type uuidDescStruct struct {
	ID string `json:"id" format:"uuid" jsonschema:"resource identifier"`
}

type uuidReadonlyStruct struct {
	ID   string `json:"id"   format:"uuid" readonly:""`
	Name string `json:"name"`
}

type exampleStruct struct {
	ID    string  `json:"id"    format:"uuid"  example:"123e4567-e89b-12d3-a456-426614174000"`
	Name  string  `json:"name"  example:"alice"`
	Age   int     `json:"age"   example:"30"`
	Score float64 `json:"score" example:"9.5"`
	Flag  bool    `json:"flag"  example:"true"`
}

type nestedUUIDStruct struct {
	Inner uuidStruct `json:"inner"`
	Label string     `json:"label"`
}

///////////////////////////////////////////////////////////////////////////////
// TYPE MAPPING TESTS

func TestFor_TypeMappings(t *testing.T) {
	type AllTypes struct {
		Str     string   `json:"str"`
		Bool    bool     `json:"bool"`
		Int     int      `json:"int"`
		Int8    int8     `json:"int8"`
		Int64   int64    `json:"int64"`
		Uint    uint     `json:"uint"`
		Float32 float32  `json:"float32"`
		Float64 float64  `json:"float64"`
		Slice   []string `json:"slice"`
		Array   [3]int   `json:"array"`
	}
	s, err := For[AllTypes]()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		field    string
		wantType string
	}{
		{"str", "string"},
		{"bool", "boolean"},
		{"int", "integer"},
		{"int8", "integer"},
		{"int64", "integer"},
		{"uint", "integer"},
		{"float32", "number"},
		{"float64", "number"},
		{"slice", "array"},
		{"array", "array"},
	}
	for _, tc := range cases {
		prop := s.Properties[tc.field]
		if prop == nil {
			t.Errorf("%s: property not found", tc.field)
			continue
		}
		// Upstream uses Type (single) or Types (multi) depending on nullability.
		if prop.Type != tc.wantType && !sliceContains(prop.Types, tc.wantType) {
			t.Errorf("%s: Type=%q Types=%v, want %q", tc.field, prop.Type, prop.Types, tc.wantType)
		}
	}
}

func TestFor_FixedArrayMinMaxItems(t *testing.T) {
	type S struct {
		Trio [3]int `json:"trio"`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["trio"]
	if prop == nil {
		t.Fatal("expected property 'trio'")
	}
	if prop.MinItems == nil || *prop.MinItems != 3 {
		t.Errorf("MinItems: got %v, want 3", prop.MinItems)
	}
	if prop.MaxItems == nil || *prop.MaxItems != 3 {
		t.Errorf("MaxItems: got %v, want 3", prop.MaxItems)
	}
}

func TestFor_PointerField_AllowsNull(t *testing.T) {
	type S struct {
		Name *string `json:"name" optional:""`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["name"]
	if prop == nil {
		t.Fatal("expected property 'name'")
	}
	// An optional *T field is not required, so null must still be allowed.
	hasString := prop.Type == "string" || sliceContains(prop.Types, "string")
	hasNull := prop.Type == "null" || sliceContains(prop.Types, "null")
	if !hasString {
		t.Errorf("expected string type in optional pointer field schema, got Type=%q Types=%v", prop.Type, prop.Types)
	}
	if !hasNull {
		t.Errorf("expected null type in optional pointer field schema, got Type=%q Types=%v", prop.Type, prop.Types)
	}
}

func TestFor_PointerField_RequiredStripsNull(t *testing.T) {
	type S struct {
		Name *string `json:"name" required:""`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["name"]
	if prop == nil {
		t.Fatal("expected property 'name'")
	}
	hasString := prop.Type == "string" || sliceContains(prop.Types, "string")
	hasNull := prop.Type == "null" || sliceContains(prop.Types, "null")
	if !hasString {
		t.Errorf("expected string type, got Type=%q Types=%v", prop.Type, prop.Types)
	}
	if hasNull {
		t.Errorf("required pointer field must not allow null, got Type=%q Types=%v", prop.Type, prop.Types)
	}
	required := false
	for _, r := range s.Required {
		if r == "name" {
			required = true
		}
	}
	if !required {
		t.Error("expected 'name' in Required")
	}
}

func TestFor_MapField_ObjectType(t *testing.T) {
	type S struct {
		Meta map[string]string `json:"meta"`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["meta"]
	if prop == nil {
		t.Fatal("expected property 'meta'")
	}
	if prop.Type != "object" {
		t.Errorf("map field type: got %q, want %q", prop.Type, "object")
	}
}

func TestFor_InterfaceField_NoType(t *testing.T) {
	type S struct {
		Data any `json:"data"`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["data"]
	if prop == nil {
		t.Fatal("expected property 'data'")
	}
	// interface{}/any produces an unrestricted schema — no type constraint.
	if prop.Type != "" {
		t.Errorf("any field should have no type constraint, got %q", prop.Type)
	}
}

func sliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

///////////////////////////////////////////////////////////////////////////////
// TESTS

func TestFor_SimpleStruct(t *testing.T) {
	s, err := For[simpleStruct]()
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
	if len(s.Properties) == 0 {
		t.Fatal("expected properties")
	}
	if _, ok := s.Properties["name"]; !ok {
		t.Error("expected property 'name'")
	}
	if _, ok := s.Properties["count"]; !ok {
		t.Error("expected property 'count'")
	}
}

func TestFor_EnumTag(t *testing.T) {
	s, err := For[enumStruct]()
	if err != nil {
		t.Fatal(err)
	}
	color := s.Properties["color"]
	if color == nil {
		t.Fatal("expected property 'color'")
	}
	if len(color.Enum) != 3 {
		t.Fatalf("expected 3 enum values for color, got %d: %v", len(color.Enum), color.Enum)
	}
	wantColor := []any{"red", "green", "blue"}
	for i, v := range wantColor {
		if color.Enum[i] != v {
			t.Errorf("color.Enum[%d]: got %v, want %v", i, color.Enum[i], v)
		}
	}

	size := s.Properties["size"]
	if size == nil {
		t.Fatal("expected property 'size'")
	}
	// values should be trimmed
	wantSize := []any{"small", "medium", "large"}
	if len(size.Enum) != len(wantSize) {
		t.Fatalf("expected %d enum values for size, got %d: %v", len(wantSize), len(size.Enum), size.Enum)
	}
	for i, v := range wantSize {
		if size.Enum[i] != v {
			t.Errorf("size.Enum[%d]: got %v, want %v", i, size.Enum[i], v)
		}
	}
}

func TestFor_NoJSONTag_UsesFieldName(t *testing.T) {
	s, err := For[noJSONTagStruct]()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Properties["Title"]; !ok {
		t.Error("expected property 'Title' (no json tag → use field name)")
	}
}

func TestFor_JSONDashTag_Excluded(t *testing.T) {
	s, err := For[noJSONTagStruct]()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Properties["Body"]; ok {
		t.Error("property 'Body' with json:\"-\" should be excluded")
	}
	if _, ok := s.Properties["-"]; ok {
		t.Error("property '-' should not appear")
	}
}

func TestFor_NestedStruct_EnumPropagates(t *testing.T) {
	s, err := For[nestedStruct]()
	if err != nil {
		t.Fatal(err)
	}
	meta := s.Properties["meta"]
	if meta == nil {
		t.Fatal("expected property 'meta'")
	}
	kind := meta.Properties["kind"]
	if kind == nil {
		t.Fatal("expected nested property 'meta.kind'")
	}
	if len(kind.Enum) != 2 {
		t.Fatalf("expected 2 enum values for meta.kind, got %d: %v", len(kind.Enum), kind.Enum)
	}
}

func TestFor_PointerField_EnumApplied(t *testing.T) {
	s, err := For[ptrStruct]()
	if err != nil {
		t.Fatal(err)
	}
	role := s.Properties["role"]
	if role == nil {
		t.Fatal("expected property 'role'")
	}
	if len(role.Enum) != 3 {
		t.Fatalf("expected 3 enum values for role, got %d: %v", len(role.Enum), role.Enum)
	}
}

func TestFor_OmitemptyJSONTag_NameStripped(t *testing.T) {
	s, err := For[omitemptyStruct]()
	if err != nil {
		t.Fatal(err)
	}
	tag := s.Properties["tag"]
	if tag == nil {
		t.Fatal("expected property 'tag' (omitempty suffix stripped)")
	}
	if len(tag.Enum) != 3 {
		t.Fatalf("expected 3 enum values, got %d: %v", len(tag.Enum), tag.Enum)
	}
}

func TestFor_NoEnumTag_EnumNil(t *testing.T) {
	s, err := For[simpleStruct]()
	if err != nil {
		t.Fatal(err)
	}
	name := s.Properties["name"]
	if name == nil {
		t.Fatal("expected property 'name'")
	}
	if name.Enum != nil {
		t.Errorf("expected nil Enum for field without enum tag, got %v", name.Enum)
	}
}

// For[*T] — exercises the pointer-dereference loop at the top of enrichSchema.
func TestFor_PointerToStruct_EnumApplied(t *testing.T) {
	s, err := For[*enumStruct]()
	if err != nil {
		t.Fatal(err)
	}
	color := s.Properties["color"]
	if color == nil {
		t.Fatal("expected property 'color'")
	}
	if len(color.Enum) != 3 {
		t.Fatalf("expected 3 enum values, got %d: %v", len(color.Enum), color.Enum)
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS FOR FromJSON

func TestFromJSON_ValidSchema(t *testing.T) {
	raw := json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age":  {"type": "integer"}
		},
		"required": ["name"]
	}`)
	schema, err := FromJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Valid input passes.
	if err := schema.Validate(json.RawMessage(`{"name":"alice","age":30}`)); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	// Missing required field fails.
	if err := schema.Validate(json.RawMessage(`{"age":30}`)); err == nil {
		t.Error("expected error for missing required field, got nil")
	}
	// Wrong type fails.
	if err := schema.Validate(json.RawMessage(`{"name":42}`)); err == nil {
		t.Error("expected error for wrong type, got nil")
	}
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	if _, err := FromJSON(json.RawMessage(`{not json}`)); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestFromJSON_StringSchema(t *testing.T) {
	schema, err := FromJSON(json.RawMessage(`{"type":"string","minLength":3}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := schema.Validate(json.RawMessage(`"hello"`)); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if err := schema.Validate(json.RawMessage(`"hi"`)); err == nil {
		t.Error("expected error for string shorter than minLength, got nil")
	}
}

// For[T] with primitive and slice types — non-struct types generate a valid schema without error.
func TestFor_NonStruct_Succeeds(t *testing.T) {
	// Primitive and slice types should generate a valid schema without error.
	if _, err := For[int](); err != nil {
		t.Errorf("For[int]: unexpected error: %v", err)
	}
	if _, err := For[string](); err != nil {
		t.Errorf("For[string]: unexpected error: %v", err)
	}
	if _, err := For[[]string](); err != nil {
		t.Errorf("For[[]string]: unexpected error: %v", err)
	}
	if _, err := For[bool](); err != nil {
		t.Errorf("For[bool]: unexpected error: %v", err)
	}
}

// struct with an unexported field — exercises the prop == nil branch.
func TestFor_UnexportedField_NoCrash(t *testing.T) {
	s, err := For[withUnexportedField]()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Properties["name"]; !ok {
		t.Error("expected exported property 'name'")
	}
}

// json:",omitempty" (empty name before comma) — exercises the empty-name
// branch in jsonFieldName which falls back to the struct field Name.
func TestFor_EmptyJSONName_UsesFieldName(t *testing.T) {
	s, err := For[emptyJSONNameStruct]()
	if err != nil {
		t.Fatal(err)
	}
	// upstream uses field name "Value" when json name is empty
	prop := s.Properties["Value"]
	if prop == nil {
		t.Fatal("expected property 'Value' when json tag name is empty")
	}
	if len(prop.Enum) != 2 {
		t.Fatalf("expected 2 enum values, got %d: %v", len(prop.Enum), prop.Enum)
	}
}

///////////////////////////////////////////////////////////////////////////////
// TAG TESTS: format, default, help, required/optional

func TestFor_HelpTag(t *testing.T) {
	s, err := For[helpStruct]()
	if err != nil {
		t.Fatal(err)
	}
	// jsonschema tag is set — help should NOT override it
	if got := s.Properties["name"].Description; got != "the name" {
		t.Errorf("name description: got %q, want %q", got, "the name")
	}
	// no jsonschema tag — help is used as fallback
	if got := s.Properties["email"].Description; got != "email address" {
		t.Errorf("email description: got %q, want %q", got, "email address")
	}
	// neither tag — description stays empty
	if got := s.Properties["bio"].Description; got != "" {
		t.Errorf("bio description: got %q, want empty", got)
	}
}

func TestFor_FormatTag(t *testing.T) {
	s, err := For[formatStruct]()
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Properties["created_at"].Format; got != "date-time" {
		t.Errorf("created_at format: got %q, want %q", got, "date-time")
	}
	if got := s.Properties["count"].Format; got != "int32" {
		t.Errorf("count format: got %q, want %q", got, "int32")
	}
}

func TestFor_DefaultTag(t *testing.T) {
	s, err := For[defaultStruct]()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		field string
		want  string
	}{
		{"name", `"alice"`},
		{"count", `42`},
		{"enabled", `true`},
		{"score", `3.14`},
	}
	for _, tc := range tests {
		prop := s.Properties[tc.field]
		if prop == nil {
			t.Fatalf("missing property %q", tc.field)
		}
		if got := string(prop.Default); got != tc.want {
			t.Errorf("default[%s]: got %s, want %s", tc.field, got, tc.want)
		}
	}
}

func TestFor_KongRequiredTag(t *testing.T) {
	s, err := For[requiredStruct]()
	if err != nil {
		t.Fatal(err)
	}
	required := map[string]bool{}
	for _, r := range s.Required {
		required[r] = true
	}
	if !required["name"] {
		t.Error("expected 'name' in Required")
	}
	if !required["email"] {
		t.Error("expected 'email' in Required")
	}
	// bio has no kong tag; upstream marks it required by default — this is unchanged
	if !required["bio"] {
		t.Error("expected 'bio' in Required (upstream default)")
	}
}

func TestFor_KongOptionalTag_RemovesFromRequired(t *testing.T) {
	// The upstream library may or may not add fields to Required by default;
	// what matters is that after enrichment, optionally-tagged fields are absent.
	s, err := For[optionalStruct]()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range s.Required {
		if r == "name" {
			t.Error("'name' should not be in Required after kong:optional")
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
// DIRECT TESTS FOR marshalDefault, appendUnique, removeString

func TestFor_MinMaxString(t *testing.T) {
	s, err := For[minMaxStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["name"]
	if prop == nil {
		t.Fatal("expected property 'name'")
	}
	if prop.MinLength == nil || *prop.MinLength != 2 {
		t.Errorf("MinLength: got %v, want 2", prop.MinLength)
	}
	if prop.MaxLength == nil || *prop.MaxLength != 50 {
		t.Errorf("MaxLength: got %v, want 50", prop.MaxLength)
	}
}

func TestFor_MinMaxInt(t *testing.T) {
	s, err := For[minMaxStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["age"]
	if prop == nil {
		t.Fatal("expected property 'age'")
	}
	if prop.Minimum == nil || *prop.Minimum != 0 {
		t.Errorf("Minimum: got %v, want 0", prop.Minimum)
	}
	if prop.Maximum == nil || *prop.Maximum != 150 {
		t.Errorf("Maximum: got %v, want 150", prop.Maximum)
	}
}

func TestFor_MinMaxFloat(t *testing.T) {
	s, err := For[minMaxStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["score"]
	if prop == nil {
		t.Fatal("expected property 'score'")
	}
	if prop.Minimum == nil || *prop.Minimum != 0.0 {
		t.Errorf("Minimum: got %v, want 0.0", prop.Minimum)
	}
	if prop.Maximum == nil || *prop.Maximum != 10.0 {
		t.Errorf("Maximum: got %v, want 10.0", prop.Maximum)
	}
}

func TestFor_MinMaxSlice(t *testing.T) {
	s, err := For[minMaxStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["tags"]
	if prop == nil {
		t.Fatal("expected property 'tags'")
	}
	if prop.MinItems == nil || *prop.MinItems != 1 {
		t.Errorf("MinItems: got %v, want 1", prop.MinItems)
	}
	if prop.MaxItems == nil || *prop.MaxItems != 5 {
		t.Errorf("MaxItems: got %v, want 5", prop.MaxItems)
	}
}

func TestFor_Deprecated(t *testing.T) {
	s, err := For[deprecatedStruct]()
	if err != nil {
		t.Fatal(err)
	}
	old := s.Properties["old_field"]
	if old == nil {
		t.Fatal("expected property 'old_field'")
	}
	if !old.Deprecated {
		t.Error("expected Deprecated=true on old_field")
	}
	newProp := s.Properties["new_field"]
	if newProp == nil {
		t.Fatal("expected property 'new_field'")
	}
	if newProp.Deprecated {
		t.Error("expected Deprecated=false on new_field")
	}
}

func TestFor_PatternTag(t *testing.T) {
	s, err := For[patternStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["slug"]
	if prop == nil {
		t.Fatal("expected property 'slug'")
	}
	if prop.Pattern != "^[a-z0-9-]+$" {
		t.Errorf("Pattern: got %q, want %q", prop.Pattern, "^[a-z0-9-]+$")
	}
}

func TestFor_ReadonlyTag(t *testing.T) {
	s, err := For[readonlyStruct]()
	if err != nil {
		t.Fatal(err)
	}
	id := s.Properties["id"]
	if id == nil {
		t.Fatal("expected property 'id'")
	}
	if !id.ReadOnly {
		t.Error("expected ReadOnly=true on id")
	}
	name := s.Properties["name"]
	if name == nil {
		t.Fatal("expected property 'name'")
	}
	if name.ReadOnly {
		t.Error("expected ReadOnly=false on name")
	}
}

///////////////////////////////////////////////////////////////////////////////
// DIRECT TESTS FOR marshalDefault, appendUnique, removeString

func TestMarshalDefault(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		val  string
		want string
	}{
		// bool
		{"bool true", reflect.TypeOf(false), "true", "true"},
		{"bool false", reflect.TypeOf(false), "false", "false"},
		{"bool invalid → string fallback", reflect.TypeOf(false), "yes", `"yes"`},
		// int
		{"int", reflect.TypeOf(0), "42", "42"},
		{"int negative", reflect.TypeOf(0), "-7", "-7"},
		{"int invalid → string fallback", reflect.TypeOf(0), "abc", `"abc"`},
		// int variants
		{"int8", reflect.TypeOf(int8(0)), "127", "127"},
		{"int16", reflect.TypeOf(int16(0)), "1000", "1000"},
		{"int32", reflect.TypeOf(int32(0)), "32000", "32000"},
		{"int64", reflect.TypeOf(int64(0)), "9999999", "9999999"},
		// uint
		{"uint", reflect.TypeOf(uint(0)), "10", "10"},
		{"uint8", reflect.TypeOf(uint8(0)), "255", "255"},
		{"uint16", reflect.TypeOf(uint16(0)), "65535", "65535"},
		{"uint32", reflect.TypeOf(uint32(0)), "4294967295", "4294967295"},
		{"uint64", reflect.TypeOf(uint64(0)), "18446744073709551615", "18446744073709551615"},
		{"uint invalid → string fallback", reflect.TypeOf(uint(0)), "-1", `"-1"`},
		// float
		{"float32", reflect.TypeOf(float32(0)), "1.5", "1.5"},
		{"float64", reflect.TypeOf(float64(0)), "3.14", "3.14"},
		{"float invalid → string fallback", reflect.TypeOf(float64(0)), "nan!", `"nan!"`},
		// pointer — should dereference
		{"*int", reflect.TypeOf((*int)(nil)), "99", "99"},
		{"*bool", reflect.TypeOf((*bool)(nil)), "true", "true"},
		// string / other → JSON string
		{"string", reflect.TypeOf(""), "hello", `"hello"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := marshalDefault(tc.typ, tc.val)
			if string(got) != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}

func TestAppendUnique(t *testing.T) {
	// adding a new element
	got := appendUnique([]string{"a", "b"}, "c")
	if len(got) != 3 || got[2] != "c" {
		t.Errorf("expected append, got %v", got)
	}
	// duplicate — should not grow
	got2 := appendUnique([]string{"a", "b"}, "a")
	if len(got2) != 2 {
		t.Errorf("expected no-op for duplicate, got %v", got2)
	}
}

func TestRemoveString(t *testing.T) {
	// element present — should remove
	got := removeString([]string{"a", "b", "c"}, "b")
	if len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Errorf("expected removal, got %v", got)
	}
	// element absent — should return unchanged
	got2 := removeString([]string{"a", "b"}, "z")
	if len(got2) != 2 {
		t.Errorf("expected no-op for absent element, got %v", got2)
	}
}

///////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////
// TESTS FOR Decode

func TestDecode_AppliesDefault(t *testing.T) {
	type Config struct {
		Name string `json:"name"`
		Port int    `json:"port" default:"8080"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := schema.Decode(json.RawMessage(`{"name":"alice"}`), &c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "alice" {
		t.Errorf("Name: got %q, want %q", c.Name, "alice")
	}
	if c.Port != 8080 {
		t.Errorf("Port: got %d, want 8080", c.Port)
	}
}

func TestDecode_ValidationError(t *testing.T) {
	type Config struct {
		Count int `json:"count" min:"1" max:"10"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := schema.Decode(json.RawMessage(`{"count":99}`), &c); err == nil {
		t.Error("expected validation error for out-of-range value, got nil")
	}
}

func TestDecode_InvalidJSON(t *testing.T) {
	type Config struct {
		Name string `json:"name"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := schema.Decode(json.RawMessage(`{not json}`), &c); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDecode_StringSchema(t *testing.T) {
	schema, err := For[string]()
	if err != nil {
		t.Fatal(err)
	}
	var s string
	if err := schema.Decode(json.RawMessage(`"hello"`), &s); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if s != "hello" {
		t.Errorf("got %q, want %q", s, "hello")
	}
	if err := schema.Decode(json.RawMessage(`42`), &s); err == nil {
		t.Error("expected error for number against string schema, got nil")
	}
}

func TestDecode_IntSchema(t *testing.T) {
	schema, err := For[int]()
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := schema.Decode(json.RawMessage(`7`), &n); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if n != 7 {
		t.Errorf("got %d, want 7", n)
	}
	if err := schema.Decode(json.RawMessage(`"oops"`), &n); err == nil {
		t.Error("expected error for string against integer schema, got nil")
	}
}

func TestDecode_BoolSchema(t *testing.T) {
	schema, err := For[bool]()
	if err != nil {
		t.Fatal(err)
	}
	var b bool
	if err := schema.Decode(json.RawMessage(`true`), &b); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !b {
		t.Error("got false, want true")
	}
	if err := schema.Decode(json.RawMessage(`"yes"`), &b); err == nil {
		t.Error("expected error for string against boolean schema, got nil")
	}
}

func TestDecode_ArraySchema(t *testing.T) {
	schema, err := For[[]string]()
	if err != nil {
		t.Fatal(err)
	}
	var ss []string
	if err := schema.Decode(json.RawMessage(`["a","b","c"]`), &ss); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(ss) != 3 || ss[0] != "a" || ss[1] != "b" || ss[2] != "c" {
		t.Errorf("got %v, want [a b c]", ss)
	}
	if err := schema.Decode(json.RawMessage(`"not-an-array"`), &ss); err == nil {
		t.Error("expected error for string against array schema, got nil")
	}
}

func TestFor_TimeDuration_StringSchema(t *testing.T) {
	schema, err := For[time.Duration]()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(schema)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if m["type"] != "string" {
		t.Errorf("type: got %v, want string", m["type"])
	}
	if m["format"] != "duration" {
		t.Errorf("format: got %v, want duration", m["format"])
	}
}

func TestFor_TimeDuration_StructField(t *testing.T) {
	type Config struct {
		Timeout time.Duration `json:"timeout"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(schema)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	props := m["properties"].(map[string]any)
	timeout := props["timeout"].(map[string]any)
	if timeout["type"] != "string" {
		t.Errorf("timeout type: got %v, want string", timeout["type"])
	}
	if timeout["format"] != "duration" {
		t.Errorf("timeout format: got %v, want duration", timeout["format"])
	}
}

func TestValidate_TimeDuration(t *testing.T) {
	schema, err := For[time.Duration]()
	if err != nil {
		t.Fatal(err)
	}
	// Valid duration string.
	if err := schema.Validate(json.RawMessage(`"5s"`)); err != nil {
		t.Errorf("expected no error for duration string, got: %v", err)
	}
	// Integer should fail (schema expects string).
	if err := schema.Validate(json.RawMessage(`5000000000`)); err == nil {
		t.Error("expected error for integer against duration schema, got nil")
	}
}

func TestDecode_TimeDuration(t *testing.T) {
	schema, err := For[time.Duration]()
	if err != nil {
		t.Fatal(err)
	}
	var d time.Duration
	if err := schema.Decode(json.RawMessage(`"5s"`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 5*time.Second {
		t.Errorf("got %v, want 5s", d)
	}
	if err := schema.Decode(json.RawMessage(`"1h30m"`), &d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 90*time.Minute {
		t.Errorf("got %v, want 1h30m0s", d)
	}
	if err := schema.Decode(json.RawMessage(`"not-a-duration"`), &d); err == nil {
		t.Error("expected error for invalid duration string, got nil")
	}
	// Integer should fail schema validation.
	if err := schema.Decode(json.RawMessage(`5000000000`), &d); err == nil {
		t.Error("expected error for integer against duration schema, got nil")
	}
}

func TestDecode_StructWithDuration(t *testing.T) {
	type Config struct {
		Name    string        `json:"name"`
		Timeout time.Duration `json:"timeout"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := schema.Decode(json.RawMessage(`{"name":"svc","timeout":"30s"}`), &c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name != "svc" {
		t.Errorf("Name: got %q, want svc", c.Name)
	}
	if c.Timeout != 30*time.Second {
		t.Errorf("Timeout: got %v, want 30s", c.Timeout)
	}
	// Invalid duration string in struct.
	if err := schema.Decode(json.RawMessage(`{"name":"svc","timeout":"bad"}`), &c); err == nil {
		t.Error("expected error for invalid duration in struct, got nil")
	}
}

func TestDecode_StructWithDuration_Default(t *testing.T) {
	type Config struct {
		Timeout time.Duration `json:"timeout" default:"10s"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := schema.Decode(json.RawMessage(`{}`), &c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Timeout != 10*time.Second {
		t.Errorf("Timeout: got %v, want 10s", c.Timeout)
	}
}

func TestDecode_NestedStructWithDuration(t *testing.T) {
	type Inner struct {
		Timeout time.Duration `json:"timeout"`
	}
	type Outer struct {
		Name string `json:"name"`
		DB   Inner  `json:"db"`
	}
	schema, err := For[Outer]()
	if err != nil {
		t.Fatal(err)
	}
	var o Outer
	if err := schema.Decode(json.RawMessage(`{"name":"svc","db":{"timeout":"5s"}}`), &o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if o.Name != "svc" {
		t.Errorf("Name: got %q, want svc", o.Name)
	}
	if o.DB.Timeout != 5*time.Second {
		t.Errorf("DB.Timeout: got %v, want 5s", o.DB.Timeout)
	}
	// Invalid nested duration string.
	if err := schema.Decode(json.RawMessage(`{"name":"svc","db":{"timeout":"bad"}}`), &o); err == nil {
		t.Error("expected error for invalid nested duration, got nil")
	}
}

func TestDecode_PointerToPrimitive(t *testing.T) {
	schema, err := For[string]()
	if err != nil {
		t.Fatal(err)
	}
	var sp *string
	if err := schema.Decode(json.RawMessage(`"hello"`), &sp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sp == nil || *sp != "hello" {
		t.Errorf("got %v, want pointer to \"hello\"", sp)
	}
}

func TestDecode_PointerToStruct(t *testing.T) {
	type Config struct {
		Name string `json:"name"`
		Port int    `json:"port" default:"8080"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	var cp *Config
	if err := schema.Decode(json.RawMessage(`{"name":"alice"}`), &cp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp == nil {
		t.Fatal("expected non-nil pointer, got nil")
	}
	if cp.Name != "alice" {
		t.Errorf("Name: got %q, want alice", cp.Name)
	}
	if cp.Port != 8080 {
		t.Errorf("Port: got %d, want 8080 (default)", cp.Port)
	}
}

func TestDecode_StructWithPointerFields(t *testing.T) {
	type Config struct {
		Name *string `json:"name"`
		Age  *int    `json:"age,omitempty"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	var c Config
	if err := schema.Decode(json.RawMessage(`{"name":"bob"}`), &c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Name == nil || *c.Name != "bob" {
		t.Errorf("Name: got %v, want pointer to \"bob\"", c.Name)
	}
	if c.Age != nil {
		t.Errorf("Age: got %v, want nil (omitted)", c.Age)
	}
}

///////////////////////////////////////////////////////////////////////////////
// TESTS FOR Validate

func TestValidate_ValidObject(t *testing.T) {
	type Config struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	data := json.RawMessage(`{"name":"alice","age":30}`)
	if err := schema.Validate(data); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidate_InvalidJSON(t *testing.T) {
	type Config struct {
		Name string `json:"name"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(json.RawMessage(`{not json}`)); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestValidate_WrongType(t *testing.T) {
	type Config struct {
		Count int `json:"count"`
	}
	schema, err := For[Config]()
	if err != nil {
		t.Fatal(err)
	}
	// Pass a string where an integer is expected.
	if err := schema.Validate(json.RawMessage(`{"count":"not-a-number"}`)); err == nil {
		t.Error("expected validation error for wrong type, got nil")
	}
}

func TestValidate_StringSchema(t *testing.T) {
	schema, err := For[string]()
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(json.RawMessage(`"hello"`)); err != nil {
		t.Errorf("expected no error for string, got: %v", err)
	}
	if err := schema.Validate(json.RawMessage(`42`)); err == nil {
		t.Error("expected error for number against string schema, got nil")
	}
	if err := schema.Validate(json.RawMessage(`true`)); err == nil {
		t.Error("expected error for bool against string schema, got nil")
	}
}

func TestValidate_IntSchema(t *testing.T) {
	schema, err := For[int]()
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(json.RawMessage(`42`)); err != nil {
		t.Errorf("expected no error for integer, got: %v", err)
	}
	if err := schema.Validate(json.RawMessage(`"hello"`)); err == nil {
		t.Error("expected error for string against integer schema, got nil")
	}
}

func TestValidate_BoolSchema(t *testing.T) {
	schema, err := For[bool]()
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(json.RawMessage(`true`)); err != nil {
		t.Errorf("expected no error for true, got: %v", err)
	}
	if err := schema.Validate(json.RawMessage(`false`)); err != nil {
		t.Errorf("expected no error for false, got: %v", err)
	}
	if err := schema.Validate(json.RawMessage(`"yes"`)); err == nil {
		t.Error("expected error for string against boolean schema, got nil")
	}
}

func TestValidate_ArraySchema(t *testing.T) {
	schema, err := For[[]string]()
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(json.RawMessage(`["a","b","c"]`)); err != nil {
		t.Errorf("expected no error for string array, got: %v", err)
	}
	if err := schema.Validate(json.RawMessage(`"not-an-array"`)); err == nil {
		t.Error("expected error for string against array schema, got nil")
	}
}

func TestValidate_InvalidJSONPrimitive(t *testing.T) {
	schema, err := For[string]()
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(json.RawMessage(`{not json}`)); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

///////////////////////////////////////////////////////////////////////////////
// DIRECT TESTS FOR parseEnumTag

func TestParseEnumTag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []any
	}{
		{"empty string", "", nil},
		{"single value", "foo", []any{"foo"}},
		{"multiple values", "a,b,c", []any{"a", "b", "c"}},
		{"leading/trailing spaces", " a , b , c ", []any{"a", "b", "c"}},
		{"whitespace-only value", "a, ,b", []any{"a", "b"}},
		{"all whitespace", "  ,  ,  ", nil},
		{"single whitespace-only", "   ", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseEnumTag(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("[%d]: got %v, want %v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestMustFor_Success(t *testing.T) {
	type S struct {
		Name string `json:"name"`
	}
	s := MustFor[S]()
	if s == nil {
		t.Fatal("MustFor returned nil for valid type")
	}
}

func TestMustFor_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustFor did not panic for invalid type")
		}
	}()
	// A channel cannot be represented as a JSON Schema; upstream returns an error.
	MustFor[chan int]()
}

///////////////////////////////////////////////////////////////////////////////
// UUID FIELD TESTS

// A string field tagged format:"uuid" produces type=string, format=uuid.
func TestFor_UUID_StringWithUUIDFormat(t *testing.T) {
	s, err := For[uuidStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["id"]
	if prop == nil {
		t.Fatal("expected property 'id'")
	}
	if prop.Type != "string" && !sliceContains(prop.Types, "string") {
		t.Errorf("id type: got Type=%q Types=%v, want string", prop.Type, prop.Types)
	}
	if prop.Format != "uuid" {
		t.Errorf("id format: got %q, want \"uuid\"", prop.Format)
	}
}

// A *string field tagged format:"uuid" still gets format=uuid applied.
func TestFor_UUID_PointerField(t *testing.T) {
	s, err := For[uuidPointerStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["id"]
	if prop == nil {
		t.Fatal("expected property 'id'")
	}
	if prop.Format != "uuid" {
		t.Errorf("id format: got %q, want \"uuid\"", prop.Format)
	}
	// Pointer field must allow null.
	hasNull := prop.Type == "null" || sliceContains(prop.Types, "null")
	if !hasNull {
		t.Errorf("expected null type for pointer field, got Type=%q Types=%v", prop.Type, prop.Types)
	}
}

// A UUID field with a default emits the default as a JSON string and is optional.
func TestFor_UUID_DefaultTag(t *testing.T) {
	s, err := For[uuidDefaultStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["id"]
	if prop == nil {
		t.Fatal("expected property 'id'")
	}
	if prop.Format != "uuid" {
		t.Errorf("id format: got %q, want \"uuid\"", prop.Format)
	}
	want := `"00000000-0000-0000-0000-000000000000"`
	if string(prop.Default) != want {
		t.Errorf("id default: got %s, want %s", prop.Default, want)
	}
	// Having a default makes the field optional.
	for _, r := range s.Required {
		if r == "id" {
			t.Error("'id' should not be in Required when it has a default")
		}
	}
}

// A UUID field with required:"" appears in the parent Required list.
func TestFor_UUID_RequiredTag(t *testing.T) {
	s, err := For[uuidRequiredStruct]()
	if err != nil {
		t.Fatal(err)
	}
	required := map[string]bool{}
	for _, r := range s.Required {
		required[r] = true
	}
	if !required["id"] {
		t.Error("expected 'id' in Required")
	}
}

// A UUID field with jsonschema:"…" gets the description applied.
func TestFor_UUID_DescriptionTag(t *testing.T) {
	s, err := For[uuidDescStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["id"]
	if prop == nil {
		t.Fatal("expected property 'id'")
	}
	if prop.Description != "resource identifier" {
		t.Errorf("id description: got %q, want \"resource identifier\"", prop.Description)
	}
}

// A UUID field with readonly:"" has ReadOnly=true in the schema.
func TestFor_UUID_ReadonlyTag(t *testing.T) {
	s, err := For[uuidReadonlyStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["id"]
	if prop == nil {
		t.Fatal("expected property 'id'")
	}
	if !prop.ReadOnly {
		t.Error("expected ReadOnly=true on id")
	}
	if s.Properties["name"].ReadOnly {
		t.Error("expected ReadOnly=false on name")
	}
}

// UUID format propagates into a nested struct field.
func TestFor_UUID_NestedStruct(t *testing.T) {
	s, err := For[nestedUUIDStruct]()
	if err != nil {
		t.Fatal(err)
	}
	inner := s.Properties["inner"]
	if inner == nil {
		t.Fatal("expected property 'inner'")
	}
	id := inner.Properties["id"]
	if id == nil {
		t.Fatal("expected nested property 'inner.id'")
	}
	if id.Format != "uuid" {
		t.Errorf("inner.id format: got %q, want \"uuid\"", id.Format)
	}
}

// Validate passes for a string UUID value and fails for a non-string.
func TestValidate_UUID_TypeCheck(t *testing.T) {
	s, err := For[uuidStruct]()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Validate(json.RawMessage(`{"id":"123e4567-e89b-12d3-a456-426614174000"}`)); err != nil {
		t.Errorf("expected no error for valid UUID string, got: %v", err)
	}
	// A non-string value must fail the type check.
	if err := s.Validate(json.RawMessage(`{"id":42}`)); err == nil {
		t.Error("expected error for integer UUID, got nil")
	}
}

// Decode populates the UUID string field correctly.
func TestDecode_UUID_Field(t *testing.T) {
	s, err := For[uuidStruct]()
	if err != nil {
		t.Fatal(err)
	}
	var v uuidStruct
	if err := s.Decode(json.RawMessage(`{"id":"123e4567-e89b-12d3-a456-426614174000"}`), &v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("ID: got %q, want \"123e4567-e89b-12d3-a456-426614174000\"", v.ID)
	}
}

// Decode applies the default UUID when the field is absent.
func TestDecode_UUID_Default(t *testing.T) {
	s, err := For[uuidDefaultStruct]()
	if err != nil {
		t.Fatal(err)
	}
	var v uuidDefaultStruct
	if err := s.Decode(json.RawMessage(`{}`), &v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("ID: got %q, want zero UUID", v.ID)
	}
}

///////////////////////////////////////////////////////////////////////////////
// EXAMPLE TAG TESTS

func TestFor_ExampleTag_String(t *testing.T) {
	s, err := For[exampleStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["name"]
	if prop == nil {
		t.Fatal("expected property 'name'")
	}
	if len(prop.Examples) != 1 {
		t.Fatalf("Examples: got %d, want 1", len(prop.Examples))
	}
	if prop.Examples[0] != "alice" {
		t.Errorf("Examples[0]: got %v, want \"alice\"", prop.Examples[0])
	}
}

func TestFor_ExampleTag_Int(t *testing.T) {
	s, err := For[exampleStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["age"]
	if prop == nil {
		t.Fatal("expected property 'age'")
	}
	if len(prop.Examples) != 1 {
		t.Fatalf("Examples: got %d, want 1", len(prop.Examples))
	}
	// JSON unmarshal decodes numbers as float64.
	if prop.Examples[0] != float64(30) {
		t.Errorf("Examples[0]: got %v, want 30", prop.Examples[0])
	}
}

func TestFor_ExampleTag_Float(t *testing.T) {
	s, err := For[exampleStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["score"]
	if prop == nil {
		t.Fatal("expected property 'score'")
	}
	if len(prop.Examples) != 1 {
		t.Fatalf("Examples: got %d, want 1", len(prop.Examples))
	}
	if prop.Examples[0] != float64(9.5) {
		t.Errorf("Examples[0]: got %v, want 9.5", prop.Examples[0])
	}
}

func TestFor_ExampleTag_Bool(t *testing.T) {
	s, err := For[exampleStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["flag"]
	if prop == nil {
		t.Fatal("expected property 'flag'")
	}
	if len(prop.Examples) != 1 {
		t.Fatalf("Examples: got %d, want 1", len(prop.Examples))
	}
	if prop.Examples[0] != true {
		t.Errorf("Examples[0]: got %v, want true", prop.Examples[0])
	}
}

func TestFor_ExampleTag_UUID(t *testing.T) {
	s, err := For[exampleStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["id"]
	if prop == nil {
		t.Fatal("expected property 'id'")
	}
	if len(prop.Examples) != 1 {
		t.Fatalf("Examples: got %d, want 1", len(prop.Examples))
	}
	if prop.Examples[0] != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("Examples[0]: got %v, want UUID string", prop.Examples[0])
	}
}

func TestFor_NoExampleTag_ExamplesNil(t *testing.T) {
	s, err := For[simpleStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["name"]
	if prop == nil {
		t.Fatal("expected property 'name'")
	}
	if prop.Examples != nil {
		t.Errorf("expected nil Examples for field without example tag, got %v", prop.Examples)
	}
}

///////////////////////////////////////////////////////////////////////////////
// uuid.UUID FIELD TESTS

// A uuid.UUID struct field produces type=string, format=uuid without any tag.
func TestFor_UUIDNative_StructField(t *testing.T) {
	s, err := For[uuidNativeStruct]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["id"]
	if prop == nil {
		t.Fatal("expected property 'id'")
	}
	if prop.Type != "string" && !sliceContains(prop.Types, "string") {
		t.Errorf("id type: got Type=%q Types=%v, want string", prop.Type, prop.Types)
	}
	if prop.Format != "uuid" {
		t.Errorf("id format: got %q, want \"uuid\"", prop.Format)
	}
}

// For[uuid.UUID] itself produces type=string, format=uuid.
func TestFor_UUIDNative_TopLevel(t *testing.T) {
	s, err := For[uuid.UUID]()
	if err != nil {
		t.Fatal(err)
	}
	if s.Type != "string" && !sliceContains(s.Types, "string") {
		t.Errorf("type: got Type=%q Types=%v, want string", s.Type, s.Types)
	}
	if s.Format != "uuid" {
		t.Errorf("format: got %q, want \"uuid\"", s.Format)
	}
}

// A uuid.UUID field without a format tag does not get format overwritten by
// an explicit format tag — the uuid detection runs first and then the tag wins.
func TestFor_UUIDNative_ExplicitFormatTagOverrides(t *testing.T) {
	type S struct {
		ID uuid.UUID `json:"id" format:"custom-id"`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["id"]
	if prop == nil {
		t.Fatal("expected property 'id'")
	}
	// The explicit format tag should win over the auto-detected uuid format.
	if prop.Format != "custom-id" {
		t.Errorf("id format: got %q, want \"custom-id\"", prop.Format)
	}
}

func TestFor_TimeTime_TopLevel(t *testing.T) {
	s, err := For[time.Time]()
	if err != nil {
		t.Fatal(err)
	}
	if s.Type != "string" {
		t.Errorf("type: got %q, want \"string\"", s.Type)
	}
	if s.Format != "date-time" {
		t.Errorf("format: got %q, want \"date-time\"", s.Format)
	}
}

func TestFor_TimeTime_StructField(t *testing.T) {
	type S struct {
		Created time.Time `json:"created"`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["created"]
	if prop == nil {
		t.Fatal("expected property 'created'")
	}
	if prop.Type != "string" {
		t.Errorf("created type: got %q, want \"string\"", prop.Type)
	}
	if prop.Format != "date-time" {
		t.Errorf("created format: got %q, want \"date-time\"", prop.Format)
	}
	if len(prop.Properties) != 0 {
		t.Errorf("created properties: got %d, want 0 (should not leak struct fields)", len(prop.Properties))
	}
}

func TestFor_URL_TopLevel(t *testing.T) {
	s, err := For[url.URL]()
	if err != nil {
		t.Fatal(err)
	}
	if s.Type != "string" {
		t.Errorf("type: got %q, want \"string\"", s.Type)
	}
	if s.Format != "uri" {
		t.Errorf("format: got %q, want \"uri\"", s.Format)
	}
}

func TestFor_URL_StructField(t *testing.T) {
	type S struct {
		Link *url.URL `json:"link"`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["link"]
	if prop == nil {
		t.Fatal("expected property 'link'")
	}
	if prop.Type != "string" {
		t.Errorf("link type: got %q, want \"string\"", prop.Type)
	}
	if prop.Format != "uri" {
		t.Errorf("link format: got %q, want \"uri\"", prop.Format)
	}
	if len(prop.Properties) != 0 {
		t.Errorf("link properties: got %d, want 0 (should not leak struct fields)", len(prop.Properties))
	}
}

func TestFor_ByteSlice_TopLevel(t *testing.T) {
	s, err := For[[]byte]()
	if err != nil {
		t.Fatal(err)
	}
	if s.Type != "string" {
		t.Errorf("type: got %q, want \"string\"", s.Type)
	}
	if s.Format != "byte" {
		t.Errorf("format: got %q, want \"byte\"", s.Format)
	}
}

func TestFor_ByteSlice_StructField(t *testing.T) {
	type S struct {
		Data []byte `json:"data"`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	prop := s.Properties["data"]
	if prop == nil {
		t.Fatal("expected property 'data'")
	}
	if prop.Type != "string" {
		t.Errorf("data type: got %q, want \"string\"", prop.Type)
	}
	if prop.Format != "byte" {
		t.Errorf("data format: got %q, want \"byte\"", prop.Format)
	}
}

func TestFor_SliceField_NoNull(t *testing.T) {
	type S struct {
		Tags []string `json:"tags"`
		IDs  []int    `json:"ids,omitempty"`
	}
	s, err := For[S]()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"tags", "ids"} {
		prop := s.Properties[name]
		if prop == nil {
			t.Fatalf("expected property %q", name)
		}
		// Type should be "array" with no "null" in Types.
		for _, typ := range prop.Types {
			if typ == "null" {
				t.Errorf("%s: Types contains \"null\", want no null for slice fields", name)
			}
		}
		if prop.Type == "null" {
			t.Errorf("%s: Type is \"null\", want \"array\"", name)
		}
	}
}
