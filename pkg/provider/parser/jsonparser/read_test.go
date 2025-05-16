package jsonparser_test

import (
	"fmt"
	"strings"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/provider/parser/jsonparser"
	"github.com/stretchr/testify/assert"
)

func Test_Read_001(t *testing.T) {
	assert := assert.New(t)
	test := `{"label":"main","bool":true,"null":null,"array":[1,2,3],"object":{"key":"value"},"number":12345678901234567890}`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_002(t *testing.T) {
	assert := assert.New(t)
	test := `{"log":{"debug":true}}`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_003(t *testing.T) {
	assert := assert.New(t)
	test := `true`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_004(t *testing.T) {
	assert := assert.New(t)
	test := `null`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_005(t *testing.T) {
	assert := assert.New(t)
	test := `"hello, world"`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_006(t *testing.T) {
	assert := assert.New(t)
	test := `156`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_007(t *testing.T) {
	assert := assert.New(t)
	test := `[1,2,3]`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_008(t *testing.T) {
	assert := assert.New(t)
	test := `{"A":1,"B":2,"C":3}`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_009(t *testing.T) {
	assert := assert.New(t)
	test := `{"A":[1,2,3],"B":[1,2,3],"C":[1,2,3]}`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_010(t *testing.T) {
	assert := assert.New(t)
	test := `[{"A":[1,2,3],"B":[1,2,3]},{"C":[1,2,3],"D":[1,2,3]}]`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_011(t *testing.T) {
	assert := assert.New(t)
	test := `[]`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}

func Test_Read_012(t *testing.T) {
	assert := assert.New(t)
	test := `{}`
	tree, err := jsonparser.Read(strings.NewReader(test))
	assert.NoError(err)
	assert.Equal(test, fmt.Sprint(tree))
}
