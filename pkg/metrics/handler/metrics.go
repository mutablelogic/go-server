package handler

import (
	"context"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	metrics "github.com/mutablelogic/go-server/pkg/metrics"
	schema "github.com/mutablelogic/go-server/pkg/metrics/schema"
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
			_ = writeMetrics(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

func writeMetrics(w http.ResponseWriter, _ *http.Request, manager *metrics.MetricManager) error {
	httpresponse.Write(w, http.StatusOK, schema.ContentTypeMetrics)
	return manager.Write(w)
}
