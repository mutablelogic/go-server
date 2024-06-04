package auth_test

import (
	"encoding/json"
	"testing"
	"time"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/auth"
	"github.com/stretchr/testify/assert"
)

const (
	ScopeRead  = "read"
	ScopeWrite = "write"
)

func Test_token_001(t *testing.T) {
	assert := assert.New(t)
	// Create a token with 100 bytes and no expiry
	token := auth.NewToken("test", 100, 0)
	assert.NotNil(token)
	assert.Equal(200, len(token.Value))
	assert.True(token.IsValid())
	t.Log(token)
}

func Test_token_002(t *testing.T) {
	assert := assert.New(t)
	// Create a token with 100 bytes and expiry in the past
	token := auth.NewToken("test", 100, -time.Second)
	assert.NotNil(token)
	assert.False(token.IsValid())
	assert.Equal(200, len(token.Value))
	t.Log(token)
}

func Test_token_003(t *testing.T) {
	assert := assert.New(t)
	// Create a token with 100 bytes and one scope
	token := auth.NewToken("test", 100, 0, ScopeRead)
	assert.NotNil(token)
	assert.True(token.IsScope(ScopeRead))
	assert.False(token.IsScope(ScopeWrite))
	t.Log(token)
}

func Test_token_004(t *testing.T) {
	assert := assert.New(t)
	// Create a token with 100 bytes and one scope
	token := auth.NewToken("test", 100, -1, ScopeRead)
	assert.NotNil(token)
	assert.False(token.IsScope(ScopeRead))
	t.Log(token)
}

func Test_token_005(t *testing.T) {
	assert := assert.New(t)

	// Create a token with 100 bytes and one scope
	token := auth.NewToken("test", 100, -1, ScopeRead)
	assert.NotNil(token)

	// Write to string
	bytes, err := json.MarshalIndent(token, "", "  ")
	assert.NoError(err)
	assert.NotNil(bytes)
	t.Log(string(bytes))
}
