package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"slices"
	"sync"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/kqlite/kqlite/pkg/cluster"
	"github.com/kqlite/kqlite/pkg/connpool"
	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/util/command"
	"github.com/kqlite/kqlite/pkg/util/pgerror"
)

var (
	ErrTxExists   = errors.New("Syntax error, an active transaction is already present")
	ErrNoActiveTx = errors.New("Syntax error, missing active transaction")
	txLock        sync.Mutex
)

// DataStore is an SQLite abstracted replicated database instance with local and remote transaction context.
type DataStore struct {
	dbase      *db.Database // sqlite database instance.
	localTx    *sql.Tx      // Local sqlite transaction.
	remoteTx   pgx.Tx       // Remote replicated transaction over the PostgreSQL wire protocol.
	replicated bool         // Should this db connection be replicated.
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

// Open a sqlite replicated store with options from DBConfig.
func Open(replicated bool, dbconf DBConfig) (*DataStore, error) {
	var err error
	var dbase *db.Database

	// Open connection to SQLite database.
	if dbase, err = db.Open(dbconf.OnDiskPath, dbconf.ReadOnly,
		dbconf.FKConstraints, dbconf.WalEnabled); err != nil {
		return nil, err
	}
	return &DataStore{
		dbase:      dbase,
		localTx:    nil,
		remoteTx:   nil,
		replicated: replicated,
	}, err
}

// Close the replicated sqlite store.
func (ds *DataStore) Close() {
	ds.dbase.Close()
}

// Get store's underlying sqlite database instance.
func (ds *DataStore) GetDatabase() *db.Database {
	return ds.dbase
}

type execQuery interface {
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

// Handle transaction commands separately from the other SQL commands and create transaction context.
func (ds *DataStore) handleRemoteTransaction(ctx context.Context, stmt Statement) (bool, error) {
	var err error
	// Handle start of transaction and create context.
	if stmt.CmdType == command.BEGIN {
		if ds.remoteTx != nil {
			return true, ErrTxExists
		}
		options := pgx.TxOptions{
			BeginQuery: "BEGIN",
		}
		dbname := ds.dbase.GetName()
		log.Printf("Send begin query\n")
		if ds.remoteTx, err = connpool.BeginTx(ctx, dbname, options); err != nil {
			return true, err
		}
		return true, nil
	}
	if stmt.CmdType == command.COMMIT || stmt.CmdType == command.ROLLBACK {
		if ds.remoteTx == nil {
			return true, ErrNoActiveTx
		}
		if stmt.CmdType == command.COMMIT {
			log.Printf("Send commit query\n")
			err = ds.remoteTx.Commit(ctx)
		} else {
			log.Printf("Send rollback query\n")
			err = ds.remoteTx.Rollback(ctx)
		}
		// Clear current transaction context when complete.
		ds.remoteTx = nil
		return true, err
	}
	return false, err
}

// Handle transaction commands separately from the other SQL commands and create transaction context.
func (ds *DataStore) handleLocalTransaction(ctx context.Context, stmt Statement) (QueryResponse, error) {
	var err error
	var resp QueryResponse

	// Handle start of transaction and create context.
	if stmt.CmdType == command.BEGIN {
		if ds.localTx != nil {
			return resp, ErrTxExists
		}
		log.Printf("Recv begin query\n")
		ds.localTx, err = ds.dbase.BeginTx(ctx, nil)
		resp = QueryResponse{
			Error:      err,
			CommandTag: fmt.Sprintf("%s", string(stmt.CmdType)),
		}
		return resp, err
	}
	// Handle transaction end.
	if stmt.CmdType == command.COMMIT || stmt.CmdType == command.ROLLBACK {
		if ds.localTx == nil {
			return resp, ErrNoActiveTx
		}
		if stmt.CmdType == command.COMMIT {
			log.Printf("Recv commit query\n")
			err = ds.localTx.Commit()
		} else {
			log.Printf("Recv rollback query\n")
			err = ds.localTx.Rollback()
		}
		resp = QueryResponse{
			Error:      err,
			CommandTag: fmt.Sprintf("%s", string(stmt.CmdType)),
		}
		// Clear current local transaction context when complete.
		ds.localTx = nil
		return resp, err
	}
	return resp, nil
}

// Filter out only write (INSERT, UPDATE, DELETE...) statements.
func (ds *DataStore) getWriteStaments(statements []Statement, failed []uint) []Statement {
	var writeStmts []Statement
	for idx := range statements {
		if !slices.Contains(failed, uint(idx)) {
			if statements[idx].CmdType != command.BEGIN &&
				statements[idx].CmdType != command.COMMIT &&
				statements[idx].CmdType != command.ROLLBACK &&
				statements[idx].CmdType != command.UNKNOWN {
				readOnly, _ := ds.dbase.StmtReadOnly(statements[idx].Query)
				if readOnly || statements[idx].CmdType == command.SELECT {
					continue
				}
			}
			writeStmts = append(writeStmts, statements[idx])
		}
	}
	return writeStmts
}

// Filter out failed statements from the local execution response.
func mapFailedStatements(response []QueryResponse) []uint {
	var failed []uint
	for idx := range response {
		if response[idx].Error != nil {
			failed = append(failed, uint(idx))
		}
	}
	return failed
}

// Convert a Statement type to a SQL text with embed query arguments as text.
func StmtToSql(stmt Statement) string {
	sql := stmt.Query

	if len(stmt.Parameters) == 0 {
		return sql
	}
	// Prepare replace regex and replace arguments.
	re := regexp.MustCompile(`\$\d+`)
	sql = re.ReplaceAllString(sql, "%v")

	values := []any{}
	for idx := range stmt.Parameters {
		if reflect.TypeOf(stmt.Parameters[idx]).Kind() == reflect.String {
			values = append(values, fmt.Sprintf("'%s'", stmt.Parameters[idx]))
		} else {
			values = append(values, stmt.Parameters[idx])
		}
	}
	sql = fmt.Sprintf(sql, values...)
	return sql
}

// Request execution of on or more SQL statements.
func (ds *DataStore) ExecuteRemoteRequest(ctx context.Context, statements []Statement) error {
	dbname := ds.dbase.GetName()

	for _, stmt := range statements {
		handled, err := ds.handleRemoteTransaction(ctx, stmt)
		if err != nil {
			return err
		}
		// transaction command handled, process next statement.
		if handled {
			log.Printf("Handled Tx: %s\n", stmt.Query)
			continue
		}

		if ds.remoteTx != nil {
			_, err = ds.remoteTx.Exec(ctx, stmt.Query, stmt.Parameters...)
			if err != nil {
				ds.remoteTx.Rollback(ctx)
				ds.remoteTx = nil
				return err
			}
		} else {
			log.Printf("Remote query: %s\n", stmt.Query)
			if err := connpool.ExecContext(ctx, dbname, stmt.Query, stmt.Parameters...); err != nil {
				return err
			}
		}
	}
	return nil
}

// Request execution of on or more SQL statements that can contain both executes(transactions) and queries returning rows.
func (ds *DataStore) ExecuteLocalRequest(ctx context.Context, statements []Statement) ([]QueryResponse, error) {
	var err error
	var response []QueryResponse
	executor := execQuery(ds.dbase)

	// abortOnError indicates whether the caller should continue
	// processing or break.
	abortOnError := func(err error) bool {
		if err != nil && ds.localTx != nil {
			ds.localTx.Rollback()
			ds.localTx = nil
			return true
		}
		return false
	}
	// Process statements
	for _, stmt := range statements {
		resp, err := ds.handleLocalTransaction(ctx, stmt)
		if err != nil {
			return response, err
		}
		// Transaction command handled, proceed wit other statements if present.
		if resp.CommandTag != "" {
			response = append(response, resp)
			continue
		}
		// set executor if in transaction context.
		if ds.localTx != nil {
			executor = ds.localTx
		}

		returnsRows := StmtReturnsRows(stmt)
		if returnsRows {
			rows, err := executor.QueryContext(ctx, stmt.Query, stmt.Parameters...)
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
			log.Printf("Got query %s\n", stmt.Query)
			result, err := executor.ExecContext(ctx, stmt.Query, stmt.Parameters...)
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

// Request execution of on or more SQL statements that can contain both executes(transactions) and queries returning rows.
// Will replicate write statements to connected replicas.
func (ds *DataStore) Request(ctx context.Context, statements []Statement) ([]QueryResponse, error) {
	var err error
	var response []QueryResponse
	isConnected := connpool.IsConnected(ds.dbase.GetName())

	if cluster.IsPrimary() {
		// Execute locally first.
		response, err = ds.ExecuteLocalRequest(ctx, statements)
		if err != nil {
			return response, err
		}
		// Replicate data to remote.
		if ds.replicated && isConnected {
			// Exclude failed local statements to keep consistent.
			failed := mapFailedStatements(response)
			writeStmts := ds.getWriteStaments(statements, failed)
			// Replicate writes to remote replica node (secondary)
			if len(writeStmts) != 0 {
				err := ds.ExecuteRemoteRequest(ctx, writeStmts)
				// TODO handle remote errors, drop remote replica or demote primary
				if err != nil {
					// TODO Handle network error
					log.Printf("Got error from remote execution on master: %s", err.Error())
				}
			}
		}
	} else {
		// Secondary
		// execute remote if data-sync is on.
		if ds.replicated && isConnected {
			writeStmts := ds.getWriteStaments(statements, nil)
			if len(writeStmts) != 0 {
				err = ds.ExecuteRemoteRequest(ctx, writeStmts)
				// TODO handle remote error.
				if err != nil {
					log.Printf("Got error from remote execution on replica: %s", err.Error())
				}
			}
		}
		// Execute locally after replicated write statements.
		response, err = ds.ExecuteLocalRequest(ctx, statements)
	}
	return response, err
}
