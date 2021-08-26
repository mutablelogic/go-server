package mdns_test

import (
	"testing"

	. "github.com/djthorpe/go-server/pkg/mdns"
)

func Test_ServiceDatabase_001(t *testing.T) {
	if db, err := ReadServiceDatabase(DefaultServiceDatabase); err != nil {
		t.Error(err)
	} else {
		t.Log(db)
	}
}
