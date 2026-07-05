package httphandler

import (
	_ "embed"
	"errors"
	"io"
	"net/http"

	// Packages
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	jsonschema "github.com/mutablelogic/go-server/pkg/jsonschema"
	openapi "github.com/mutablelogic/go-server/pkg/openapi"
	openapischema "github.com/mutablelogic/go-server/pkg/openapi/schema"
	static "github.com/mutablelogic/go-server/pkg/openapi/static"
	types "github.com/mutablelogic/go-server/pkg/types"
	yaml "gopkg.in/yaml.v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

const (
	pathJSON = "openapi.json"
	pathYAML = "openapi.yaml"
	pathHTML = "openapi.html"
)

//go:embed README.md
var readme []byte

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// RegisterHandler registers GET handlers that serve the router's OpenAPI
// specification at two paths relative to the router's prefix:
//   - openapi.json — JSON (application/json)
//   - openapi.yaml — YAML (application/yaml)
//   - openapi.html — HTML documentation (text/html)
func RegisterHandler(router *httprouter.Router) error {
	var documentation = openapi.ParseMarkdown(readme)
	router.Spec().AddTag("OpenAPI", documentation.Section(1, "OpenAPI Operations").Body)
	return errors.Join(
		router.Register(pathJSON, nil, func(path httprequest.PathItem) {
			path.Tag("OpenAPI")
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), router.Spec())
			}, func(op httprequest.PathOperation) {
				op.Summary("Return JSON Specification")
				op.Description(documentation.Section(2, "GET /openapi.json").Body)
				op.JSONResponse(http.StatusOK, jsonschema.MustFor[openapischema.Spec]())
			})
		}),
		router.Register(pathYAML, nil, func(path httprequest.PathItem) {
			path.Tag("OpenAPI")
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				data, err := yaml.Marshal(router.Spec())
				if err != nil {
					_ = httpresponse.Error(w, err)
					return
				}
				_ = httpresponse.Write(w, http.StatusOK, types.ContentTypeYAML, func(out io.Writer) (int, error) {
					return out.Write(data)
				})
			}, func(op httprequest.PathOperation) {
				op.Summary("Return YAML Specification")
				op.Description(documentation.Section(2, "GET /openapi.yaml").Body)
				op.Response(http.StatusOK, types.ContentTypeYAML, "OpenAPI specification in YAML format")
			})
		}),
		router.Register(pathHTML, nil, func(path httprequest.PathItem) {
			path.Tag("OpenAPI")
			path.Get(func(w http.ResponseWriter, r *http.Request) {
				_ = httpresponse.Write(w, http.StatusOK, types.ContentTypeHTML+"; charset=utf-8", func(out io.Writer) (int, error) {
					return out.Write(static.OpenAPIHTML)
				})
			}, func(op httprequest.PathOperation) {
				op.Summary("Return HTML Documentation")
				op.Description(documentation.Section(2, "GET /openapi.html").Body)
				op.Response(http.StatusOK, types.ContentTypeHTML, "OpenAPI specification in HTML format")
			})
		}),
	)
}
