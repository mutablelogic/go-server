package ddregister_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server/pkg/ddregister"
)

func Test_000(t *testing.T) {
	dd := New()
	if dd == nil {
		t.Error("dd is nil")
	} else {
		t.Log(dd)
	}
}

func Test_001(t *testing.T) {
	dd := New()
	ip, err := dd.GetExternalAddress()
	if err != nil {
		t.Error(err)
	}
	t.Log("addr=", ip)
}

func Test_002(t *testing.T) {
	hostuserpasswd := strings.SplitN(os.Getenv("GOOGLE_DOMAINS_AUTH"), ":", 3)
	if len(hostuserpasswd) != 3 {
		t.Skip("GOOGLE_DOMAINS_AUTH environment variable is not set, skipping test")
	}
	dd := New()
	addr, err := dd.GetExternalAddress()
	if err != nil {
		t.Error(err)
	}
	if err := dd.RegisterAddress(hostuserpasswd[0], hostuserpasswd[1], hostuserpasswd[2], addr, false); err != nil && !errors.Is(err, ErrNotModified) {
		t.Error(err)
	}
}
