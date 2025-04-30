package parser_test

import (
	"testing"

	// Packages
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter/config"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver/config"
	log "github.com/mutablelogic/go-server/pkg/logger/config"
	parser "github.com/mutablelogic/go-server/pkg/parser"
	assert "github.com/stretchr/testify/assert"
)

func Test_Parser_001(t *testing.T) {
	assert := assert.New(t)

	// Create a new parser
	parser, err := parser.New(log.Config{})
	assert.NoError(err)
	assert.NotNil(parser)
}

func Test_Parser_002(t *testing.T) {
	assert := assert.New(t)

	// Create a new parser
	parser, err := parser.New(log.Config{})
	if !assert.NoError(err) {
		t.FailNow()
	}

	err = parser.Parse("testdata/log.json")
	assert.NoError(err)
}

func Test_Parser_003(t *testing.T) {
	assert := assert.New(t)

	// Create a new parser
	parser, err := parser.New(httpserver.Config{}, httprouter.Config{})
	if !assert.NoError(err) {
		t.FailNow()
	}

	err = parser.Parse("testdata/httpserver.json")
	assert.NoError(err)
}
