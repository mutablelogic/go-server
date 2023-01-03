package context

import "net/http"

// RequireScope returns a http.NotAutorized error if the request does not
// contain the required scope and the enforced flag is true
func RequireScope(handler http.HandlerFunc, enforced bool, scope ...string) http.HandlerFunc {
	if len(scope) == 0 || !enforced {
		return handler
	}
	// TODO
	return nil
}
