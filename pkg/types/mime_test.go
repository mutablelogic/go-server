package types_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-service/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_Mime_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("1", func(t *testing.T) {
		_, err := types.ParseContentType("")
		if assert.Error(err) {
			assert.ErrorContains(err, "no media type")
		}
	})

	t.Run("2", func(t *testing.T) {
		mime, err := types.ParseContentType(types.ContentTypeJSON)
		if assert.NoError(err) {
			assert.Equal(mime, types.ContentTypeJSON)
		}
	})
}
