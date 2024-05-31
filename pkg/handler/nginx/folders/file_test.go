package folders_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/nginx/folders"
	"github.com/stretchr/testify/assert"
)

func Test_file_001(t *testing.T) {
	assert := assert.New(t)

	// Read a file
	f, err := folders.NewFile("../conf", "enabled/default.conf")
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(f)
	assert.NotEmpty(f.Path)
	assert.NotEmpty(f.Hash)
	t.Log(f)
}

func Test_file_002(t *testing.T) {
	assert := assert.New(t)

	dest := t.TempDir()

	// Read a file
	f, err := folders.NewFile("../conf", "enabled/default.conf")
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(f)

	// Copy a file to dest/enabled/default.conf
	err = f.Copy(dest, 0755)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Read the new file - check hash is the same
	f2, err := folders.NewFile(dest, f.Path)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(f2)
	assert.Equal(f.Hash, f2.Hash)
}
