package indexer

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	// Modules
	. "github.com/djthorpe/go-server"
	sq "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"
	sqobj "github.com/djthorpe/go-sqlite/pkg/sqobj"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Database struct {
	sq.SQConnection
	schema string
}

type File struct {
	Id      string    `sqlite:"id,primary"`
	Name    string    `sqlite:"name,not null,index:doc_name"`
	Path    string    `sqlite:"path,not null,unique:doc_path"`
	Ext     string    `sqlite:"ext,index:doc_ext"`
	ModTime time.Time `sqlite:"modtime,index:doc_modtime"`
	Size    int64     `sqlite:"size,not null"`
	Mark    bool      `sqlite:"mark,not null"`
}

type Doc struct {
	File        int64  `sqlite:"file,primary"`
	Title       string `sqlite:"title,not null"`
	Description string `sqlite:"description"`
	Shortform   string `sqlite:"shortform"`
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	tableNameFile = "file"
	tableNameDoc  = "doc"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewDatabase(conn sq.SQConnection, schema string) (*Database, error) {
	this := new(Database)

	// Check arguments
	if conn == nil || schema == "" {
		return nil, ErrBadParameter.With("Invalid connnection or schema")
	} else if schemas := conn.Schemas(); arrayContains(schemas, schema) == false {
		return nil, ErrNotFound.With("schema: ", schema)
	} else {
		this.SQConnection = conn
		this.schema = schema
	}

	// Create schema
	if err := this.createSchema(); err != nil {
		return nil, err
	}

	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Database) String() string {
	str := "<indexer.database"
	str += fmt.Sprintf("  schema=%q", this.schema)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func (this *Database) MarkDocuments() error {
	return ErrNotImplemented
}

// AddFile adds a file row into the database, and return primary key for adding
// in a document and metadata in a second step
func (this *Database) AddFile(key, path string, info fs.FileInfo) (int64, error) {
	f := &File{
		Id:   key,
		Path: filepath.Dir(path),
		Name: filepath.Base(path),
	}
	f.Ext = filepath.Ext(f.Name)
	if info != nil {
		f.ModTime = info.ModTime()
		f.Size = info.Size()
	}
	var id int64
	if err := this.Do(func(txn sq.SQTransaction) error {
		if params, err := sqobj.InsertParams(f); err != nil {
			return err
		} else if r, err := txn.Exec(sqobj.InsertRow(tableNameFile, f), params...); err != nil {
			return err
		} else {
			id = r.LastInsertId
		}
		// Return success
		return nil
	}); err != nil {
		return 0, err
	} else {
		return id, nil
	}
}

func (this *Database) RemoveDocumentWithKey(key string) error {
	return ErrNotImplemented
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func arrayContains(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

func (this *Database) createSchema() error {
	source := N(tableNameFile).WithSchema(this.schema)
	return this.Do(func(txn sq.SQTransaction) error {
		for _, statement := range sqobj.CreateTableAndIndexes(source, true, File{}) {
			if _, err := txn.Exec(statement); err != nil {
				return err
			}
		}
		// Return success
		return nil
	})
}
