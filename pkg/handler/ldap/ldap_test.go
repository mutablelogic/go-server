package ldap_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/ldap"
	"github.com/mutablelogic/go-server/pkg/handler/ldap/schema"
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
	if !assert.NoError(err) {
		t.SkipNow()
	}
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

	groups, err := ldap.GetGroups()
	assert.NoError(err)
	t.Log(groups)

	// End
	cancel()

	// Wait for the task to complete
	wg.Wait()
}

func Test_ldap_003(t *testing.T) {
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
	time.Sleep(500 * time.Millisecond)

	// Create a new group
	group, err := ldap.CreateGroup("testgroup3", schema.OptDescription("This is a test group"))
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Create user
	user, err := ldap.CreateUser("testuser3",
		schema.OptGroupId(group.GroupId()),
		schema.OptName("Test", "User"),
		schema.OptDescription("This is a test user"),
		schema.OptHomeDirectory("/home/testuser3"),
		schema.OptMail("djt@test.com"),
	)
	if !assert.NoError(err) {
		ldap.Delete(group)
		t.SkipNow()
	}

	t.Log(user, group)

	if passwd, err := ldap.ChangePassword(user, "", ""); !assert.NoError(err) {
		t.SkipNow()
	} else {
		t.Log("password=", passwd)
	}

	if err := ldap.Delete(group); !assert.NoError(err) {
		t.SkipNow()
	}

	if err := ldap.Delete(user); !assert.NoError(err) {
		t.SkipNow()
	}

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
