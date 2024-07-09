package provider_test

import (
	"os"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_jsonparser_001(t *testing.T) {
	assert := assert.New(t)

	parser, err := provider.NewJSONParser()
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(parser)

	r, err := os.Open("../../etc/json/nginx-proxy.json")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	tree, err := parser.Read(r)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(tree)
	t.Log(tree)
}

func Test_jsonparser_002(t *testing.T) {
	assert := assert.New(t)

	parser, err := provider.NewJSONParser()
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(parser)

	r, err := os.Open("../../etc/json/parser-tests.json")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	tree, err := parser.Read(r)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(tree)
	t.Log(tree)
}
