package store

import (
	"context"
	"database/sql"

	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/util/command"
)

// Store is a SQLite replicated database instance.
type Store struct {
	dbase      *db.Database         // Local sqlite database instance.
	localExec  *LocalQueryExecutor  // Local query executor/processor, stores the current session transaction.
	remoteExec *RemoteQueryExecutor // Remote query executor/processor, stores session transaction.
}

// Common query response for queries and executes.
type QueryResponse struct {
	Rows       *sql.Rows
	CommandTag string
	Error      error
}

// Represents a single SQL statement.
type Statement struct {
	// SQL Query text
	Query string

	// SQL Command type (ex. SELECT, INSERT, UPDATE ...)
	CmdType command.SQLCommandType

	// Statement parameter values if any.
	Parameters []any

	// Indicates whether statement returns rows even in case of INSERT, UPDATE or others ..
	ReturnsRows bool
}

func Open(dbconf DBConfig) (*Store, error) {
	var err error
	var dbase *db.Database

	// Open connection to SQLite database.
	if dbase, err = db.Open(dbconf.OnDiskPath, dbconf.FKConstraints, dbconf.WalEnabled); err != nil {
		return nil, err
	}
	return &Store{dbase: dbase}, nil
}

func (s *Store) Close() {
	s.dbase.Close()
}

func (s *Store) GetDatabase() *db.Database {
	return s.dbase
}

func (s *Store) Request(ctx context.Context, statements []Statement) ([]QueryResponse, error) {
	if s.localExec == nil {
		s.localExec = CreateLocalExecutor(s.dbase)
	}
	return s.localExec.Request(ctx, statements)
}
