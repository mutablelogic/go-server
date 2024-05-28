package folders_test

import (
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/nginx/folders"
	"github.com/stretchr/testify/assert"
)

func Test_config_001(t *testing.T) {
	assert := assert.New(t)

	// Create temp folders for available and enabled
	available, enabled := t.TempDir(), t.TempDir()

	// Copy the file
	src, err := folders.NewFile("../conf/enabled", "default.conf")
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Put default.conf into enabled folder
	assert.NoError(src.Copy(enabled, 0755))

	// Should copy into the available folder
	config, err := folders.New(available, enabled, ".conf", false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(config)

	// Should be one file in each folder
	assert.Len(config.Available.Files, 1)
	assert.Len(config.Enabled.Files, 1)

	// Should be one template, which is enabled
	templates := config.Templates()
	assert.Len(templates, 1)
	assert.True(templates[0].Enabled)
}

func Test_config_002(t *testing.T) {
	assert := assert.New(t)

	// Create temp folders for available and enabled
	available, enabled := t.TempDir(), t.TempDir()

	// Get the source file
	src, err := folders.NewFile("../conf/enabled", "default.conf")
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Put default.conf into available folder
	assert.NoError(src.Copy(available, 0755))

	// Make a config object
	config, err := folders.New(available, enabled, ".conf", false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(config)

	// Should be one file in available, but nothing in enabled
	assert.Len(config.Available.Files, 1)
	assert.Len(config.Enabled.Files, 0)

	// Should be one template, which is not enabled
	templates := config.Templates()
	assert.Len(templates, 1)
	assert.False(templates[0].Enabled)

	// Enable it
	assert.NoError(config.Enable("default.conf"))

	// Should be one template, which is now enabled
	templates = config.Templates()
	assert.Len(templates, 1)
	assert.True(templates[0].Enabled)

	// Disable it
	assert.NoError(config.Disable("default.conf"))

	// Should be one template, which is now disabled
	templates = config.Templates()
	assert.Len(templates, 1)
	assert.False(templates[0].Enabled)
}

func Test_config_003(t *testing.T) {
	assert := assert.New(t)

	// Create temp folders for available and enabled
	available, enabled := t.TempDir(), t.TempDir()

	// Get the source file
	src, err := folders.NewFile("../conf/enabled", "default.conf")
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Make a config object
	config, err := folders.New(available, enabled, ".conf", false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(config)

	// Should be zero files
	assert.Len(config.Available.Files, 0)
	assert.Len(config.Enabled.Files, 0)

	// Read the body
	data, err := src.Render()
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Should be one template, which is not enabled
	assert.NoError(config.Create("default.conf", data))

	// Should be one file in available, but nothing in enabled
	templates := config.Templates()
	assert.Len(templates, 1)
	assert.Equal(templates[0].Name, "default.conf")
	assert.False(templates[0].Enabled)
}
