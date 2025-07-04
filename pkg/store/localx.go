package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgerrcode"

	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/util/command"
	"github.com/kqlite/kqlite/pkg/util/pgerror"
)

// Execute queries in a connection context, stores the current transaction for the given the client connection.
type LocalQueryExecutor struct {
	tx    *sql.Tx
	dbase *db.Database
}

type execerQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Checks if statemt must return rows.
func StmtReturnsRows(stmt Statement) bool {
	if stmt.ReturnsRows || stmt.CmdType == command.SELECT {
		return true
	}
	return false
}

// Handle transaction commands separately from the other common SQL commands.
func (executor *LocalQueryExecutor) handleTransaction(ctx context.Context, stmt Statement) (bool, error) {
	var err error
	var handled bool

	switch stmt.CmdType {
	case command.BEGIN: // Check for transaction start
		handled = true
		if executor.tx != nil {
			return handled, fmt.Errorf("Syntax error, an active transaction is already present")
		}
		if executor.tx, err = executor.dbase.BeginTx(ctx, nil); err != nil {
			return handled, err
		}
		break

	case command.COMMIT: // Check for transaction end
		handled = true
		if executor.tx == nil {
			return handled, fmt.Errorf("No active transaction to COMMIT")
		}

		if err := executor.tx.Commit(); err != nil {
			return handled, err
		}
		executor.tx = nil
		break

	case command.ROLLBACK: // Check for transaction abort/rollback
		handled = true
		if executor.tx == nil {
			return handled, fmt.Errorf("No active transaction to ROLLBACK")
		}
		if err := executor.tx.Rollback(); err != nil {
			return handled, err
		}
		executor.tx = nil
		break
	}

	return handled, err
}

// Get a ExecuteQueryContext for a local database.
func CreateLocalExecutor(dbase *db.Database) *LocalQueryExecutor {
	return &LocalQueryExecutor{
		tx:    nil,
		dbase: dbase,
	}
}

// Request execution of on or more SQL statements that can contain both executes(transactions) and queries returning rows.
func (executor *LocalQueryExecutor) Request(ctx context.Context, statements []Statement) ([]QueryResponse, error) {
	var err error
	var response []QueryResponse

	// abortOnError indicates whether the caller should continue
	// processing or break.
	abortOnError := func(err error) bool {
		if err != nil && executor.tx != nil {
			executor.tx.Rollback()
			executor.tx = nil
			return true
		}
		return false
	}

	// point executor to default DB connection
	eq := execerQueryer(executor.dbase)

	if executor.tx != nil {
		eq = executor.tx
	}

	for _, stmt := range statements {
		if handled, err := executor.handleTransaction(ctx, stmt); handled || err != nil {
			response = append(response, QueryResponse{
				Error:      err,
				CommandTag: fmt.Sprintf("%s", string(stmt.CmdType)),
			})
			return response, err
		}

		// Check if statement must return rows
		returnsRows := StmtReturnsRows(stmt)

		if returnsRows {
			rows, err := eq.QueryContext(ctx, stmt.Query, stmt.Parameters...)
			if err != nil {
				fmt.Printf("Error from query %s", err.Error())
			}

			response = append(response, QueryResponse{
				Error:      err,
				Rows:       rows,
				CommandTag: fmt.Sprintf("%s 1", stmt.Query),
			})

			if abortOnError(err) {
				break
			}
		} else {
			result, err := eq.ExecContext(ctx, stmt.Query, stmt.Parameters...)

			if err != nil {
				fmt.Printf("Error from query %s", err.Error())
				// TODO SQLITE_CONSTRAINT_UNIQUE
				if stmt.CmdType == command.INSERT {
					err = pgerror.New(pgerrcode.UniqueViolation, err.Error())
				}
			}

			var rowsAffected int64
			if result != nil {
				rowsAffected, _ = result.RowsAffected()
			}

			response = append(response, QueryResponse{
				Error:      err,
				CommandTag: fmt.Sprintf("%s, %d", string(stmt.CmdType), rowsAffected),
			})

			if abortOnError(err) {
				break
			}
		}
	}
	return response, err
}
