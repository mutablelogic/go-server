package auth_test

import (
	"context"
	"testing"
	"time"

	// Packages

	"github.com/mutablelogic/go-server/pkg/handler/auth"
	"github.com/mutablelogic/go-server/pkg/handler/tokenjar"
	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_auth_001(t *testing.T) {
	assert := assert.New(t)

	// Create a token jar and auth object
	jar, err := tokenjar.New(tokenjar.Config{
		DataPath: t.TempDir(),
	})
	assert.NoError(err)
	tokens, err := auth.New(auth.Config{
		TokenJar: jar,
	})
	assert.NoError(err)

	// Create a provider
	provider := provider.NewProvider(jar, tokens)
	assert.NotNil(provider)

	// Run the provider
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Run the provider
	assert.NoError(provider.Run(ctx))
}
