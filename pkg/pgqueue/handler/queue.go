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

func registerQueue(ctx context.Context, router server.HTTPRouter, prefix string, manager *pgqueue.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "queue"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = queueList(w, r, manager)
		case http.MethodPost:
			_ = queueCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "queue/{name}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodDelete, http.MethodPatch)

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = queueGet(w, r, manager, r.PathValue("name"))
		case http.MethodDelete:
			_ = queueDelete(w, r, manager, r.PathValue("name"))
		case http.MethodPatch:
			_ = queueUpdate(w, r, manager, r.PathValue("name"))
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func queueList(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager) error {
	// Parse request
	var req schema.QueueListRequest
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// List the queues
	response, err := manager.ListQueues(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), response)
}

func queueCreate(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager) error {
	// Parse request
	var req schema.QueueMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// If the queue already exists, return an error
	if _, err := manager.GetQueue(r.Context(), req.Queue); err == nil {
		return httpresponse.Error(w, httpresponse.ErrConflict.With("queue already exists"), req.Queue)
	} else if !errors.Is(err, httpresponse.ErrNotFound) {
		return httpresponse.Error(w, err)
	}

	// Register the queue
	response, err := manager.RegisterQueue(r.Context(), req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), response)
}

func queueGet(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager, name string) error {
	queue, err := manager.GetQueue(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), queue)
}

func queueDelete(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager, name string) error {
	queue, err := manager.DeleteQueue(r.Context(), name)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), queue)
}

func queueUpdate(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager, name string) error {
	// Parse request
	var meta schema.QueueMeta
	if err := httprequest.Read(r, &meta); err != nil {
		return httpresponse.Error(w, err)
	}

	// Perform update
	queue, err := manager.UpdateQueue(r.Context(), name, meta)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusOK, httprequest.Indent(r), queue)
}
