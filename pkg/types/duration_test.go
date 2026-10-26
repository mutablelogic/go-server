package types_test

import (
	json "encoding/json/v2"
	"testing"
	"time"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_ParseDuration(t *testing.T) {
	for value, want := range map[string]time.Duration{
		"5s":                  5 * time.Second,
		"1h30m":               90 * time.Minute,
		"5000000000":          5 * time.Second,
		"0":                   0,
		"-1s":                 -time.Second,
		"9223372036854775807": time.Duration(1<<63 - 1),
		"7d":                  7 * 24 * time.Hour,
		"1w":                  7 * 24 * time.Hour,
		"1w2d":                9 * 24 * time.Hour,
		"1d12h":               36 * time.Hour,
		"1.5d":                36 * time.Hour,
		"-2d":                 -48 * time.Hour,
		"1d30m15s":            24*time.Hour + 30*time.Minute + 15*time.Second,
	} {
		got, err := types.ParseDuration(value)
		require.NoError(t, err, value)
		assert.Equal(t, want, got, value)
	}
	for _, value := range []string{"", "soon", "1.5", "5 s", "d", "1x", "1d-12h", "12h1d", "1.2.3d", "1d12", "999999999w"} {
		_, err := types.ParseDuration(value)
		assert.Error(t, err, value)
	}
}

func Test_DurationUnmarshalers(t *testing.T) {
	var v struct {
		Timeout time.Duration  `json:"timeout"`
		TTL     *time.Duration `json:"ttl"`
	}
	opts := json.WithUnmarshalers(types.DurationUnmarshalers)

	// A duration string, or integer nanoseconds as a number or a string
	require.NoError(t, json.Unmarshal([]byte(`{"timeout":"5s","ttl":3600000000000}`), &v, opts))
	assert.Equal(t, 5*time.Second, v.Timeout)
	require.NotNil(t, v.TTL)
	assert.Equal(t, time.Hour, *v.TTL)
	require.NoError(t, json.Unmarshal([]byte(`{"timeout":"5000000000","ttl":null}`), &v, opts))
	assert.Equal(t, 5*time.Second, v.Timeout)
	assert.Nil(t, v.TTL)

	// An invalid duration names the value, including a fractional number
	err := json.Unmarshal([]byte(`{"timeout":"soon"}`), &v, opts)
	assert.ErrorContains(t, err, `"/timeout"`)
	err = json.Unmarshal([]byte(`{"timeout":1.5}`), &v, opts)
	assert.ErrorContains(t, err, `"/timeout"`)
}
