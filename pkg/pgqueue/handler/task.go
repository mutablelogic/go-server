package handler

import (
	"context"
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

func registerTask(ctx context.Context, router server.HTTPRouter, prefix string, manager *pgqueue.Manager) {
	router.HandleFunc(ctx, types.JoinPath(prefix, "task"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet)

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = taskRetain(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	router.HandleFunc(ctx, types.JoinPath(prefix, "task/{queue}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodPost)

		// Check queue parameter
		queue := r.PathValue("queue")
		if queue == "" || !types.IsIdentifier(queue) {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing or invalid queue name"), queue)
			return
		}
		_, err := manager.GetQueue(r.Context(), queue)
		if err != nil {
			_ = httpresponse.Error(w, err, queue)
			return
		}

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodPost:
			_ = taskCreate(w, r, manager, queue)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func taskCreate(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager, queue string) error {
	// Parse request
	var req schema.TaskMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the task
	task, err := manager.CreateTask(r.Context(), queue, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), task)
}

func taskRetain(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager) error {
	// Parse request
	var req schema.TaskMeta
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the task
	task, err := manager.CreateTask(r.Context(), queue, req)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), task)
}
