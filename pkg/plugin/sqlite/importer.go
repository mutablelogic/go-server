package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	// Modules
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	"github.com/djthorpe/go-sqlite/pkg/sqobj"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ImportStatus string

type ImportJob struct {
	Id      int64         `sqlite:"config,primary,references:import_config"`
	Url     string        `sqlite:"url,notnull"`
	Created time.Time     `sqlite:"created,notnull"`
	Updated time.Time     `sqlite:"updated"`
	Status  *ImportStatus `sqlite:"status"`
	Reason  *string       `sqlite:"reason"`
}

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	ImportStatusRunning ImportStatus = "running"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *sq) CreateImportTables(schema string) error {
	if _, err := this.db.Exec(sqobj.CreateTable("import_config", SQImportConfig{}).IfNotExists().WithTemporary()); err != nil {
		return err
	}
	if _, err := this.db.Exec(sqobj.CreateTable("import_job", ImportJob{}).IfNotExists().WithTemporary()); err != nil {
		return err
	}

	// Return success
	return nil
}

func (this *sq) AddImport(url *url.URL, cfg SQImportConfig) (int64, error) {
	job := ImportJob{
		Url:     url.String(),
		Created: time.Now(),
	}
	// Insert import configuration into database
	if params, err := sqobj.InsertParams(cfg); err != nil {
		return 0, err
	} else if r, err := this.db.Exec(sqobj.InsertRow("import_config", cfg), params...); err != nil {
		return 0, err
	} else if r.RowsAffected != 1 {
		return 0, ErrInternalAppError
	} else {
		job.Id = r.LastInsertId
	}
	// Insert job into the database
	if params, err := sqobj.InsertParams(job); err != nil {
		return 0, err
	} else if r, err := this.db.Exec(sqobj.InsertRow("import_job", job), params...); err != nil {
		return 0, err
	} else if r.RowsAffected != 1 {
		return 0, ErrInternalAppError
	}

	// Return success
	return job.Id, nil
}

func (this *sq) GetImportJob() error {
	s := S(N("import_job")).
		WithLimitOffset(1, 0).
		Order(N("created"))
	rows, err := this.db.Query(s)
	if err != nil {
		return err
	}
	defer rows.Close()
	var job ImportJob
	for {
		if err := rows.Next(&job); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}
		fmt.Println(job)
	}

	// Return success
	return nil
}
