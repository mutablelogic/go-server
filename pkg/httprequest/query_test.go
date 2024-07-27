package httprequest_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/mutablelogic/go-server/pkg/httprequest"
	"github.com/stretchr/testify/assert"
)

type querystr struct {
	Str    string   `json:"str"`
	StrPtr *string  `json:"strptr"`
	StrArr []string `json:"strarr"`
}

type querybool struct {
	Bool    bool   `json:"value"`
	BoolPtr *bool  `json:"ptr"`
	BoolArr []bool `json:"arr"`
}

type queryint struct {
	Int    int   `json:"value"`
	IntPtr *int  `json:"ptr"`
	IntArr []int `json:"arr"`
}

type queryuint struct {
	Uint    uint   `json:"value"`
	UintPtr *uint  `json:"ptr"`
	UintArr []uint `json:"arr"`
}

type queryfloat struct {
	Float32 float32 `json:"value32"`
	Float64 float64 `json:"value64"`
}

type querytime struct {
	Time    time.Time       `json:"time"`
	Dur     time.Duration   `json:"dur"`
	TimePtr *time.Time      `json:"timeptr"`
	DurArr  []time.Duration `json:"durarr"`
}

func ptr(v string) *string {
	return &v
}

func boolptr(v bool) *bool {
	return &v
}

func intptr(v int) *int {
	return &v
}

func uintptr(v uint) *uint {
	return &v
}

func timeptr(v time.Time) *time.Time {
	return &v
}

func durptr(v time.Duration) *time.Duration {
	return &v
}

func timeof(str string) time.Time {
	v, _ := time.Parse(time.RFC3339, str)
	return v
}

func Test_query_00(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		In  url.Values
		Out querystr
	}{
		{url.Values{}, querystr{"", nil, nil}},
		{url.Values{"str": []string{"foo", "bar"}}, querystr{"foo", nil, nil}},
		{url.Values{"str": []string{"foo", "bar"}, "strptr": []string{"a", "b"}}, querystr{"foo", ptr("a"), nil}},
		{url.Values{"str": []string{"foo", "bar"}, "strptr": []string{"a", "b"}, "strarr": []string{"c", "d"}}, querystr{"foo", ptr("a"), []string{"c", "d"}}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			var result querystr
			err := httprequest.Query(&result, test.In)
			assert.NoError(err)
			assert.Equal(test.Out, result)
		})
	}
}

func Test_query_01(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		In  url.Values
		Out querybool
	}{
		{url.Values{}, querybool{false, nil, nil}},
		{url.Values{"value": []string{"false"}}, querybool{false, nil, nil}},
		{url.Values{"value": []string{"true"}}, querybool{true, nil, nil}},
		{url.Values{"value": []string{"true", "false"}}, querybool{true, nil, nil}},
		{url.Values{"value": []string{"false", "true"}}, querybool{false, nil, nil}},
		{url.Values{"ptr": []string{"false"}}, querybool{false, boolptr(false), nil}},
		{url.Values{"ptr": []string{"true"}}, querybool{false, boolptr(true), nil}},
		{url.Values{"arr": []string{"false"}}, querybool{false, nil, []bool{false}}},
		{url.Values{"arr": []string{"true"}}, querybool{false, nil, []bool{true}}},
		{url.Values{"arr": []string{"false", "true"}}, querybool{false, nil, []bool{false, true}}},
		{url.Values{"arr": []string{"true", "false"}}, querybool{false, nil, []bool{true, false}}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint("test", i), func(t *testing.T) {
			var result querybool
			err := httprequest.Query(&result, test.In)
			if !assert.NoError(err) {
				t.SkipNow()
			}
			assert.Equal(test.Out, result)
		})
	}
}

func Test_query_02(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		In  url.Values
		Out queryint
	}{
		{url.Values{}, queryint{0, nil, nil}},
		{url.Values{"value": []string{"-1"}}, queryint{-1, nil, nil}},
		{url.Values{"value": []string{"-1", "+1"}}, queryint{-1, nil, nil}},
		{url.Values{"value": []string{"+1", "-1"}}, queryint{+1, nil, nil}},
		{url.Values{"ptr": []string{"+2"}}, queryint{0, intptr(+2), nil}},
		{url.Values{"arr": []string{"+3"}}, queryint{0, nil, []int{+3}}},
		{url.Values{"arr": []string{"+3", "+4"}}, queryint{0, nil, []int{+3, +4}}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint("test", i), func(t *testing.T) {
			var result queryint
			err := httprequest.Query(&result, test.In)
			if !assert.NoError(err) {
				t.SkipNow()
			}
			assert.Equal(test.Out, result)
		})
	}
}

func Test_query_03(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		In  url.Values
		Out queryuint
	}{
		{url.Values{}, queryuint{0, nil, nil}},
		{url.Values{"value": []string{"1"}}, queryuint{1, nil, nil}},
		{url.Values{"ptr": []string{"1"}}, queryuint{0, uintptr(1), nil}},
		{url.Values{"arr": []string{"3", "4"}}, queryuint{0, nil, []uint{+3, +4}}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint("test", i), func(t *testing.T) {
			var result queryuint
			err := httprequest.Query(&result, test.In)
			if !assert.NoError(err) {
				t.SkipNow()
			}
			assert.Equal(test.Out, result)
		})
	}
}

func Test_query_04(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		In  url.Values
		Out queryfloat
	}{
		{url.Values{}, queryfloat{0, 0}},
		{url.Values{"value32": []string{"3.14"}}, queryfloat{3.14, 0}},
		{url.Values{"value64": []string{"3.14"}}, queryfloat{0, 3.14}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint("test", i), func(t *testing.T) {
			var result queryfloat
			err := httprequest.Query(&result, test.In)
			if !assert.NoError(err) {
				t.SkipNow()
			}
			assert.Equal(test.Out, result)
		})
	}
}

func Test_query_05(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		In  url.Values
		Out querytime
	}{
		{url.Values{}, querytime{time.Time{}, 0, nil, nil}},
		{url.Values{"time": []string{"null"}}, querytime{time.Time{}, 0, nil, nil}},
		{url.Values{"timeptr": []string{"null"}}, querytime{time.Time{}, 0, timeptr(time.Time{}), nil}},
		{url.Values{"time": []string{"2012-04-23T18:25:43.511Z"}}, querytime{timeof("2012-04-23T18:25:43.511Z"), 0, nil, nil}},
		{url.Values{"dur": []string{"1s"}}, querytime{time.Time{}, time.Second, nil, nil}},
		{url.Values{"dur": []string{"1m"}}, querytime{time.Time{}, time.Minute, nil, nil}},
		{url.Values{"durarr": []string{"1m", "1h"}}, querytime{time.Time{}, 0, nil, []time.Duration{time.Minute, time.Hour}}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint("test", i), func(t *testing.T) {
			var result querytime
			err := httprequest.Query(&result, test.In)
			if !assert.NoError(err) {
				t.SkipNow()
			}
			assert.Equal(test.Out, result)
		})
	}
}
