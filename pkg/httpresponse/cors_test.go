package httpresponse_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/types"
	"github.com/stretchr/testify/assert"
)

func Test_Cors_001(t *testing.T) {
	assert := assert.New(t)

	t.Run("Origin", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "http://example.com", nil)
		resp := httptest.NewRecorder()
		httpresponse.Cors(resp, req, "")

		assert.Equal(http.StatusOK, resp.Code)
		assert.Equal("*", resp.Header().Get("Access-Control-Allow-Origin"))
	})
	t.Run("Methods", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "http://example.com", nil)
		resp := httptest.NewRecorder()
		httpresponse.Cors(resp, req, "http://example.com", http.MethodGet, http.MethodPost)

		assert.Equal(http.StatusOK, resp.Code)
		assert.Equal("http://example.com", resp.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal("GET, POST", resp.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal("*", resp.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("Headers", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "http://example.com", nil)
		resp := httptest.NewRecorder()
		httpresponse.Cors(resp, req, "", types.ContentAcceptHeader, types.ContentTypeHeader)

		assert.Equal("*", resp.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal("Accept, Content-Type", resp.Header().Get("Access-Control-Allow-Headers"))
	})
}
