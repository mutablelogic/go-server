package htpasswd_test

import (
	"bytes"
	"testing"

	. "github.com/djthorpe/go-server/pkg/htpasswd"
)

func Test_000(t *testing.T) {
	var tests = []struct {
		user      string
		passwd    string
		algorithm HashAlgorithm
	}{
		{"a", "a", MD5},
		{"b", "a", BCrypt},
		{"c", "a", SHA},
		{"d", ":hello, world!", MD5},
		{"e", ":hello, world!:", BCrypt},
		{"f", "hello, world!", SHA},
	}

	passwd := New()
	if passwd == nil {
		t.Errorf("Expected non-nil htpasswd")
	}
	for _, test := range tests {
		if err := passwd.Set(test.user, test.passwd, test.algorithm); err != nil {
			t.Error(err)
		} else if passwd.Verify(test.user, test.passwd) == false {
			t.Errorf("Unexpected return from Verify for user %q, algorithm=%v", test.user, test.algorithm)
		}
	}

	buf := new(bytes.Buffer)
	if err := passwd.Write(buf); err != nil {
		t.Error(err)
	} else if _, err := Read(buf); err != nil {
		t.Error(err)
	} else {
		passwd.Write(buf)
		t.Log(passwd, "=", buf.String())
	}
}
