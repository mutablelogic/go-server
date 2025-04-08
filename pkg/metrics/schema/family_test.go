package schema_test

import (
	"bytes"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-server/pkg/metrics/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_Family_00(t *testing.T) {
	assert := assert.New(t)

	t.Run("unknown", func(t *testing.T) {
		family, err := schema.NewMetricFamily("test")
		if assert.NoError(err) {
			assert.NotNil(family)
		}

		var buf bytes.Buffer
		if assert.NoError(family.Write(&buf)) {
			assert.Equal("# TYPE test unknown\n", buf.String())
		}
	})

	t.Run("gauge", func(t *testing.T) {
		family, err := schema.NewMetricFamily("test", schema.WithType("gauge"), schema.WithHelp("help"))
		if assert.NoError(err) {
			assert.NotNil(family)
		}

		var buf bytes.Buffer
		if assert.NoError(family.Write(&buf)) {
			assert.Equal("# TYPE test gauge\n# HELP test help\n", buf.String())
		}
	})

	t.Run("seconds", func(t *testing.T) {
		family, err := schema.NewMetricFamily("test", schema.WithType("gauge"), schema.WithHelp("help"), schema.WithUnit("seconds"))
		if assert.NoError(err) {
			assert.NotNil(family)
		}

		var buf bytes.Buffer
		if assert.NoError(family.Write(&buf)) {
			assert.Equal("# TYPE test_seconds gauge\n# UNIT test_seconds seconds\n# HELP test_seconds help\n", buf.String())
		}
	})
}

func Test_Family_01(t *testing.T) {
	assert := assert.New(t)

	t.Run("unknown", func(t *testing.T) {
		family, err := schema.NewMetricFamily("test", schema.WithFloat("test", 13.14))
		if assert.NoError(err) {
			assert.NotNil(family)
		}

		var buf bytes.Buffer
		if assert.NoError(family.Write(&buf)) {
			assert.Equal("# TYPE test unknown\ntest_test 13.14\n", buf.String())
		}
	})
}

func Test_Family_02(t *testing.T) {
	assert := assert.New(t)

	t.Run("unknown", func(t *testing.T) {
		family, err := schema.NewMetricFamily("test",
			schema.WithInt("testa", 13),
			schema.WithFloat("testb", 1),
		)
		if assert.NoError(err) {
			assert.NotNil(family)
			var buf bytes.Buffer
			if assert.NoError(family.Write(&buf)) {
				assert.Equal("# TYPE test unknown\ntest_testa 13\ntest_testb 1\n", buf.String())
			}
		}
	})
}
