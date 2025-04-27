package metrics_test

import (
	"bytes"
	"testing"

	// Packages
	metrics "github.com/mutablelogic/go-server/pkg/metrics"
	schema "github.com/mutablelogic/go-server/pkg/metrics/schema"
	assert "github.com/stretchr/testify/assert"
)

func Test_Manager_00(t *testing.T) {
	assert := assert.New(t)

	manager := metrics.New()
	assert.NotNil(manager)

	t.Run("00", func(t *testing.T) {
		// Float Value
		family, err := manager.RegisterFamily("test", schema.WithHelp(t.Name()), schema.WithType("gauge"), schema.WithUnit("seconds"), schema.WithCounter(100), schema.WithGauge(200))
		if assert.NoError(err) {
			assert.NotNil(family)
		}
	})

	var buf bytes.Buffer
	assert.NoError(manager.Write(&buf))
	t.Log(buf.String())
}
