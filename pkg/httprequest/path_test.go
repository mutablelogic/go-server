package httprequest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi "github.com/mutablelogic/go-server/pkg/openapi/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
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
		{name: "trailing wildcard", path: "/files/{path...}", want: []string{"path"}},
		{name: "parameter and trailing wildcard", path: "/object/{volume}/{key...}", want: []string{"volume", "key"}},
		{name: "end of path marker ignored", path: "/resource/{$}", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := parametersFromPath(tt.path, nil)
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
	p := NewPathItem("summary", "description")
	spec := p.Spec("resource/{id}/child/{name}", nil)
	if len(spec.Parameters) != 2 {
		t.Fatalf("len(spec.Parameters) = %d, want 2", len(spec.Parameters))
	}
	if spec.Parameters[0].Name != "id" || spec.Parameters[1].Name != "name" {
		t.Fatalf("spec.Parameters names = [%q %q], want [\"id\" \"name\"]", spec.Parameters[0].Name, spec.Parameters[1].Name)
	}
}

func TestSpecReturnsPathAndOperations(t *testing.T) {
	p := NewPathItem("summary", "description")
	p.Get(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) { op.Summary("get resource") })

	spec := p.Spec("resource/{id}", nil)
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
	p := NewPathItem("summary", "description")
	called := false

	p.Get(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}, func(op PathOperation) { op.Summary("get resource") })

	spec := p.Spec("resource/{id}", nil)
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
	p := NewPathItem("summary", "description")

	req := httptest.NewRequest(http.MethodGet, "/resource/123", nil)
	res := httptest.NewRecorder()
	p.Handler()(res, req)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("response status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
	}
}

func TestParametersFromPathWildcardSchema(t *testing.T) {
	// The schema property for a trailing wildcard is found by its name,
	// without the "..." suffix
	type params struct {
		Volume string `json:"volume" help:"Volume name"`
		Key    string `json:"key" help:"Object key"`
	}
	params_ := parametersFromPath("/object/{volume}/{key...}", jsonschema.MustFor[params]())
	if len(params_) != 2 {
		t.Fatalf("len(params) = %d, want 2", len(params_))
	}
	for i, want := range []string{"Volume name", "Object key"} {
		if params_[i].Schema == nil || params_[i].Schema.Description != want {
			t.Fatalf("params[%d] (%q) schema description = %v, want %q", i, params_[i].Name, params_[i].Schema, want)
		}
	}
}

func TestSecurityRequiresAllSchemes(t *testing.T) {
	// Chained schemes are all required, so they are documented as a single
	// requirement rather than as alternatives
	scopes := []string{"read"}
	p := NewPathItem("summary", "description")
	p.Get(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) {
		op.Security("apiKey").Security("oauth", scopes...)
	})
	scopes[0] = "changed"

	spec := p.Spec("resource", nil)
	if spec == nil || spec.Get == nil {
		t.Fatalf("Spec().Get = nil, want populated GET operation")
	}
	if len(spec.Get.Security) != 1 {
		t.Fatalf("len(Security) = %d, want 1 requirement holding both schemes", len(spec.Get.Security))
	}
	requirement := spec.Get.Security[0]
	if got, ok := requirement["apiKey"]; !ok || got == nil || len(got) != 0 {
		t.Errorf("apiKey scopes = %#v, want empty, non-nil", got)
	}
	if got := requirement["oauth"]; len(got) != 1 || got[0] != "read" {
		t.Errorf("oauth scopes = %#v, want [read], copied from the caller", got)
	}
}

func TestResponseContentTypes(t *testing.T) {
	// A status code can have more than one content type, such as a JSON
	// response, or a stream of server-sent events
	p := NewPathItem("summary", "description")
	p.Post(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) {
		op.JSONResponse(http.StatusOK, jsonschema.MustFor[string](), "Model, or a stream of events")
		op.TextStreamResponse(http.StatusOK, "ignored, as the response exists")
		op.Response(http.StatusOK, "text/plain")
		op.ErrorResponse(http.StatusBadRequest)
	})

	spec := p.Spec("resource", nil)
	if spec == nil || spec.Post == nil {
		t.Fatalf("Spec().Post = nil, want populated POST operation")
	}
	response, ok := spec.Post.Responses["200"]
	if !ok {
		t.Fatalf("Responses[200] missing")
	}
	if response.Description != "Model, or a stream of events" {
		t.Errorf("Description = %q, want the first description", response.Description)
	}
	for _, contentType := range []string{"application/json", "text/event-stream", "text/plain"} {
		if _, ok := response.Content[contentType]; !ok {
			t.Errorf("Content[%q] missing, have %v", contentType, response.Content)
		}
	}
	if stream := response.Content["text/event-stream"]; stream.Schema == nil || stream.Schema.Type != "string" {
		t.Errorf("text/event-stream schema = %#v, want a string schema", stream.Schema)
	}

	// Another status code is a separate response, with its own description
	if response := spec.Post.Responses["400"]; response.Description != http.StatusText(http.StatusBadRequest) || len(response.Content) != 1 {
		t.Errorf("Responses[400] = %#v, want one content type, with the default description", response)
	}
}

