package unmarshal_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	. "github.com/djthorpe/go-errors"
	"github.com/mutablelogic/go-server/pkg/httpserver/unmarshal"
	"github.com/stretchr/testify/assert"
)

type q struct {
	Str      string        `decode:"string"`
	Int      int           `decode:"int"`
	Uint     uint          `decode:"uint"`
	Float    float64       `decode:"float"`
	Bool     bool          `decode:"bool"`
	Time     time.Time     `decode:"time"`
	Duration time.Duration `decode:"duration"`
}

func Test_unmarshal_000(t *testing.T) {
	assert := assert.New(t)

	t.Run("ptr", func(t *testing.T) {
		var q q
		_, err := unmarshal.WithQuery(url.Values{}, q)
		if assert.Error(err) {
			assert.ErrorIs(err, ErrBadParameter)
		}
	})
	t.Run("str", func(t *testing.T) {
		_, err := unmarshal.WithQuery(url.Values{}, "hello")
		if assert.Error(err) {
			assert.ErrorIs(err, ErrBadParameter)
		}
	})
	t.Run("emptyq", func(t *testing.T) {
		var q q
		fields, err := unmarshal.WithQuery(url.Values{}, &q)
		assert.NoError(err)
		assert.Equal(0, len(fields))
	})
	t.Run("strq", func(t *testing.T) {
		var q q
		u := url.Values{}
		u.Set("string", "hello")
		fields, err := unmarshal.WithQuery(u, &q)
		assert.NoError(err)
		assert.Equal(fields, []string{"string"})
		assert.Equal("hello", q.Str)
	})
	t.Run("intq", func(t *testing.T) {
		var q q
		u := url.Values{}
		u.Set("int", "99")
		fields, err := unmarshal.WithQuery(u, &q)
		assert.NoError(err)
		assert.Equal(fields, []string{"int"})
		assert.Equal(99, q.Int)
	})
	t.Run("uintq", func(t *testing.T) {
		var q q
		u := url.Values{}
		u.Set("uint", "99")
		fields, err := unmarshal.WithQuery(u, &q)
		assert.NoError(err)
		assert.Equal(fields, []string{"uint"})
		assert.Equal(uint(99), q.Uint)
	})
	t.Run("boolq", func(t *testing.T) {
		var q q
		u := url.Values{}
		u.Set("bool", "true")
		fields, err := unmarshal.WithQuery(u, &q)
		assert.NoError(err)
		assert.Equal(fields, []string{"bool"})
		assert.Equal(true, q.Bool)
	})
	t.Run("floatq", func(t *testing.T) {
		var q q
		u := url.Values{}
		u.Set("float", "3.14")
		fields, err := unmarshal.WithQuery(u, &q)
		assert.NoError(err)
		assert.Equal(fields, []string{"float"})
		assert.Equal(3.14, q.Float)
	})
	t.Run("timeq", func(t *testing.T) {
		var q q
		u := url.Values{}
		now := time.Now().Truncate(time.Second)
		u.Set("time", now.Format(time.RFC3339))
		fields, err := unmarshal.WithQuery(u, &q)
		assert.NoError(err)
		assert.Equal(fields, []string{"time"})
		assert.Equal(now, q.Time)
	})
	t.Run("durationq", func(t *testing.T) {
		var q q
		u := url.Values{}
		dur := time.Second * 4000
		u.Set("duration", fmt.Sprint(dur))
		fields, err := unmarshal.WithQuery(u, &q)
		assert.NoError(err)
		assert.Equal(fields, []string{"duration"})
		assert.Equal(dur, q.Duration)
	})
}
