package folders_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/nginx/folders"
	"github.com/stretchr/testify/assert"
)

func Test_folder_001(t *testing.T) {
	assert := assert.New(t)

	// Read a folder
	folder, err := folders.NewFolder("../conf/enabled", ".conf", true)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(folder)
	t.Log(folder)
}
