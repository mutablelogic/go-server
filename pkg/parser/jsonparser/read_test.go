package jsonparser_test

import (
	"strings"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/parser/jsonparser"
	"github.com/stretchr/testify/assert"
)

func Test_Read_001(t *testing.T) {
	assert := assert.New(t)
	test := `{ "label": "main", "bool": true, "null": null, "array": [1, 2, 3], "object": { "key": "value" }, "number": 12345678901234567890 }`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	t.Log(tree)
}
