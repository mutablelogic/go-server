package httprequest_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Read_JSON(t *testing.T) {
	assert := assert.New(t)

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("Valid", func(t *testing.T) {
		body := strings.NewReader(`{"name":"alice","age":30}`)
		r, _ := http.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", "application/json")
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("alice", p.Name)
		assert.Equal(30, p.Age)
	})

	t.Run("EmptyBody", func(t *testing.T) {
		body := strings.NewReader("")
		r, _ := http.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", "application/json")
		var p payload
		err := httprequest.Read(r, &p)
		assert.Error(err)
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		body := strings.NewReader("{invalid")
		r, _ := http.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", "application/json")
		var p payload
		err := httprequest.Read(r, &p)
		assert.Error(err)
	})
}

func Test_Read_TextPlain_String(t *testing.T) {
	assert := assert.New(t)

	body := strings.NewReader("hello world")
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "text/plain")
	var s string
	err := httprequest.Read(r, &s)
	assert.NoError(err)
	assert.Equal("hello world", s)
}

func Test_Read_TextPlain_Bytes(t *testing.T) {
	assert := assert.New(t)

	body := strings.NewReader("binary data")
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "text/plain")
	var b []byte
	err := httprequest.Read(r, &b)
	assert.NoError(err)
	assert.Equal([]byte("binary data"), b)
}

func Test_Read_TextPlain_UnsupportedType(t *testing.T) {
	assert := assert.New(t)

	body := strings.NewReader("hello")
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "text/plain")
	var i int
	err := httprequest.Read(r, &i)
	assert.ErrorIs(err, httpresponse.ErrBadRequest)
}

func Test_Read_TextPlain_Struct(t *testing.T) {
	assert := assert.New(t)

	// A text body read into a struct, as for a JSON body, is a bad request
	type payload struct {
		Name string `json:"name"`
	}
	body := strings.NewReader(`{"name":"test"}`)
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "text/plain")
	var p payload
	err := httprequest.Read(r, &p)
	assert.ErrorIs(err, httpresponse.ErrBadRequest)
	assert.Zero(p)
}

func Test_Read_NoContentType(t *testing.T) {
	assert := assert.New(t)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	var s string
	err := httprequest.Read(r, &s)
	assert.Error(err)
}

func Test_Read_UnsupportedContentType(t *testing.T) {
	assert := assert.New(t)

	body := strings.NewReader("<xml/>")
	r, _ := http.NewRequest(http.MethodPost, "/", body)
	r.Header.Set("Content-Type", "application/xml")
	var s string
	err := httprequest.Read(r, &s)
	assert.Error(err)
}

func Test_Indent(t *testing.T) {
	assert := assert.New(t)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(2, httprequest.Indent(r))
}

func Test_Read_FormData(t *testing.T) {
	assert := assert.New(t)

	t.Run("WithFields", func(t *testing.T) {
		type payload struct {
			Name string `json:"name"`
			Age  string `json:"age"`
		}
		// Build multipart body
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		_ = w.WriteField("name", "alice")
		_ = w.WriteField("age", "30")
		w.Close()

		r, _ := http.NewRequest(http.MethodPost, "/", &buf)
		r.Header.Set("Content-Type", w.FormDataContentType())
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("alice", p.Name)
		assert.Equal("30", p.Age)
	})
}

func Test_Read_FormURLEncoded(t *testing.T) {
	assert := assert.New(t)

	t.Run("WithFields", func(t *testing.T) {
		type payload struct {
			Name   string   `json:"name"`
			Age    string   `json:"age"`
			Scopes []string `json:"scope"`
		}

		values := url.Values{
			"name":  {"alice"},
			"age":   {"30"},
			"scope": {"openid", "profile"},
		}
		r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("alice", p.Name)
		assert.Equal("30", p.Age)
		assert.Equal([]string{"openid", "profile"}, p.Scopes)
	})
}

func Test_Read_FormData_JSON(t *testing.T) {
	assert := assert.New(t)

	t.Run("JSONContentTypeWithCharset", func(t *testing.T) {
		type payload struct {
			Name string `json:"name"`
		}
		body := strings.NewReader(`{"name":"bob"}`)
		r, _ := http.NewRequest(http.MethodPost, "/", body)
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
		var p payload
		err := httprequest.Read(r, &p)
		assert.NoError(err)
		assert.Equal("bob", p.Name)
	})
}

