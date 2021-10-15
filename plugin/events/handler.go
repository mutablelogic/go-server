package main

import (
	"context"
	"net/http"
	"regexp"
	"time"

	// Modules
	router "github.com/mutablelogic/go-server/pkg/httprouter"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type EventQuery struct {
	Limit int `json:"limit"`
}

type EventResponse struct {
	Count int64       `json:"id"`
	Time  time.Time   `json:"ts"`
	Name  string      `json:"name"`
	Value interface{} `json:"value,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRouteEvents = regexp.MustCompile(`^/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (p *plugin) AddHandlers(ctx context.Context, provider Provider) error {
	// Add handler for returning latest n events
	if err := provider.AddHandlerFuncEx(ctx, reRouteEvents, p.ServeEvents); err != nil {
		provider.Print(ctx, "Failed to add handler: ", err)
		return nil
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (p *plugin) ServeEvents(w http.ResponseWriter, req *http.Request) {
	// Decode query
	var q EventQuery
	if err := router.RequestQuery(req, &q); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if q.Limit <= 0 {
		q.Limit = p.Capacity
	} else {
		q.Limit = minInt(q.Limit, p.Capacity)
	}

	// Output events in reverse order, latest goes first
	p.Lock()
	response := []EventResponse{}
	for i := len(p.E) - 1; i >= 0; i-- {
		e := p.E[i]
		response = append(response, EventResponse{
			Count: e.count,
			Time:  e.ts,
			Name:  e.Name(),
			Value: e.Value(),
		})
		if len(response) >= q.Limit {
			break
		}
	}
	p.Unlock()

	// Return response
	router.ServeJSON(w, response, http.StatusOK, 2)
}
