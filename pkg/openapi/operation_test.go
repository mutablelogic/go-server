package openapi

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/jsonschema"
	"github.com/mutablelogic/go-server/pkg/types"
)

type testCreateRequest struct {
	Name string `json:"name" required:""`
}

type testRefreshRequest struct {
	Token string `json:"token" required:""`
}

func TestWithJSONRequestOneOf(t *testing.T) {
	operation := Operation("test",
		WithNamedJSONRequest(
			"TestRequest",
			NamedSchema("CreateRequest", jsonschema.MustFor[testCreateRequest]()),
			NamedSchema("RefreshRequest", jsonschema.MustFor[testRefreshRequest]()),
		),
	)

	if operation.RequestBody == nil {
		t.Fatal("expected request body")
	}
	mediaType, ok := operation.RequestBody.Content[types.ContentTypeJSON]
	if !ok {
		t.Fatal("expected json request media type")
	}
	if got := len(mediaType.Schema.OneOf); got != 2 {
		t.Fatalf("expected 2 oneOf schemas, got %d", got)
	}
	if mediaType.Schema.Title != "TestRequest" {
		t.Fatalf("expected wrapper title TestRequest, got %q", mediaType.Schema.Title)
	}
	if mediaType.Schema.OneOf[0].Title != "CreateRequest" {
		t.Fatalf("expected first oneOf title CreateRequest, got %q", mediaType.Schema.OneOf[0].Title)
	}
	if mediaType.Schema.OneOf[1].Title != "RefreshRequest" {
		t.Fatalf("expected second oneOf title RefreshRequest, got %q", mediaType.Schema.OneOf[1].Title)
	}
	if mediaType.Schema.OneOf[0].Properties["name"] == nil {
		t.Fatal("expected first oneOf schema to include name property")
	}
	if mediaType.Schema.OneOf[1].Properties["token"] == nil {
		t.Fatal("expected second oneOf schema to include token property")
	}
}

func TestWithJSONRequestSingleSchema(t *testing.T) {
	request := jsonschema.MustFor[testCreateRequest]()
	operation := Operation("test", WithJSONRequest(request))

	if operation.RequestBody == nil {
		t.Fatal("expected request body")
	}
	mediaType := operation.RequestBody.Content[types.ContentTypeJSON]
	if mediaType.Schema != request {
		t.Fatal("expected single schema to be passed through unchanged")
	}
}

func TestNamedSchemaDoesNotMutateInput(t *testing.T) {
	request := jsonschema.MustFor[testCreateRequest]()
	named := NamedSchema("CreateRequest", request)

	if request.Title != "" {
		t.Fatalf("expected original schema title to remain empty, got %q", request.Title)
	}
	if named.Title != "CreateRequest" {
		t.Fatalf("expected cloned schema title CreateRequest, got %q", named.Title)
	}
	if named == request {
		t.Fatal("expected NamedSchema to return a clone")
	}
}
