package auth

import (
	"net/http"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MIDDLEWARE

func (manager *Manager) HandleFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Check authentication and authorization, return if either of these is
		//   not valid
		// TODO: If authenticated, add the user to the context
		next(w, r)
	}
}