func TestRequestBodyContentTypes(t *testing.T) {
	// A request body can have more than one content type, each with its own
	// schema, such as JSON, or a multipart form with a file
	type create struct {
		Name string `json:"name"`
	}
	type form struct {
		Name string     `json:"name"`
		File types.File `json:"file"`
	}
	p := NewPathItem("summary", "description")
	p.Post(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) {
		op.RequestBody(jsonschema.MustFor[create]())
		op.RequestBody(jsonschema.MustFor[form](), types.ContentTypeFormData, types.ContentTypeForm)
	})

	spec := p.Spec("resource", nil)
	if spec == nil || spec.Post == nil || spec.Post.RequestBody == nil {
		t.Fatalf("Spec().Post.RequestBody = nil, want populated request body")
	}
	content := spec.Post.RequestBody.Content
	if len(content) != 3 {
		t.Fatalf("Content = %v, want three content types", content)
	}
	if schema := content[types.ContentTypeJSON].Schema; schema == nil || schema.Properties["file"] != nil {
		t.Errorf("application/json schema = %#v, want the create schema", schema)
	}
	for _, contentType := range []string{types.ContentTypeFormData, types.ContentTypeForm} {
		if schema := content[contentType].Schema; schema == nil || schema.Properties["file"] == nil {
			t.Errorf("%s schema = %#v, want the form schema", contentType, schema)
		}
	}
}

func TestRequestBodyDefaultContentType(t *testing.T) {
	p := NewPathItem("summary", "description")
	p.Post(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) {
		op.RequestBody(jsonschema.MustFor[string]())
	})
	spec := p.Spec("resource", nil)
	if spec == nil || spec.Post == nil || spec.Post.RequestBody == nil {
		t.Fatalf("Spec().Post.RequestBody = nil, want populated request body")
	}
	if content := spec.Post.RequestBody.Content; len(content) != 1 || content[types.ContentTypeJSON].Schema == nil {
		t.Errorf("Content = %v, want application/json only", content)
	}
}

func TestTextStreamResponseOnly(t *testing.T) {
	p := NewPathItem("summary", "description")
	p.Get(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) {
		op.TextStreamResponse(http.StatusOK)
	})
	response := p.Spec("resource", nil).Get.Responses["200"]
	if response.Description != http.StatusText(http.StatusOK) {
		t.Errorf("Description = %q, want the default", response.Description)
	}
	if stream, ok := response.Content["text/event-stream"]; !ok || len(response.Content) != 1 {
		t.Errorf("Content = %v, want text/event-stream only", response.Content)
	} else if stream.Schema == nil || stream.Schema.Type != "string" {
		t.Errorf("text/event-stream schema = %#v, want a string schema", stream.Schema)
	}
}

func TestEmptyResponse(t *testing.T) {
	// A 204 No Content response, without content, alongside a JSON response
	p := NewPathItem("summary", "description")
	p.Post(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) {
		op.JSONResponse(http.StatusOK, jsonschema.MustFor[string](), "Result")
		op.EmptyResponse("No result")
	})
	responses := p.Spec("resource", nil).Post.Responses

	response, ok := responses["204"]
	if !ok {
		t.Fatalf("Responses[204] missing")
	}
	if response.Description != "No result" {
		t.Errorf("Description = %q, want %q", response.Description, "No result")
	}
	if len(response.Content) != 0 {
		t.Errorf("Content = %v, want none", response.Content)
	}

	// The JSON response is unchanged
	if response := responses["200"]; response.Description != "Result" || len(response.Content) != 1 {
		t.Errorf("Responses[200] = %#v, want one content type", response)
	}

	// Without a description, the default HTTP status text is used
	p = NewPathItem("summary", "description")
	p.Delete(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) {
		op.EmptyResponse()
	})
	if response := p.Spec("resource", nil).Delete.Responses["204"]; response.Description != http.StatusText(http.StatusNoContent) || len(response.Content) != 0 {
		t.Errorf("Responses[204] = %#v, want no content, with the default description", response)
	}

	// Content added for a 204 response is removed, and the response keeps
	// its description
	p = NewPathItem("summary", "description")
	p.Put(func(http.ResponseWriter, *http.Request) {}, func(op PathOperation) {
		op.JSONResponse(http.StatusNoContent, jsonschema.MustFor[string](), "Nothing")
		op.EmptyResponse("ignored, as the response exists")
	})
	if response := p.Spec("resource", nil).Put.Responses["204"]; response.Description != "Nothing" || len(response.Content) != 0 {
		t.Errorf("Responses[204] = %#v, want no content, with the first description", response)
	}
}
