package types_test

import (
	"fmt"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_UUID_001(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		Expected bool
		Input    string
	}{
		{false, ""},
		{false, "123"},
		{true, "123e4567-e89b-12d3-a456-426614174000"},
		{true, "123E4567-E89B-12D3-A456-426614174000"},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert.Equal(test.Expected, types.IsUUID(test.Input))
		})
	}
}

func Test_UUID_002(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		Expected []string
		Input    string
	}{
		{nil, ""},
		{nil, "123"},
		{[]string{"123e", "4567", "e89b", "12d3", "a456", "4266", "1417", "4000"}, "123e4567-e89b-12d3-a456-426614174000"},
		{[]string{"123e", "4567", "e89b", "12d3", "a456", "4266", "1417", "4000"}, "123E4567-E89B-12D3-A456-426614174000"},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			assert.Equal(test.Expected, types.UUIDSplit(test.Input))
		})
	}
}
