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
type DBQueryContext struct {
	db  *db.DB
	tx  *sql.Tx
	ctx context.Context
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
func (dbctx *DBQueryContext) handleTransaction(stmt Statement) (bool, error) {
	var err error
	var handled bool

	switch stmt.CmdType {
	case command.BEGIN: // Check for transaction start
		handled = true
		if dbctx.tx != nil {
			return handled, fmt.Errorf("Syntax error, an active transaction is already present")
		}
		if dbctx.tx, err = dbctx.db.BeginTx(dbctx.ctx, nil); err != nil {
			return handled, err
		}
		break

	case command.COMMIT: // Check for transaction end
		handled = true
		if dbctx.tx == nil {
			return handled, fmt.Errorf("No active transaction to COMMIT")
		}

		if err := dbctx.tx.Commit(); err != nil {
			return handled, err
		}
		dbctx.tx = nil
		break

	case command.ROLLBACK: // Check for transaction abort/rollback
		handled = true
		if dbctx.tx == nil {
			return handled, fmt.Errorf("No active transaction to ROLLBACK")
		}
		if err := dbctx.tx.Rollback(); err != nil {
			return handled, err
		}
		dbctx.tx = nil
		break
	}

	return handled, err
}

// Get a ExecuteQueryContext for a local database.
func CreateDBContext(ctx context.Context, db *db.DB) *DBQueryContext {
	return &DBQueryContext{
		db:  db,
		ctx: ctx,
		tx:  nil,
	}
}

// Request execution of on or more SQL statements that can contain both executes(transactions) and queries returning rows.
func (dbctx *DBQueryContext) Request(statements []Statement) ([]QueryResponse, error) {
	var err error
	var response []QueryResponse

	// abortOnError indicates whether the caller should continue
	// processing or break.
	abortOnError := func(err error) bool {
		if err != nil && dbctx.tx != nil {
			dbctx.tx.Rollback()
			dbctx.tx = nil
			return true
		}
		return false
	}

	// point executor to default DB connection
	eq := execerQueryer(dbctx.db)

	if dbctx.tx != nil {
		eq = dbctx.tx
	}

	for _, stmt := range statements {
		if handled, err := dbctx.handleTransaction(stmt); handled || err != nil {
			response = append(response, QueryResponse{
				Error:      err,
				CommandTag: fmt.Sprintf("%s", string(stmt.CmdType)),
			})
			return response, err
		}

		// Check if statement must return rows
		returnsRows := StmtReturnsRows(stmt)

		if returnsRows {
			rows, err := eq.QueryContext(dbctx.ctx, stmt.Query, stmt.Parameters...)
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
			result, err := eq.ExecContext(dbctx.ctx, stmt.Query, stmt.Parameters...)

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
