package tokenjar_test

import (
	"context"
	"sync"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/auth"
	"github.com/mutablelogic/go-server/pkg/handler/tokenjar"
	"github.com/stretchr/testify/assert"
)

const (
	ScopeRead  = "read"
	ScopeWrite = "write"
)

func Test_tokenjar_001(t *testing.T) {
	assert := assert.New(t)

	// Create a persistent token jar
	tokens, err := tokenjar.New(tokenjar.Config{
		DataPath: t.TempDir(),
	})
	assert.NoError(err)
	assert.NotNil(tokens)
}

func Test_tokenjar_002(t *testing.T) {
	assert := assert.New(t)

	// Create a persistent token jar
	tokens, err := tokenjar.New(tokenjar.Config{
		DataPath: t.TempDir(),
	})
	assert.NoError(err)
	assert.NotNil(tokens)

	// Run the token jar
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(tokens.Run(ctx))
	}()

	// Add a token
	token := auth.NewToken("test", 100, 0)
	assert.NoError(tokens.Create(token))
	t.Log(token)

	// Get a token
	token2 := tokens.Get(token.Value)
	assert.NotNil(token2)
	assert.True(token.Equals(token2))

	// Update a token
	token2.Scope = []string{ScopeRead}
	assert.NoError(tokens.Update(token2))

	token3 := tokens.Get(token.Value)
	assert.NotNil(token3)
	assert.True(token3.Equals(token2))

	// Remove a token
	assert.NoError(tokens.Delete(token.Value))

	// Cancel the context and wait
	cancel()
	wg.Wait()
}
