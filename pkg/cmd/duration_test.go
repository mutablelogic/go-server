package cmd

import (
	"reflect"
	"testing"
	"time"

	// Packages
	kong "github.com/alecthomas/kong"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func Test_DurationMapper(t *testing.T) {
	// A duration flag can have days and weeks, as can a pointer to one
	var cli struct {
		Interval time.Duration  `name:"interval"`
		TTL      *time.Duration `name:"ttl" optional:""`
	}
	parser, err := kong.New(&cli, kong.TypeMapper(reflect.TypeOf(time.Duration(0)), durationMapper()))
	require.NoError(t, err)

	_, err = parser.Parse([]string{"--interval=7d", "--ttl=1d12h"})
	require.NoError(t, err)
	assert.Equal(t, 7*24*time.Hour, cli.Interval)
	require.NotNil(t, cli.TTL)
	assert.Equal(t, 36*time.Hour, *cli.TTL)

	_, err = parser.Parse([]string{"--interval=90m"})
	require.NoError(t, err)
	assert.Equal(t, 90*time.Minute, cli.Interval)

	_, err = parser.Parse([]string{"--interval=soon"})
	assert.ErrorContains(t, err, `expected duration but got "soon"`)
}
