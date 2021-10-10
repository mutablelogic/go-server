package template_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	// Modules
	"github.com/mutablelogic/go-server/pkg/template"
)

const (
	DIR_PATH = "../.."
)

func Test_Mime_000(t *testing.T) {
	m := template.NewContentTypeDetect()
	if m == nil {
		t.Fatal("NewContentTypeDetect() returned nil")
	}
	abspath, err := filepath.Abs(DIR_PATH)
	if err != nil {
		t.Fatal(err)
	}
	dirs := []string{abspath}
	for {
		if len(dirs) == 0 {
			break
		}
		dir := dirs[0]
		dirs = dirs[1:]
		readDir(t, dir, func(path string, info os.FileInfo) {
			if info.IsDir() {
				dirs = append(dirs, path)
				return
			}
			r, err := os.Open(path)
			if err != nil {
				t.Error(err)
			}
			defer r.Close()
			if mimetype, charset, err := m.DetectContentType(r, info); err != nil {
				t.Error(path, "=>", err)
			} else {
				t.Log(path, "=>", mimetype, "[", charset, "]")
			}
		})
	}
}

func readDir(t *testing.T, path string, f func(string, os.FileInfo)) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		f(filepath.Join(path, file.Name()), file)
	}
}
