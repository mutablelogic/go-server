package httprouter

import "net/http"

////////////////////////////////////////////////////////////////////////////////
// TYPES

// HTTPMiddlewareFunc is a function that wraps an [http.HandlerFunc], returning
// a new handler that may run logic before and/or after calling the next handler.
type HTTPMiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// middlewareFuncs is an ordered slice of [HTTPMiddlewareFunc]. The first element
// is the outermost wrapper and executes first when a request arrives.
type middlewareFuncs []HTTPMiddlewareFunc

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Wrap applies the middleware chain to handler and returns the resulting
// [http.HandlerFunc]. The slice is applied in reverse so that the first
// element becomes the outermost wrapper. If the slice is empty the original
// handler is returned unchanged.
func (w middlewareFuncs) Wrap(handler http.HandlerFunc) http.HandlerFunc {
	if len(w) == 0 {
		return handler
	}
	for i := len(w) - 1; i >= 0; i-- {
		handler = w[i](handler)
	}
	return handler
}
