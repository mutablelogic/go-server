package tokenauth_test

import (
	"context"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpserver/tokenauth"
	"github.com/mutablelogic/go-server/pkg/task"
	"github.com/stretchr/testify/assert"
)

func Test_tokenauth_001(t *testing.T) {
	assert := assert.New(t)
	provider, err := task.NewProvider(context.Background(), tokenauth.WithLabel(t.Name()).WithPath(t.TempDir()))
	assert.NoError(err)
	assert.NotNil(provider)
	t.Log(provider)
}

/*
func Test_TokenAuth_001(t *testing.T) {
	// Create an "tokens" folder
	path, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	task, err := Config{
		Path: path,
	}.New(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(task)
}

func Test_TokenAuth_002(t *testing.T) {
	// Create an "tokens" folder
	path, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	// Create a configuration
	auth, err := Config{Path: path, Delta: time.Second}.New(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Write out events
	go func() {
		for evt := range auth.C() {
			t.Log(evt)
		}
	}()

	// Revoke the admin token to rotate it, three times
	go func() {
		for i := 0; i < 3; i++ {
			time.Sleep(time.Second)
			t.Log("Revoke admin token")
			if err := auth.(plugin.TokenAuth).Revoke(AdminToken); err != nil {
				t.Error(err)
			}
		}
	}()

	// Run in foreground for a few seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := auth.Run(ctx); err != nil {
		t.Error(err)
	}
}

func Test_TokenAuth_003(t *testing.T) {
	// Create an "tokens" folder
	path, err := os.MkdirTemp("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	// Create a configuration
	auth, err := Config{Path: path, Delta: time.Millisecond * 100}.New(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	// Write out events
	go func() {
		for evt := range auth.C() {
			t.Log(evt)
		}
	}()

	// Create and then revoke a token
	go func() {
		value, err := auth.(plugin.TokenAuth).Create("test")
		if err != nil {
			t.Error(err)
		} else {
			t.Log("token=", value)
		}
		time.Sleep(2 * time.Second)
		t.Log("Testing test token value for match")
		if matches := auth.(plugin.TokenAuth).Matches(value); matches != "test" {
			t.Error("Expected token to be 'test'")
		}
		time.Sleep(2 * time.Second)
		t.Log("Revoking test token")
		if err := auth.(plugin.TokenAuth).Revoke("test"); err != nil {
			t.Error(err)
		} else {
			t.Log("Revoked token=", value)
		}
		time.Sleep(2 * time.Second)
		if matches := auth.(plugin.TokenAuth).Matches(value); matches != "" {
			t.Error("Expected token to be empty")
		}
	}()

	// Run in foreground for a few seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := auth.Run(ctx); err != nil {
		t.Error(err)
	}
}
*/
