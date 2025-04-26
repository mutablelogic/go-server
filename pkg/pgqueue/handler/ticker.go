package handler

import (
	"context"
	"errors"
	"net/http"

	// Packages
	server "github.com/mutablelogic/go-server"
	httprequest "github.com/mutablelogic/go-server/pkg/httprequest"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	pgqueue "github.com/mutablelogic/go-server/pkg/pgqueue"
	schema "github.com/mutablelogic/go-server/pkg/pgqueue/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func registerTicker(ctx context.Context, router server.HTTPRouter, prefix string, manager *pgqueue.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "ticker"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet)

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = tickerList(w, r, manager)
		case http.MethodPost:
			_ = tickerCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// List all tickers in all namespaces
func tickerList(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager) error {
	// Parse request
	var req schema.TickerListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the tickers
	response, err := manager.ListTickers(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func tickerCreate(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager) error {
	// Parse request
	var req schema.TickerMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// If the ticker already exists, return an error
	if _, err := manager.GetTicker(r.Context(), req.Ticker); err == nil {
		return httpresponse.Error(w, httpresponse.ErrConflict.With("ticker already exists"), req.Ticker)
	} else if !errors.Is(err, httpresponse.ErrNotFound) {
		return httpresponse.Error(w, err)
	}

	// Register the ticker
	response, err := manager.RegisterTicker(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}
