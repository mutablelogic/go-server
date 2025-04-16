package schema

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	// Packages
	pg "github.com/djthorpe/go-pg"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////////
// TYPES

type TokenStatus string

type TokenId struct {
	User string `json:"user" arg:"" help:"User name"`
	Id   uint64 `json:"id" arg:"" help:"Token ID"`
}

type TokenMeta struct {
	Desc *string `json:"desc,omitempty" help:"Description"`
}

type Token struct {
	TokenId
	Ts     time.Time `json:"timestamp"`
	Status string    `json:"status"`
	TokenMeta
	Value *string `json:"token,omitempty"`
}

type TokenNew struct {
	User string `json:"user"`
	TokenMeta
	Value string `json:"token"`
	Hash  string `json:"hash"`
}

type TokenListRequest struct {
	Status *string `json:"status,omitempty" help:"Status"`
	pg.OffsetLimit
}

type TokenList struct {
	TokenListRequest
	Count uint64  `json:"count"`
	Body  []Token `json:"body,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////////
// TYPES

const (
	TokenStatusLive     = TokenStatus("live")
	TokenStatusArchived = TokenStatus("archived")
)

///////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t TokenMeta) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

func (t Token) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

func (t TokenNew) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

func (t TokenListRequest) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

func (t TokenList) String() string {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// SELECTOR

func (t TokenNew) Select(bind *pg.Bind, op pg.Op) (string, error) {
	switch op {
	case pg.Get:
		return tokenGenerate, nil
	default:
		return "", httpresponse.ErrBadRequest.Withf("TokenNew: operation %q is not supported", op)
	}
}

func (t TokenId) Select(bind *pg.Bind, op pg.Op) (string, error) {
	if user := strings.TrimSpace(t.User); user == "" {
		return "", httpresponse.ErrBadRequest.With("user is missing")
	} else {
		bind.Set("user", user)
	}
	if t.Id == 0 {
		return "", httpresponse.ErrBadRequest.With("id is missing")
	} else {
		bind.Set("id", t.Id)
	}

	switch op {
	case pg.Get:
		return tokenGet, nil
	case pg.Update:
		return tokenUpdate, nil
	case pg.Delete:
		return tokenDelete, nil
	default:
		return "", httpresponse.ErrBadRequest.Withf("TokenId: operation %q is not supported", op)
	}
}

func (list *TokenListRequest) Select(bind *pg.Bind, op pg.Op) (string, error) {
	// Order
	bind.Set("orderby", `ORDER BY ts DESC`)

	// Where
	bind.Del("where")

	// User
	if !bind.Has("user") {
		return "", httpresponse.ErrBadRequest.With("user is missing")
	} else {
		bind.Append("where", `"user" = @user`)
	}

	// Status
	if status := types.TrimStringPtr(list.Status); status == nil {
		list.Status = types.StringPtr(string(TokenStatusLive))
	} else {
		list.Status = status
	}
	if list.Status != nil {
		bind.Append("where", `"status" = `+bind.Set("status", list.Status))
	}

	// Where AND
	if where := bind.Join("where", " AND "); where != "" {
		bind.Set("where", `WHERE `+where)
	} else {
		bind.Set("where", "")
	}

	// Offset and limit
	list.OffsetLimit.Bind(bind, UserListLimit)

	switch op {
	case pg.List:
		return tokenList, nil
	default:
		return "", httpresponse.ErrBadRequest.Withf("TokenListRequest: operation %q is not supported", op)
	}
}

//////////////////////////////////////////////////////////////////////////////
// READER

func (t *Token) Scan(row pg.Row) error {
	return row.Scan(&t.Id, &t.User, &t.Ts, &t.Status, &t.Desc)
}

func (t *TokenNew) Scan(row pg.Row) error {
	return row.Scan(&t.Value, &t.Hash)
}

func (list *TokenList) ScanCount(row pg.Row) error {
	list.Body = []Token{}
	return row.Scan(&list.Count)
}

func (list *TokenList) Scan(row pg.Row) error {
	var token Token
	if err := token.Scan(row); err != nil {
		return err
	}
	list.Body = append(list.Body, token)
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// WRITER

func (t TokenNew) Insert(bind *pg.Bind) (string, error) {
	if user := strings.TrimSpace(t.User); user == "" {
		return "", httpresponse.ErrBadRequest.With("user is missing")
	} else {
		bind.Set("user", user)
	}
	if hash := strings.TrimSpace(t.Hash); hash == "" {
		return "", httpresponse.ErrBadRequest.With("hash is missing")
	} else {
		bind.Set("hash", hash)
	}

	// Set description
	bind.Set("desc", types.TrimStringPtr(t.Desc))

	// Return insert
	return tokenInsert, nil
}

func (t TokenNew) Update(bind *pg.Bind) error {
	return httpresponse.ErrNotImplemented.With("TokenNew: Update not implemented")
}

func (t TokenMeta) Insert(bind *pg.Bind) (string, error) {
	return "", httpresponse.ErrNotImplemented.With("TokenMeta.Insert")
}

func (t TokenMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")
	if t.Desc != nil {
		bind.Append("patch", `"desc" = `+bind.Set("desc", types.TrimStringPtr(t.Desc)))
	}

	// Set patch
	if patch := bind.Join("patch", ", "); patch != "" {
		bind.Set("patch", patch)
	} else {
		return httpresponse.ErrBadRequest.With("no fields to update")
	}

	// Return success
	return nil
}

func (status TokenStatus) Update(bind *pg.Bind) error {
	if status == TokenStatusLive || status == TokenStatusArchived {
		bind.Set("patch", `status = `+bind.Set("status", status))
	} else {
		return httpresponse.ErrBadRequest.Withf("invalid status %q", status)
	}

	// Return success
	return nil
}

func (status TokenStatus) Insert(bind *pg.Bind) (string, error) {
	return "", httpresponse.ErrNotImplemented.With("TokenStatus.Insert")
}

//////////////////////////////////////////////////////////////////////////////
// SQL

// Create objects in the schema
func bootstrapToken(ctx context.Context, conn pg.Conn) error {
	q := []string{
		tokenCreateExtension,
		userCreateStatusType,
		tokenCreateTable,
	}
	for _, query := range q {
		if err := conn.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

const (
	tokenCreateExtension = `
		CREATE EXTENSION IF NOT EXISTS pgcrypto CASCADE
	`
	tokenCreateTable = `
		CREATE TABLE IF NOT EXISTS ${"schema"}."token" (
			"id"           SERIAL PRIMARY KEY,  	
			"user"         TEXT REFERENCES ${"schema"}.user("name") ON DELETE CASCADE,  -- token owner
			"ts"           TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,          -- timestamp of creation or update
			"status"       ${"schema"}.USER_STATUS NOT NULL DEFAULT 'live',             -- status of the token
			"desc"         TEXT,                           			                    -- description of the token
			"hash"		   TEXT NOT NULL,                                               -- hashed token value	
			UNIQUE("hash")		
		)	
	`
	tokenGenerate = `
		WITH generate_token AS (
			SELECT gen_random_uuid()::TEXT AS "token"
		) SELECT
			"token", ENCODE(DIGEST("token", ${'algorithm'}), 'hex') AS "hash"
		FROM
			generate_token
		;
	`
	tokenInsert = `
		INSERT INTO 
			${"schema"}.token ("user", "ts", "status", "desc", "hash") 
		VALUES 
			(@user, DEFAULT, DEFAULT, @desc, @hash)
		RETURNING 
			"id", "user", "ts", "status", "desc"	
	`
	tokenUpdate = `
		UPDATE
			${"schema"}.token
		SET
			${patch}, "ts" = CURRENT_TIMESTAMP
		WHERE
			"user" = @user AND "id" = @id
		RETURNING
			"id", "user", "ts", "status", "desc"
	`
	tokenDelete = `
		DELETE FROM
			${"schema"}.token
		WHERE
			"user" = @user AND "id" = @id
		RETURNING
			"id", "user", "ts", 'deleted' AS "status", "desc"
	`
	tokenSelect = `
		SELECT
			T."id", T."user", T."ts", 
			CASE 
				WHEN U."status" = 'archived' THEN U."status"
				ELSE T."status"
			END AS "status",
			COALESCE(T."desc", U."desc") AS "desc"
		FROM
			${"schema"}."token" T
		JOIN
			${"schema"}."user" U ON T."user" = U."name"
	`
	tokenGet  = tokenSelect + ` WHERE T."user" = @user AND T."id" = @id`
	tokenList = `WITH tq AS (` + tokenSelect + `) SELECT * FROM tq ${where} ${orderby}`
)
