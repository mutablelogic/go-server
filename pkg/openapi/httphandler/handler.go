// Package httphandler provides HTTP handler functions for an [httprouter.Router].
package httphandler

import (
	"io"
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	static "github.com/mutablelogic/go-server/pkg/openapi/static"
	types "github.com/mutablelogic/go-server/pkg/types"
	yaml "gopkg.in/yaml.v3"
)

const (
	pathJSON = "openapi.json"
	pathYAML = "openapi.yaml"
	pathHTML = "openapi.html"
)

// RegisterHandler registers GET handlers that serve the router's OpenAPI
// specification at two paths relative to the router's prefix:
//
//   - openapi.json — JSON (application/json)
//   - openapi.yaml — YAML (application/yaml)
//   - openapi.html — HTML documentation (text/html)
//
// All other HTTP methods return 405 Method Not Allowed. When middleware is
// true the handlers are wrapped by the router's middleware chain.
func RegisterHandler(router *httprouter.Router, middleware bool) error {
	if err := router.RegisterFunc(pathJSON, jsonHandler(router), middleware, nil); err != nil {
		return err
	}
	if err := router.RegisterFunc(pathYAML, yamlHandler(router), middleware, nil); err != nil {
		return err
	}
	return router.RegisterFunc(pathHTML, htmlHandler(), middleware, nil)
}

func jsonHandler(router *httprouter.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
			return
		}
		_ = httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), router.Spec())
	}
}

func yamlHandler(router *httprouter.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
			return
		}
		data, err := yaml.Marshal(router.Spec())
		if err != nil {
			_ = httpresponse.Error(w, err)
			return
		}
		_ = httpresponse.Write(w, http.StatusOK, types.ContentTypeYAML, func(out io.Writer) (int, error) {
			return out.Write(data)
		})
	}
}

func htmlHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
			return
		}
		_ = httpresponse.Write(w, http.StatusOK, types.ContentTypeHTML, func(out io.Writer) (int, error) {
			return out.Write(static.OpenAPIHTML)
		})
	}
}
