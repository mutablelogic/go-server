package provider_test

import (
	"os"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/handler/auth"
	"github.com/mutablelogic/go-server/pkg/handler/logger"
	"github.com/mutablelogic/go-server/pkg/handler/nginx"
	"github.com/mutablelogic/go-server/pkg/handler/router"
	"github.com/mutablelogic/go-server/pkg/handler/tokenjar"
	"github.com/mutablelogic/go-server/pkg/httpserver"
	"github.com/mutablelogic/go-server/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_parser_001(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../etc/json/nginx-proxy.json")
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	// Make a parser
	parser, err := provider.NewParser(
		logger.Config{}, nginx.Config{}, tokenjar.Config{}, auth.Config{}, router.Config{}, httpserver.Config{},
	)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Parse JSON file
	if err := parser.ParseJSON(r); !assert.NoError(err) {
		t.SkipNow()
	}

	// Bind
	if err := parser.Bind(); !assert.NoError(err) {
		t.SkipNow()
	}
}
