package httprequest_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/stretchr/testify/assert"
)

func Test_Query_String(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Name string `json:"name"`
	}

	t.Run("SetString", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"name": {"alice"}}, &p)
		assert.NoError(err)
		assert.Equal("alice", p.Name)
	})

	t.Run("MissingKey", func(t *testing.T) {
		var p params
		p.Name = "default"
		err := httprequest.Query(url.Values{}, &p)
		assert.NoError(err)
		assert.Equal("", p.Name) // zeroed
	})

	t.Run("EmptyValues", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"name": {}}, &p)
		assert.NoError(err)
		assert.Equal("", p.Name) // zeroed by empty slice
	})
}

func Test_Query_Bool(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Active bool `json:"active"`
	}

	t.Run("True", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"active": {"true"}}, &p)
		assert.NoError(err)
		assert.True(p.Active)
	})

	t.Run("False", func(t *testing.T) {
		var p params
		p.Active = true
		err := httprequest.Query(url.Values{"active": {"false"}}, &p)
		assert.NoError(err)
		assert.False(p.Active)
	})

	t.Run("Invalid", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"active": {"yes-please"}}, &p)
		assert.Error(err)
	})
}

func Test_Query_Int(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Count  int   `json:"count"`
		Small  int8  `json:"small"`
		Big    int64 `json:"big"`
	}

	t.Run("SetInt", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"count": {"42"}, "small": {"-5"}, "big": {"9999999"}}, &p)
		assert.NoError(err)
		assert.Equal(42, p.Count)
		assert.Equal(int8(-5), p.Small)
		assert.Equal(int64(9999999), p.Big)
	})

	t.Run("Invalid", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"count": {"abc"}}, &p)
		assert.Error(err)
	})
}

func Test_Query_Uint(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Offset uint   `json:"offset"`
		Limit  uint64 `json:"limit"`
	}

	t.Run("SetUint", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"offset": {"10"}, "limit": {"100"}}, &p)
		assert.NoError(err)
		assert.Equal(uint(10), p.Offset)
		assert.Equal(uint64(100), p.Limit)
	})

	t.Run("Invalid", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"offset": {"-1"}}, &p)
		assert.Error(err)
	})
}

func Test_Query_Float(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Score   float64 `json:"score"`
		Score32 float32 `json:"score32"`
	}

	t.Run("SetFloat", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"score": {"3.14"}, "score32": {"2.5"}}, &p)
		assert.NoError(err)
		assert.InDelta(3.14, p.Score, 0.001)
		assert.InDelta(2.5, p.Score32, 0.001)
	})

	t.Run("Invalid", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"score": {"abc"}}, &p)
		assert.Error(err)
	})
}

func Test_Query_Time(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		CreatedAt time.Time `json:"created_at"`
	}

	t.Run("RFC3339", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"created_at": {"2025-01-15T10:30:00Z"}}, &p)
		assert.NoError(err)
		assert.Equal(2025, p.CreatedAt.Year())
		assert.Equal(time.Month(1), p.CreatedAt.Month())
		assert.Equal(15, p.CreatedAt.Day())
	})

	t.Run("Invalid", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"created_at": {"not-a-date"}}, &p)
		assert.Error(err)
	})
}

func Test_Query_Pointer(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Name *string `json:"name"`
	}

	t.Run("SetPointer", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"name": {"bob"}}, &p)
		assert.NoError(err)
		if assert.NotNil(p.Name) {
			assert.Equal("bob", *p.Name)
		}
	})

	t.Run("NilWhenMissing", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{}, &p)
		assert.NoError(err)
		assert.Nil(p.Name)
	})
}

func Test_Query_StringSlice(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Tags []string `json:"tags"`
	}

	t.Run("MultipleValues", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"tags": {"a", "b", "c"}}, &p)
		assert.NoError(err)
		assert.Equal([]string{"a", "b", "c"}, p.Tags)
	})
}

func Test_Query_TagIgnore(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Name    string `json:"name"`
		Ignored string `json:"-"`
	}

	t.Run("IgnoredField", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"name": {"alice"}, "Ignored": {"should-not-set"}}, &p)
		assert.NoError(err)
		assert.Equal("alice", p.Name)
		assert.Equal("", p.Ignored)
	})
}

func Test_Query_UnsupportedType(t *testing.T) {
	assert := assert.New(t)

	type inner struct {
		X int
	}

	type params struct {
		Nested inner `json:"nested"`
	}

	t.Run("UnsupportedStruct", func(t *testing.T) {
		var p params
		err := httprequest.Query(url.Values{"nested": {"value"}}, &p)
		assert.Error(err)
	})
}

func Test_Query_Errors(t *testing.T) {
	assert := assert.New(t)

	t.Run("NotAPointer", func(t *testing.T) {
		type params struct{ Name string }
		var p params
		err := httprequest.Query(url.Values{}, p)
		assert.Error(err)
	})

	t.Run("NotAStruct", func(t *testing.T) {
		s := "hello"
		err := httprequest.Query(url.Values{}, &s)
		assert.Error(err)
	})
}

func Test_Query_NoTagUsesFieldName(t *testing.T) {
	assert := assert.New(t)

	type params struct {
		Name string
	}

	var p params
	err := httprequest.Query(url.Values{"Name": {"alice"}}, &p)
	assert.NoError(err)
	assert.Equal("alice", p.Name)
}
