package nginx_test

import (
	"testing"

	"github.com/mutablelogic/go-server/pkg/nginx"
)

func Test_File_000(t *testing.T) {
	f := nginx.File{Path: "test/test.conf"}
	t.Log("enabled=", f.EnabledBase(), " available=", f.AvailableBase())
}
