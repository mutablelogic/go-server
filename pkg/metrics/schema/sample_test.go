package schema_test

import (
	"bytes"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/metrics/schema"
	"github.com/stretchr/testify/assert"
)

func Test_Sample_00(t *testing.T) {
	assert := assert.New(t)

	t.Run("00", func(t *testing.T) {
		sample, err := schema.NewFloat("test", 13.14)
		if assert.NoError(err) {
			assert.NotNil(sample)
		}

		var buf bytes.Buffer
		if assert.NoError(sample.Write(&buf, "family")) {
			assert.Equal("family_test 13.14\n", buf.String())
		}
	})

	t.Run("01", func(t *testing.T) {
		sample, err := schema.NewInt("test", 13)
		if assert.NoError(err) {
			assert.NotNil(sample)
		}

		var buf bytes.Buffer
		if assert.NoError(sample.Write(&buf, "family")) {
			assert.Equal("family_test 13\n", buf.String())
		}
	})

	t.Run("02", func(t *testing.T) {
		sample, err := schema.NewInt("test", 0)
		if assert.NoError(err) {
			assert.NotNil(sample)
		}

		var buf bytes.Buffer
		if assert.NoError(sample.Write(&buf, "family")) {
			assert.Equal("family_test 0\n", buf.String())
		}
	})

	t.Run("03", func(t *testing.T) {
		sample, err := schema.NewFloat("test", 0)
		if assert.NoError(err) {
			assert.NotNil(sample)
		}

		var buf bytes.Buffer
		if assert.NoError(sample.Write(&buf, "family")) {
			assert.Equal("family_test 0\n", buf.String())
		}
	})
}

func Test_Sample_01(t *testing.T) {
	assert := assert.New(t)

	t.Run("00", func(t *testing.T) {
		sample, err := schema.NewFloat("test", 13.14, "a", "b")
		if assert.NoError(err) {
			assert.NotNil(sample)
		}

		var buf bytes.Buffer
		if assert.NoError(sample.Write(&buf, "family")) {
			assert.Equal("family_test{a=\"b\"} 13.14\n", buf.String())
		}
	})

	t.Run("01", func(t *testing.T) {
		sample, err := schema.NewInt("test", 13, "a", "b", "c", "d")
		if assert.NoError(err) {
			assert.NotNil(sample)
		}

		var buf bytes.Buffer
		if assert.NoError(sample.Write(&buf, "family")) {
			assert.Equal("family_test{a=\"b\",c=\"d\"} 13\n", buf.String())
		}
	})
}
