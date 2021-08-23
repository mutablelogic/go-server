package main

import (
	"io"
	"net/http"
	"net/url"
	"regexp"

	// Modules
	router "github.com/djthorpe/go-server/pkg/httprouter"
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	sqimport "github.com/djthorpe/go-sqlite/pkg/sqimport"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PingResponse struct {
	Version string   `json:"version"`
	Modules []string `json:"modules"`
	Schemas []string `json:"schemas"`
}

type SchemaObjectResponse struct {
	Schema string      `json:"schema"`
	Name   interface{} `json:"name"`
	Table  interface{} `json:"table,omitempty"`
	Type   interface{} `json:"type"`
	Sql    interface{} `json:"sql,omitempty"`
	Count  int64       `json:"count,omitempty"`
}

type TableQuery struct {
	Offset uint `json:"offset"`
	Limit  uint `json:"limit"`
}

type TableResponse struct {
	Schema string                 `json:"schema"`
	Name   interface{}            `json:"name"`
	Offset uint                   `json:"offset,omitempty"`
	Limit  uint                   `json:"limit,omitempty"`
	Count  int64                  `json:"count"`
	Sql    string                 `json:"sql,omitempty"`
	Cols   []SchemaColumnResponse `json:"columns,omitempty"`
	Rows   [][]interface{}        `json:"result"`
}

type ImportRequest struct {
	Url  string `json:"url"`
	Name string `json:"name,omitempty"`
}

type ImportResponse struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
	Url    string `json:"url,omitempty"`
}

type SchemaColumnResponse struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRoutePing   = regexp.MustCompile(`^/?$`)
	reRouteSchema = regexp.MustCompile(`^/(\w+)/?$`)
	reRouteTable  = regexp.MustCompile(`^/(\w+)/([^/]+)/?$`)
	reRouteImport = regexp.MustCompile(`^/(\w+)/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	maxResultLimit = 1000
)

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (this *sq) ServePing(w http.ResponseWriter, req *http.Request) {
	response := PingResponse{
		Version: Version(),
		Modules: []string{},
		Schemas: []string{},
	}
	if this.db != nil {
		response.Modules = append(response.Modules, this.db.Modules()...)
		response.Schemas = append(response.Schemas, this.db.Schemas()...)
	}
	router.ServeJSON(w, response, http.StatusOK, 0)
}

func (this *sq) ServeSchema(w http.ResponseWriter, req *http.Request) {
	params := router.RequestParams(req)
	response := []*SchemaObjectResponse{}

	// Obtain table information
	rs, err := this.db.Query(Q("SELECT name,type,sql,tbl_name FROM " + QuoteIdentifier(params[0]) + ".sqlite_master"))
	if err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for {
		row := rs.NextArray()
		if row == nil {
			break
		}
		response = append(response, &SchemaObjectResponse{
			Schema: params[0],
			Name:   row[0],
			Type:   row[1],
			Sql:    row[2],
			Table:  row[3],
		})
	}

	// Add row count to table information
	for _, object := range response {
		if object.Type != "table" {
			continue
		}
		if count, err := this.count(object.Schema, object.Name.(string)); err != nil {
			router.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		} else {
			object.Count = count
		}
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

/*
func (this *sq) ServeTable(w http.ResponseWriter, req *http.Request) {
	params := server.RequestParams(req)
	response := TableResponse{
		Schema: params[0],
		Name:   params[1],
	}

	// Decode request, set offset and limit
	q := TableQuery{}
	if err := server.RequestQuery(req, &q); err != nil {
		server.ServeError(w, http.StatusBadRequest, err.Error())
		return
	} else {
		response.Offset = q.Offset
		if q.Limit == 0 {
			response.Limit = maxLimit
		} else {
			response.Limit = minUint(q.Limit, maxLimit)
		}
	}

	// Get row count in response
	if count, err := this.count(params[0], params[1]); err != nil {
		server.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	} else {
		response.Count = count
	}

	// Get columns in response
	cols := this.db.ColumnsEx(params[1], params[0])
	sources := []sqlite.SQSource{}
	if cols == nil {
		server.ServeError(w, http.StatusInternalServerError)
		return
	}
	for _, col := range cols {
		response.Cols = append(response.Cols, SchemaColumnResponse{
			Name:     col.Name(),
			Type:     col.Type(),
			Nullable: col.Nullable(),
		})
		sources = append(sources, N(col.Name()).WithSchema(params[1]))
	}

	// Create Query
	sql := S(N(params[1]).
		WithAlias(params[1]).
		WithSchema(params[0])).
		To(sources...).
		WithLimitOffset(response.Limit, response.Offset)

	// Query table
	rs, err := this.db.Query(sql)
	if err != nil {
		server.ServeError(w, http.StatusInternalServerError, err.Error(), sql.Query())
		return
	} else {
		response.Sql = sql.Query()
	}
	defer rs.Close()
	for {
		row := rs.NextArray()
		if row == nil {
			break
		}
		response.Rows = append(response.Rows, row)
	}

	// Serve response
	server.ServeJSON(w, response, http.StatusOK, 2)
}
*/

func (this *sq) ServeImport(w http.ResponseWriter, req *http.Request) {
	params := router.RequestParams(req)

	// Decode the body of the request
	query := ImportRequest{}
	if err := router.RequestBody(req, &query); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if query.Url == "" {
		router.ServeError(w, http.StatusBadRequest, "Missing url")
		return
	}

	// Parse the query
	url, err := url.Parse(query.Url)
	if err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	} else if url.Scheme != "http" && url.Scheme != "https" {
		router.ServeError(w, http.StatusBadRequest, "Only http and https import schemes are supported")
		return
	}

	// Create the response
	response := ImportResponse{
		Schema: params[0],
		Url:    url.String(),
	}

	// TODO
	cfg := SQImportConfig{
		Schema:     params[0],
		Name:       query.Name,
		Header:     true,
		TrimSpace:  true,
		LazyQuotes: true,
		Overwrite:  true,
	}
	writer, err := sqimport.NewSQLWriter(cfg, this.db)
	if err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}
	importer, err := sqimport.NewImporter(cfg, response.Url, writer)
	if err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Read from source until EOF
	for {
		if err := importer.Read(); err == io.EOF {
			break
		} else if err != nil {
			router.ServeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Set response
	response.Name = importer.Name()

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *sq) count(schema, name string) (int64, error) {
	rs, err := this.db.Query(Q("SELECT COUNT(*) FROM " + QuoteIdentifier(schema) + "." + QuoteIdentifier(name)))
	if err != nil {
		return 0, err
	}
	defer rs.Close()
	row := rs.NextArray()
	if row != nil {
		return row[0].(int64), nil
	} else {
		return 0, ErrInternalAppError
	}
}

func minUint(a, b uint) uint {
	if a < b {
		return a
	} else {
		return b
	}
}
