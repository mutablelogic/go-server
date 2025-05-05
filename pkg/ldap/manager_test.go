package ldap_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	// Packages

	"github.com/mutablelogic/go-server/pkg/httpresponse"
	ldap "github.com/mutablelogic/go-server/pkg/ldap"
	schema "github.com/mutablelogic/go-server/pkg/ldap/schema"
	assert "github.com/stretchr/testify/assert"
)

/////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Test_Manager_001(t *testing.T) {
	assert := assert.New(t)

	// Get opts
	opts := []ldap.Opt{}
	if url := os.Getenv("LDAP_URL"); url != "" {
		opts = append(opts, ldap.WithUrl(url))
	} else {
		t.Skip("Skipping test, LDAP_URL not set")
	}
	if dn := os.Getenv("LDAP_BASE_DN"); dn != "" {
		opts = append(opts, ldap.WithBaseDN(dn))
	} else {
		t.Skip("Skipping test, LDAP_BASE_DN not set")
	}

	// Create a new queue manager
	manager, err := ldap.NewManager(opts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// Cancellation context
	ctx, cancel := context.WithCancel(context.Background())

	// Run the manager in the background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(manager.Run(ctx))
	}()

	// Wait for the manager to start
	time.Sleep(time.Second)

	// Retrieve all objects
	objects, err := manager.List(ctx, schema.ObjectListRequest{})
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotZero(objects.Count)

	// Get each object
	for _, object := range objects.Body {
		object2, err := manager.Get(ctx, object.DN)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.Equal(object, object2)
	}

	// Cancel the context, wait for the manager to stop
	cancel()
	wg.Wait()
}

func Test_Manager_002(t *testing.T) {
	assert := assert.New(t)

	// Get opts
	opts := []ldap.Opt{}
	if url := os.Getenv("LDAP_URL"); url != "" {
		opts = append(opts, ldap.WithUrl(url))
	} else {
		t.Skip("Skipping test, LDAP_URL not set")
	}
	if dn := os.Getenv("LDAP_BASE_DN"); dn != "" {
		opts = append(opts, ldap.WithBaseDN(dn))
	} else {
		t.Skip("Skipping test, LDAP_BASE_DN not set")
	}

	// Create a new queue manager
	manager, err := ldap.NewManager(opts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// Cancellation context
	ctx, cancel := context.WithCancel(context.Background())

	// Run the manager in the background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(manager.Run(ctx))
	}()

	// Wait for the manager to start
	time.Sleep(time.Second)

	// Bind
	t.Log("Binding")
	_, err = manager.Bind(ctx, "uid=djt,ou=users,dc=mutablelogic,dc=com", "test")
	assert.ErrorIs(err, httpresponse.ErrNotAuthorized)

	// Cancel the context, wait for the manager to stop
	cancel()
	wg.Wait()
}

func Test_Manager_003(t *testing.T) {
	assert := assert.New(t)

	// Get opts
	opts := []ldap.Opt{}
	if url := os.Getenv("LDAP_URL"); url != "" {
		opts = append(opts, ldap.WithUrl(url))
	} else {
		t.Skip("Skipping test, LDAP_URL not set")
	}
	if dn := os.Getenv("LDAP_BASE_DN"); dn != "" {
		opts = append(opts, ldap.WithBaseDN(dn))
	} else {
		t.Skip("Skipping test, LDAP_BASE_DN not set")
	}

	// Create a new queue manager
	manager, err := ldap.NewManager(opts...)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(manager)

	// Cancellation context
	ctx, cancel := context.WithCancel(context.Background())

	// Run the manager in the background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(manager.Run(ctx))
	}()

	// Wait for the manager to start
	time.Sleep(time.Second)

	// Object Classes
	_, err = manager.ListObjectClasses(ctx)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Attribute Types
	types, err := manager.ListAttributeTypes(ctx)
	if !assert.NoError(err) {
		t.FailNow()
	}

	t.Log(types)

	// Cancel the context, wait for the manager to stop
	cancel()
	wg.Wait()
}
