package handler

import (
	"context"
	"net/http"
	"strconv"

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
	// GET /task gets retains a task from any queue
	// POST /task creates a new task
	router.HandleFunc(ctx, types.JoinPath(prefix, "task"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodGet, http.MethodPost)

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodGet:
			_ = taskRetain(w, r, manager)
		case http.MethodPost:
			_ = taskCreate(w, r, manager)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})

	// PATCH /task/{id} marks a task as succeeded or failed with a payload
	// DELETE /task/{id} marks a task as succeeded
	router.HandleFunc(ctx, types.JoinPath(prefix, "task/{id}"), func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		httpresponse.Cors(w, r, router.Origin(), http.MethodPatch, http.MethodDelete)

		// Check id parameter
		id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			_ = httpresponse.Error(w, httpresponse.ErrBadRequest.With("missing or invalid task id"), r.PathValue("id"))
			return
		}

		switch r.Method {
		case http.MethodOptions:
			_ = httpresponse.Empty(w, http.StatusOK)
		case http.MethodPatch:
			_ = taskRelease(w, r, manager, id)
		case http.MethodDelete:
			_ = taskRelease(w, r, manager, id)
		default:
			_ = httpresponse.Error(w, httpresponse.Err(http.StatusMethodNotAllowed), r.Method)
		}
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func taskCreate(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager) error {
	// Parse request
	var req struct {
		Queue string `json:"queue"`
		schema.TaskMeta
	}
	if err := httprequest.Read(r, &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Check queue name
	queue, err := manager.GetQueue(r.Context(), req.Queue)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Create the task
	task, err := manager.CreateTask(r.Context(), queue.Queue, req.TaskMeta)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), task)
}

func taskRetain(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager) error {
	// Parse request
	var req struct {
		Worker string `json:"worker"`
	}
	if err := httprequest.Query(r.URL.Query(), &req); err != nil {
		return httpresponse.Error(w, err)
	}

	// Retain the task
	task, err := manager.NextTask(r.Context(), pgqueue.OptWorker(req.Worker))
	if err != nil {
		return httpresponse.Error(w, err)
	} else if task == nil {
		return httpresponse.Error(w, httpresponse.ErrNotFound.With("no task available"))
	}

	// Return success
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), task)
}

func taskRelease(w http.ResponseWriter, r *http.Request, manager *pgqueue.Manager, id uint64) error {
	// Only parse request if method is PATCH
	var req struct {
		Result any `json:"result,omitempty"`
	}
	if r.Method == http.MethodPatch {
		if err := httprequest.Read(r, &req); err != nil {
			return httpresponse.Error(w, err)
		}
	}

	// Release the task - success if no result
	var status string
	task, err := manager.ReleaseTask(r.Context(), id, req.Result == nil, req.Result, &status)
	if err != nil {
		return httpresponse.Error(w, err)
	}

	// Return success - embed status
	return httpresponse.JSON(w, http.StatusCreated, httprequest.Indent(r), schema.TaskWithStatus{
		Task:   *task,
		Status: status,
	})
}
