package handler

import (
	"context"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	"github.com/mutablelogic/go-server/pkg/metrics"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func RegisterMetrics(ctx context.Context, router server.HTTPRouter, prefix string, manager *metrics.MetricManager) {
	router.HandleFunc(ctx, prefix, func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet)

		// Handle method
		switch r.Method {
		case http.MethodGet:
			_ = manager.Write(w)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}
