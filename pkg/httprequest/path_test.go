package httprequest

import (
	"net/http"
	"net/http/httptest"
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
	if len(p.parameters) != 2 {
		t.Fatalf("len(p.parameters) = %d, want 2", len(p.parameters))
	}
	if p.parameters[0].Name != "id" || p.parameters[1].Name != "name" {
		t.Fatalf("p.parameters names = [%q %q], want [\"id\" \"name\"]", p.parameters[0].Name, p.parameters[1].Name)
	}
}

func TestSpecReturnsDefensiveCopy(t *testing.T) {
	p := NewPath("resource/{id}", "summary")
	err := p.Register(http.MethodGet, func(http.ResponseWriter, *http.Request) {}, "get resource")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	_, spec := p.Spec()
	if spec == nil || spec.Get == nil || len(spec.Get.Parameters) != 1 {
		t.Fatalf("Spec().Get.Parameters = %#v, want single parameter", spec)
	}
	spec.Get.Parameters[0].Name = "mutated"
	spec.Get.Summary = "mutated summary"

	if p.parameters[0].Name != "id" {
		t.Fatalf("mutating returned spec should not mutate p.parameters")
	}

	_, fresh := p.Spec()
	if fresh == nil || fresh.Get == nil {
		t.Fatalf("fresh Spec().Get = nil, want populated GET operation")
	}
	if fresh.Get.Parameters[0].Name != "id" {
		t.Fatalf("fresh Spec().Get.Parameters[0].Name = %q, want %q", fresh.Get.Parameters[0].Name, "id")
	}
	if fresh.Get.Summary != "get resource" {
		t.Fatalf("fresh Spec().Get.Summary = %q, want %q", fresh.Get.Summary, "get resource")
	}
}

func TestRegisterNormalizesMethod(t *testing.T) {
	p := NewPath("resource/{id}", "summary")
	called := false

	err := p.Register("get", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}, "get resource")
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	path, spec := p.Spec()
	if path != "resource/{id}" {
		t.Fatalf("Spec path = %q, want %q", path, "resource/{id}")
	}
	if spec == nil || spec.Get == nil {
		t.Fatalf("Spec().Get = nil, want populated GET operation")
	}
	if len(spec.Get.Parameters) != 1 || spec.Get.Parameters[0].Name != "id" {
		t.Fatalf("Spec().Get.Parameters = %#v, want single path parameter id", spec.Get.Parameters)
	}

	req := httptest.NewRequest(http.MethodGet, "/resource/123", nil)
	res := httptest.NewRecorder()
	p.Handler()(res, req)

	if !called {
		t.Fatalf("registered lowercase method handler was not called")
	}
	if res.Code != http.StatusNoContent {
		t.Fatalf("response status = %d, want %d", res.Code, http.StatusNoContent)
	}
}

func TestRegisterRejectsUnsupportedMethodWithoutMutation(t *testing.T) {
	p := NewPath("resource/{id}", "summary")
	err := p.Register("brew", func(http.ResponseWriter, *http.Request) {}, "brew resource")
	if err == nil {
		t.Fatalf("Register returned nil error for unsupported method")
	}
	if len(p.handlers) != 0 {
		t.Fatalf("len(p.handlers) = %d, want 0", len(p.handlers))
	}
	if p.spec.Get != nil || p.spec.Post != nil || p.spec.Put != nil || p.spec.Patch != nil || p.spec.Delete != nil || p.spec.Head != nil || p.spec.Options != nil || p.spec.Trace != nil {
		t.Fatalf("spec operations were mutated for unsupported method")
	}

	req := httptest.NewRequest(http.MethodGet, "/resource/123", nil)
	res := httptest.NewRecorder()
	p.Handler()(res, req)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("response status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
	}
}
