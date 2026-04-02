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
//   - openapi.json — JSON (application/json)
//   - openapi.yaml — YAML (application/yaml)
//   - openapi.html — HTML documentation (text/html)
func RegisterHandler(router *httprouter.Router) error {
	if err := router.RegisterPath(pathJSON, nil,
		httprequest.NewPathItem("OpenAPI JSON", "Serve the OpenAPI specification as JSON").Get(jsonHandler(router), "Get OpenAPI JSON"),
	); err != nil {
		return err
	}
	if err := router.RegisterPath(pathYAML, nil,
		httprequest.NewPathItem("OpenAPI YAML", "Serve the OpenAPI specification as YAML").Get(yamlHandler(router), "Get OpenAPI YAML"),
	); err != nil {
		return err
	}
	return router.RegisterPath(pathHTML, nil,
		httprequest.NewPathItem("OpenAPI HTML", "Serve the OpenAPI documentation UI").Get(htmlHandler(), "Get OpenAPI HTML"),
	)
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
		_ = httpresponse.Write(w, http.StatusOK, types.ContentTypeHTML+"; charset=utf-8", func(out io.Writer) (int, error) {
			return out.Write(static.OpenAPIHTML)
		})
	}
}
