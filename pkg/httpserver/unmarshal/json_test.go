package unmarshal_test

import (
	"strings"
	"testing"

	"github.com/mutablelogic/go-server/pkg/httpserver/unmarshal"
	"github.com/stretchr/testify/assert"
)

func Test_json_000(t *testing.T) {
	assert := assert.New(t)

	t.Run("t0", func(t *testing.T) {
		var q q
		r := strings.NewReader(`{ "string": "hello", "int": 99, "uint": 99, "float": 99.99, "bool": true, "time": "2019-01-01T01:02:03Z", "duration": "1h" }`)
		_, err := unmarshal.WithJson(r, &q)
		assert.NoError(err)
		assert.Equal("hello", q.Str)
		assert.Equal(99, q.Int)
		assert.Equal(uint(99), q.Uint)
		assert.Equal(99.99, q.Float)
		assert.Equal(true, q.Bool)
		assert.Equal("2019-01-01 01:02:03 +0000 UTC", q.Time.String())
		assert.Equal("1h0m0s", q.Duration.String())
	})
}
