package location_test

import (
	"testing"

	"github.com/djthorpe/goapp/pkg/location"
)

func Test_Location_000(t *testing.T) {
	l := location.NewLocation("World")
	if l == nil {
		t.Fatal("Unexpected nil return")
	}
}

func Test_Location_001(t *testing.T) {
	l := location.NewLocation("World")
	if l.Name != "World" {
		t.Fatal("Unexpected name value")
	}
}
