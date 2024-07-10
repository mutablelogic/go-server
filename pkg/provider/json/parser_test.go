package json_test

import (
	"os"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/provider/json"
	"github.com/stretchr/testify/assert"
)

func Test_parser_001(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../../etc/json/nginx-proxy.json")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	tree, err := json.Parse(r)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(tree)
	t.Log(tree)
}

func Test_parser_002(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../../etc/json/parser-test-002.json")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	tree, err := json.Parse(r)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(tree)
	t.Log(tree)
}

func Test_parser_003(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../../etc/json/parser-test-003.json")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	tree, err := json.Parse(r)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(tree)
	t.Log(tree)
}

func Test_parser_004(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../../etc/json/parser-test-004.json")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	tree, err := json.Parse(r)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	assert.NotNil(tree)
	t.Log(tree)
}
