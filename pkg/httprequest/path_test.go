package httprequest

import (
	"testing"

	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
)

func TestParametersFromPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{name: "no parameters", path: "/resource", want: nil},
		{name: "single parameter", path: "/resource/{id}", want: []string{"id"}},
		{name: "multiple parameters", path: "/resource/{id}/child/{name}", want: []string{"id", "name"}},
		{name: "duplicate parameters ignored", path: "/resource/{id}/child/{id}", want: []string{"id"}},
		{name: "invalid segments ignored", path: "/resource/{id}/child/{}/literal{bad}", want: []string{"id"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := parametersFromPath(tt.path)
			if len(params) != len(tt.want) {
				t.Fatalf("len(parametersFromPath(%q)) = %d, want %d", tt.path, len(params), len(tt.want))
			}
			for i, want := range tt.want {
				if params[i].Name != want {
					t.Fatalf("parametersFromPath(%q)[%d].Name = %q, want %q", tt.path, i, params[i].Name, want)
				}
				if params[i].In != openapi.ParameterInPath {
					t.Fatalf("parametersFromPath(%q)[%d].In = %q, want %q", tt.path, i, params[i].In, openapi.ParameterInPath)
				}
				if !params[i].Required {
					t.Fatalf("parametersFromPath(%q)[%d].Required = false, want true", tt.path, i)
				}
				if params[i].Schema == nil {
					t.Fatalf("parametersFromPath(%q)[%d].Schema = nil, want string schema", tt.path, i)
				}
			}
		})
	}
}

func TestNewPathParameters(t *testing.T) {
	p := NewPath("resource/{id}/child/{name}", "summary")
	params := append([]openapi.Parameter(nil), p.parameters...)

	if len(params) != 2 {
		t.Fatalf("len(p.parameters) = %d, want 2", len(params))
	}
	if params[0].Name != "id" || params[1].Name != "name" {
		t.Fatalf("p.parameters names = [%q %q], want [\"id\" \"name\"]", params[0].Name, params[1].Name)
	}

	params[0].Name = "mutated"
	if p.parameters[0].Name != "id" {
		t.Fatalf("test copy should not mutate p.parameters")
	}
}