func Test_Read_JSON_Duration(t *testing.T) {
	type Inner struct {
		Delay time.Duration `json:"delay"`
	}
	type Embedded struct {
		TTL *time.Duration `json:"ttl,omitempty"`
	}
	type payload struct {
		Embedded
		Name     string                   `json:"name"`
		Timeout  time.Duration            `json:"timeout"`
		Inner    Inner                    `json:"inner"`
		Pointer  *Inner                   `json:"pointer,omitempty"`
		Delays   []time.Duration          `json:"delays,omitempty"`
		Named    map[string]time.Duration `json:"named,omitempty"`
		Count    int64                    `json:"count"`
		Raw      json.RawMessage          `json:"raw,omitempty"`
		Interval time.Duration            `json:"interval"`
	}

	// read decodes the body into a payload
	read := func(t *testing.T, body string) (payload, error) {
		t.Helper()
		r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		var p payload
		return p, httprequest.Read(r, &p)
	}

	t.Run("Strings", func(t *testing.T) {
		// Duration strings are decoded wherever there's a duration
		p, err := read(t, `{"name":"a","ttl":"1h","timeout":"5s","inner":{"delay":"1m"},"pointer":{"delay":"2m"},"delays":["1s","2s"],"named":{"a":"3s"}}`)
		require.NoError(t, err)
		assert.Equal(t, "a", p.Name)
		require.NotNil(t, p.TTL)
		assert.Equal(t, time.Hour, *p.TTL)
		assert.Equal(t, 5*time.Second, p.Timeout)
		assert.Equal(t, time.Minute, p.Inner.Delay)
		require.NotNil(t, p.Pointer)
		assert.Equal(t, 2*time.Minute, p.Pointer.Delay)
		assert.Equal(t, []time.Duration{time.Second, 2 * time.Second}, p.Delays)
		assert.Equal(t, map[string]time.Duration{"a": 3 * time.Second}, p.Named)
	})

	t.Run("Integers", func(t *testing.T) {
		// Integer nanoseconds, as a time.Duration is encoded
		p, err := read(t, `{"ttl":3600000000000,"timeout":5000000000,"delays":[1000000000]}`)
		require.NoError(t, err)
		assert.Equal(t, time.Hour, *p.TTL)
		assert.Equal(t, 5*time.Second, p.Timeout)
		assert.Equal(t, []time.Duration{time.Second}, p.Delays)
	})

	t.Run("Exact", func(t *testing.T) {
		// Other values are decoded exactly: large integers keep their
		// precision, and raw JSON is kept
		p, err := read(t, `{"count":9007199254740993,"raw":{"b":1,"a":2},"interval":"1s"}`)
		require.NoError(t, err)
		assert.Equal(t, int64(9007199254740993), p.Count)
		assert.JSONEq(t, `{"b":1,"a":2}`, string(p.Raw))
		assert.Equal(t, time.Second, p.Interval)
	})

	t.Run("Invalid", func(t *testing.T) {
		_, err := read(t, `{"timeout":"soon"}`)
		assert.ErrorIs(t, err, httpresponse.ErrBadRequest)
		assert.ErrorContains(t, err, "timeout")
		_, err = read(t, ``)
		assert.ErrorIs(t, err, httpresponse.ErrBadRequest)
	})

	t.Run("NoDurations", func(t *testing.T) {
		// A body without durations is decoded as it was sent
		type raw struct {
			Raw json.RawMessage `json:"raw"`
		}
		r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(`{"raw":{"b": 1, "a": 2}}`))
		r.Header.Set("Content-Type", "application/json")
		var p raw
		require.NoError(t, httprequest.Read(r, &p))
		assert.Equal(t, `{"b": 1, "a": 2}`, string(p.Raw))
	})
}

// customTimeout decodes itself from {"timeout":"5s"}, so must be given the
// body as it was sent
type customTimeout struct {
	Timeout time.Duration
}

func (c *customTimeout) UnmarshalJSON(data []byte) error {
	var v struct {
		Timeout string `json:"timeout"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	d, err := time.ParseDuration(v.Timeout)
	c.Timeout = d
	return err
}

// Durations are decoded with encoding/json's rules for everything else
func Test_Read_JSON_DurationRules(t *testing.T) {
	read := func(t *testing.T, body string, v any) error {
		t.Helper()
		r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		return httprequest.Read(r, v)
	}

	t.Run("Unmarshaler", func(t *testing.T) {
		// A type with its own UnmarshalJSON is given the body as it was sent,
		// at the top level, and nested
		var top customTimeout
		require.NoError(t, read(t, `{"timeout":"5s"}`, &top))
		assert.Equal(t, 5*time.Second, top.Timeout)

		var nested struct {
			Custom customTimeout `json:"custom"`
		}
		require.NoError(t, read(t, `{"custom":{"timeout":"5s"}}`, &nested))
		assert.Equal(t, 5*time.Second, nested.Custom.Timeout)
	})

	t.Run("Embedded", func(t *testing.T) {
		// The fields of embedded structs with the same Go name and different
		// JSON names are both decoded, and an ignored embedded struct isn't
		type A struct {
			Delay time.Duration `json:"a"`
		}
		type B struct {
			Delay time.Duration `json:"b"`
		}
		type Ignored struct {
			Delay time.Duration `json:"ignored"`
		}
		var v struct {
			A
			B
			Ignored `json:"-"`
		}
		require.NoError(t, read(t, `{"a":"1s","b":"2s","ignored":"not a duration"}`, &v))
		assert.Equal(t, time.Second, v.A.Delay)
		assert.Equal(t, 2*time.Second, v.B.Delay)
		assert.Zero(t, v.Ignored.Delay)
	})

	t.Run("CaseInsensitive", func(t *testing.T) {
		var v struct {
			Timeout time.Duration `json:"timeout"`
		}
		require.NoError(t, read(t, `{"TIMEOUT":"5s"}`, &v))
		assert.Equal(t, 5*time.Second, v.Timeout)
	})

	t.Run("StringOption", func(t *testing.T) {
		// A field with the string option is integer nanoseconds as a string
		var v struct {
			Timeout time.Duration  `json:"timeout,string"`
			TTL     *time.Duration `json:"ttl,string"`
		}
		require.NoError(t, read(t, `{"timeout":"5000000000","ttl":"3600000000000"}`, &v))
		assert.Equal(t, 5*time.Second, v.Timeout)
		require.NotNil(t, v.TTL)
		assert.Equal(t, time.Hour, *v.TTL)
	})

	t.Run("DuplicateKeys", func(t *testing.T) {
		// The last of the keys which match a field wins
		var v struct {
			Timeout time.Duration `json:"timeout"`
			Name    string        `json:"name"`
		}
		require.NoError(t, read(t, `{"timeout":1,"name":"first","Name":"last"}`, &v))
		assert.Equal(t, "last", v.Name)
	})
}
