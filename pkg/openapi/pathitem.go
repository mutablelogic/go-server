package openapi

import (
	"net/http"

	// Packages
	"github.com/mutablelogic/go-server/pkg/openapi/schema"
)

// Operations calls fn for each non-nil operation in the PathItem,
// passing the HTTP method name and a pointer to the operation.
func Operations(pathitem *schema.PathItem, fn func(method string, op *schema.Operation)) {
	for _, entry := range []struct {
		method string
		op     *schema.Operation
	}{
		{http.MethodGet, pathitem.Get},
		{http.MethodHead, pathitem.Head},
		{http.MethodPost, pathitem.Post},
		{http.MethodPut, pathitem.Put},
		{http.MethodPatch, pathitem.Patch},
		{http.MethodDelete, pathitem.Delete},
		{http.MethodOptions, pathitem.Options},
		{http.MethodTrace, pathitem.Trace},
	} {
		if entry.op != nil {
			fn(entry.method, entry.op)
		}
	}
}
