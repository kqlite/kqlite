package store

import (
	"database/sql"

	"github.com/kqlite/kqlite/pkg/util/command"
)

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

// Common interface for executing statements with different database Store types (Local database, Replicated/local database, Reflected/local database ..).
type ExecQueryContext interface {
	// Request execution of on or more SQL statements that can contain both executes/transactions and queries returning rows.
	Request(statements []Statement) ([]QueryResponse, error)
}
