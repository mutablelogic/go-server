package context_test

import (
	"net/http"
	"testing"

	// Packages
	context "github.com/mutablelogic/go-server/pkg/context"
	"github.com/mutablelogic/go-server/pkg/httpserver/util"
	"github.com/stretchr/testify/assert"
)

func Test_Context_001(t *testing.T) {
	ctx, cancel := context.WithCancel()
	defer cancel()
	if ctx == nil || cancel == nil {
		t.Fatal("Expected non-nil")
	}
}

func Test_Context_002(t *testing.T) {
	ctx, cancel := context.WithCancel()
	defer cancel()

	ctx = context.WithNameLabel(ctx, "name", "label")
	if v := context.Name(ctx); v != "name" {
		t.Error("Unexpected name", v)
	}
	if v := context.Label(ctx); v != "label" {
		t.Error("Unexpected label", v)
	}
}

func Test_Context_003(t *testing.T) {
	ctx, cancel := context.WithCancel()
	defer cancel()

	ctx = context.WithAddress(ctx, "address")
	if v := context.Address(ctx); v != "address" {
		t.Error("Unexpected address", v)
	}
}

func Test_Context_004(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithCancel()
	defer cancel()

	req, err := http.NewRequestWithContext(context.WithPrefixPathParams(ctx, "prefix", "path", nil), "GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	prefix, path, params := util.ReqPrefixPathParams(req)
	assert.Equal("prefix", prefix)
	assert.Equal("path", path)
	assert.Nil(params)
}

func Test_Context_005(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithCancel()
	defer cancel()

	params := []string{"a", "b", "c"}
	req, err := http.NewRequestWithContext(context.WithPrefixPathParams(ctx, "prefix", "path", params), "GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	prefix, path, params2 := util.ReqPrefixPathParams(req)
	assert.Equal("prefix", prefix)
	assert.Equal("path", path)
	assert.Equal(params, params2)
}

func Test_Context_006(t *testing.T) {
	ctx, cancel := context.WithCancel()
	defer cancel()

	ctx = context.WithAdmin(ctx, true)
	if v := context.Admin(ctx); !v {
		t.Error("Unexpected admin", v)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v := util.ReqAdmin(req); !v {
		t.Error("Unexpected admin", v)
	}
}
