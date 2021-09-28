package template_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	// Modules
	"github.com/mutablelogic/go-server/pkg/template"
)

const (
	TEMPLATE_PATH = "../../etc/templates"
)

func Test_Cache_000(t *testing.T) {
	if path, err := filepath.Abs(TEMPLATE_PATH); err != nil {
		t.Fatal(err)
	} else if cache, err := template.NewCache(os.DirFS(path)); err != nil {
		t.Fatal(err)
	} else {
		t.Log(path, "=>", cache)
	}
}
func Test_Cache_001(t *testing.T) {
	if path, err := filepath.Abs(TEMPLATE_PATH); err != nil {
		t.Fatal(err)
	} else if cache, err := template.NewCache(os.DirFS(path)); err != nil {
		t.Fatal(err)
	} else if templates := cache.Templates(); len(templates) == 0 {
		t.Error("Expected non-empty templates")
	} else {
		w := bytes.NewBuffer(nil)
		if err := cache.Exec(w, templates[0], nil); err != nil {
			t.Error(err)
		} else {
			t.Logf("%q => %q", templates[0], w.String())
		}
	}
}
