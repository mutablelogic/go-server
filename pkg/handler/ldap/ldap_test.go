package ldap_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/ldap"
	provider "github.com/mutablelogic/go-server/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_ldap_001(t *testing.T) {
	assert := assert.New(t)
	ldap, err := ldap.New(ldap.Config{
		URL:  Get(t, "LDAP_URL", true),
		DN:   Get(t, "LDAP_DN", true),
		User: Get(t, "LDAP_USER", false),
	})
	assert.NoError(err)
	assert.NotNil(ldap)
}

func Test_ldap_002(t *testing.T) {
	assert := assert.New(t)
	ldap, err := ldap.New(ldap.Config{
		URL:      Get(t, "LDAP_URL", true),
		DN:       Get(t, "LDAP_DN", true),
		User:     Get(t, "LDAP_USER", false),
		Password: Get(t, "LDAP_PASSWORD", false),
	})
	if !assert.NoError(err) {
		t.SkipNow()
	}

	provider := provider.NewProvider(ldap)

	// Run the task in the background
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := provider.Run(ctx); err != nil {
			t.Error(err)
		}
	}()

	// Wait for connection
	time.Sleep(1 * time.Second)

	users, err := ldap.Get("posixAccount")
	assert.NoError(err)
	t.Log(users)

	// End
	cancel()

	// Wait for the task to complete
	wg.Wait()
}

///////////////////////////////////////////////////////////////////////////////
// ENVIRONMENT

func Get(t *testing.T, name string, required bool) string {
	key := os.Getenv(name)
	if key == "" && required {
		t.Skip("not set:", name)
		t.SkipNow()
	}
	return key
}
