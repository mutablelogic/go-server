package nginx_test

import (
	"context"
	"os/exec"
	"regexp"
	"sync"
	"testing"
	"time"

	// Packages
	nginx "github.com/mutablelogic/go-server/pkg/handler/nginx"
	cmd "github.com/mutablelogic/go-server/pkg/handler/nginx/cmd"
	assert "github.com/stretchr/testify/assert"
)

func Test_nginx_001(t *testing.T) {
	assert := assert.New(t)
	config := nginx.Config{}
	assert.NotEmpty(config.Name())
	assert.NotEmpty(config.Description())
}

func Test_nginx_002(t *testing.T) {
	assert := assert.New(t)
	task, err := nginx.New(nginx.Config{
		BinaryPath: BinaryExec(t),
	})
	assert.NoError(err)
	t.Log(task.Version())
}

func Test_nginx_003(t *testing.T) {
	var wg sync.WaitGroup

	// Create a new task
	assert := assert.New(t)
	task, err := nginx.New(nginx.Config{
		BinaryPath: BinaryExec(t),
	})
	assert.NoError(err)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel and then wait for the task to finish
	defer wg.Wait()
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := task.Run(ctx); err != nil {
			t.Error(err)
		}
	}()

	// Sleep until the task is running
	time.Sleep(500 * time.Millisecond)

	// Test server configuration
	err = task.Test()
	assert.NoError(err)
}

func BinaryExec(t *testing.T) string {
	var version string

	bin, err := exec.LookPath("nginx")
	if err != nil {
		t.Skip("Skipping test, nginx binary not found")
		return ""
	}

	// TODO minimum version is 1.19.15
	reVersion := regexp.MustCompile(`nginx/(\d+)\.(\d+)\.(\d+)`)
	cmd, err := cmd.New(bin, "-v")
	cmd.Err = func(data []byte) {
		version += string(data)
	}
	cmd.Out = func(data []byte) {
		version += string(data)
	}
	if err != nil {
		t.Skip(err.Error())
	} else if err := cmd.Run(); err != nil {
		t.Skip(err.Error())
	} else if args := reVersion.FindStringSubmatch(version); args == nil {
		t.Skip("Missing version: " + version)
	} else if v := args[1] + args[2]; v < "119" {
		t.Skip("Invalid version (needs to be at least 1.19.X): " + version)
	} else if v == "119" && args[3] < "15" {
		t.Skip("Invalid version (needs to be >= 1.19.15): " + version)
	}

	// Return binary path
	return bin
}
