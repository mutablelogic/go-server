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

func TestNewPathItemParameters(t *testing.T) {
	p := NewPathItem("resource/{id}/child/{name}", "summary")
	if len(p.parameters) != 2 {
		t.Fatalf("len(p.parameters) = %d, want 2", len(p.parameters))
	}
	if p.parameters[0].Name != "id" || p.parameters[1].Name != "name" {
		t.Fatalf("p.parameters names = [%q %q], want [\"id\" \"name\"]", p.parameters[0].Name, p.parameters[1].Name)
	}
}

func TestSpecReturnsPathAndOperations(t *testing.T) {
	p := NewPathItem("resource/{id}", "summary")
	p.Get(func(http.ResponseWriter, *http.Request) {}, "get resource")

	path, spec := p.Spec()
	if path != "resource/{id}" {
		t.Fatalf("Spec() path = %q, want %q", path, "resource/{id}")
	}
	if spec == nil || spec.Get == nil {
		t.Fatalf("Spec().Get = nil, want populated GET operation")
	}
	if len(spec.Parameters) != 1 || spec.Parameters[0].Name != "id" {
		t.Fatalf("Spec().Parameters = %#v, want single path parameter id", spec.Parameters)
	}
	if spec.Get.Summary != "get resource" {
		t.Fatalf("Spec().Get.Summary = %q, want %q", spec.Get.Summary, "get resource")
	}
}

func TestHandlerDispatchesMethod(t *testing.T) {
	p := NewPathItem("resource/{id}", "summary")
	called := false

	p.Get(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}, "get resource")

	path, spec := p.Spec()
	if path != "resource/{id}" {
		t.Fatalf("Spec path = %q, want %q", path, "resource/{id}")
	}
	if spec == nil || spec.Get == nil {
		t.Fatalf("Spec().Get = nil, want populated GET operation")
	}
	if len(spec.Parameters) != 1 || spec.Parameters[0].Name != "id" {
		t.Fatalf("Spec().Parameters = %#v, want single path parameter id", spec.Parameters)
	}

	req := httptest.NewRequest(http.MethodGet, "/resource/123", nil)
	res := httptest.NewRecorder()
	p.Handler()(res, req)

	if !called {
		t.Fatalf("registered GET handler was not called")
	}
	if res.Code != http.StatusNoContent {
		t.Fatalf("response status = %d, want %d", res.Code, http.StatusNoContent)
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	p := NewPathItem("resource/{id}", "summary")

	req := httptest.NewRequest(http.MethodGet, "/resource/123", nil)
	res := httptest.NewRecorder()
	p.Handler()(res, req)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("response status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
	}
}
