package parser_test

import (
	"testing"

	// Packages
	parser "github.com/mutablelogic/go-server/pkg/parser"
	log "github.com/mutablelogic/go-server/plugin/log/config"
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
	assert.NoError(err)

	err = parser.Parse("testdata/log.json")
	assert.NoError(err)
}
