package client_test

import (
	"os"
	"testing"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	auth "github.com/mutablelogic/go-server/pkg/handler/auth/client"
	assert "github.com/stretchr/testify/assert"
)

func Test_client_001(t *testing.T) {
	assert := assert.New(t)
	opts := []client.ClientOpt{
		client.OptTrace(os.Stderr, true),
	}
	client, err := auth.New(GetEndpoint(t), opts...)
	assert.NoError(err)
	assert.NotNil(client)
}

func Test_client_002(t *testing.T) {
	assert := assert.New(t)
	opts := []client.ClientOpt{
		client.OptTrace(os.Stderr, true),
	}
	if token := GetToken(t); token != "" {
		opts = append(opts, client.OptReqToken(client.Token{
			Scheme: "Bearer",
			Value:  token,
		}))
	}

	client, err := auth.New(GetEndpoint(t), opts...)
	assert.NoError(err)

	tokens, err := client.List()
	assert.NoError(err)
	assert.NotEmpty(tokens)
}

func Test_client_003(t *testing.T) {
	assert := assert.New(t)
	opts := []client.ClientOpt{
		client.OptTrace(os.Stderr, true),
	}
	if token := GetToken(t); token != "" {
		opts = append(opts, client.OptReqToken(client.Token{
			Scheme: "Bearer",
			Value:  token,
		}))
	}
	client, err := auth.New(GetEndpoint(t), opts...)
	assert.NoError(err)

	// Create a new token which doesn't expire
	token, err := client.Create(t.Name(), 0)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Get the token
	token2, err := client.Get(t.Name())
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Delete the token
	err = client.Delete(t.Name())
	assert.NoError(err)

	// Check tokens
	assert.Equal(token.Name, token2.Name)
}

func Test_client_004(t *testing.T) {
	assert := assert.New(t)
	opts := []client.ClientOpt{
		client.OptTrace(os.Stderr, true),
	}
	if token := GetToken(t); token != "" {
		opts = append(opts, client.OptReqToken(client.Token{
			Scheme: "Bearer",
			Value:  token,
		}))
	}
	client, err := auth.New(GetEndpoint(t), opts...)
	assert.NoError(err)

	// Create a new token with scopes "a", "b" and "c" which expires in 1 minute
	token, err := client.Create(t.Name(), time.Minute, "a", "b", "c")
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Get the token
	token2, err := client.Get(t.Name())
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Check the scopes
	assert.ElementsMatch(token.Scope, token2.Scope)
	assert.Equal(token.Expire, token2.Expire)

	// Delete the token
	err = client.Delete(t.Name())
	assert.NoError(err)

	t.Log(token, token2)
}

///////////////////////////////////////////////////////////////////////////////
// ENVIRONMENT

func GetEndpoint(t *testing.T) string {
	key := os.Getenv("AUTH_ENDPOINT")
	if key == "" {
		t.Skip("AUTH_ENDPOINT not set")
		t.SkipNow()
	}
	if key[len(key)-1] != '/' {
		key += "/"
	}
	return key
}

func GetToken(t *testing.T) string {
	return os.Getenv("AUTH_TOKEN")
}
