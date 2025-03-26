package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/kqlite/kqlite/pkg/sysdb"
	"github.com/kqlite/kqlite/pkg/util/pgerror"

	"github.com/jackc/pgerrcode"

	"github.com/mattn/go-sqlite3"
)

// DB is the SQL database backend.
type DB struct {
	path      string  // Path to database file.
	fkEnabled bool    // Foreign key constraints enabled.
	wal       bool    // WAL enabled.
	rwdb      *sql.DB // Database connection for database reads and writes.
	rodb      *sql.DB // Database connection for database reads only.
}

type execerQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Execute queries in a connection context, holds the current transaction specific to the client connection.
type ExecuteQueryContext struct {
	db  *DB
	tx  *sql.Tx
	ctx context.Context
}

type ExecuteQueryResponse struct {
	Rows       *sql.Rows
	CommandTag string
	Error      error
}

type SqlCmdType string

const (
	CMD_SELECT   SqlCmdType = "SELECT"
	CMD_UPDATE   SqlCmdType = "UPDATE"
	CMD_INSERT   SqlCmdType = "INSERT"
	CMD_DELETE   SqlCmdType = "DELETE"
	CMD_BEGIN    SqlCmdType = "BEGIN"
	CMD_COMMIT   SqlCmdType = "COMMIT"
	CMD_ROLLBACK SqlCmdType = "ROLLBACK"
	CMD_UNKNOWN  SqlCmdType = "UNKNOWN"
)

// Represents a single SQL statement.
type Statement struct {
	// SQL Query text
	Query string

	// SQL Command type (ex. SELECT, INSERT, UPDATE ...)
	CmdType SqlCmdType

	// Statement parameter values if any.
	Parameters []any

	// Indicates whether statement returns rows even in case of INSERT, UPDATE or others ..
	ReturnsRows bool
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Open opens a file-based database using the default driver.
func Open(dbPath string, fkEnabled, wal bool) (*DB, error) {
	rwdb, err := openDBforWrite(dbPath, fkEnabled, wal)
	if err != nil {
		return nil, err
	}

	readOnly := true
	rodb, err := openSQLiteDB(dbPath, readOnly, fkEnabled, wal)
	if err != nil {
		return nil, err
	}

	return &DB{
		path:      dbPath,
		fkEnabled: fkEnabled,
		wal:       wal,
		rwdb:      rwdb,
		rodb:      rodb,
	}, nil
}

// Open database implementation for sqlite.
func openSQLiteDB(dbPath string, readOnly, fkEnabled, wal bool) (*sql.DB, error) {
	if !fileExists(dbPath) {
		// Check if database file exists otherwise create it.
		if f, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0644); err != nil {
			return nil, err
		} else if err = f.Close(); err != nil {
			return nil, err
		}
	}

	// Read-only connection
	if readOnly {
		rodsn := makeDSN(dbPath, readOnly, fkEnabled, wal)
		rodb, err := sql.Open(sysdb.DriverName, rodsn)
		if err != nil {
			return nil, err
		}

		rodb.SetConnMaxIdleTime(30 * time.Second)
		rodb.SetConnMaxLifetime(0)

		return rodb, nil
	}

	// RW connection
	rwdsn := makeDSN(dbPath, readOnly, fkEnabled, wal)
	rwdb, err := sql.Open(sysdb.DriverName, rwdsn)
	if err != nil {
		return nil, err
	}

	if err := rwdb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping on-disk database: %s", err.Error())
	}

	if wal && !fileExists(dbPath+"-wal") {
		// Force creation of the WAL files so any external read-only connections
		// can read the database. See https://www.sqlite.org/draft/wal.html, section 5.
		if _, err := rwdb.Exec("BEGIN IMMEDIATE"); err != nil {
			return nil, err
		}

		if _, err := rwdb.Exec("ROLLBACK"); err != nil {
			return nil, err
		}
	}

	// Set connection pool behaviour.
	rwdb.SetConnMaxLifetime(0)
	//rwdb.SetMaxOpenConns(1)

	return rwdb, nil
}

// makeDSN returns a SQLite DSN(Data source name) for the given path, with the given options.
func makeDSN(path string, readOnly, fkEnabled, walEnabled bool) string {
	opts := url.Values{}

	opts.Add("_fk", strconv.FormatBool(fkEnabled))
	opts.Add("_journal", "WAL")
	if !walEnabled {
		opts.Set("_journal", "DELETE")
	}

	if readOnly {
		opts.Add("mode", "ro")
	}

	opts.Add("_sync", "1")

	opts.Add("cache", "shared")

	opts.Add("_busy_timeout", "3000")

	return fmt.Sprintf("file:%s?%s", path, opts.Encode())
}

// Vacuum runs a VACUUM on the database.
func (db *DB) Vacuum() error {
	_, err := db.rwdb.Exec("VACUUM")
	return err
}

// VacuumInto VACUUMs the database into the file at path
func (db *DB) VacuumInto(path string) error {
	_, err := db.rwdb.Exec(fmt.Sprintf("VACUUM INTO '%s'", path))
	return err
}

// Closes the underlying database connection.
func (db *DB) Close() error {
	return db.rodb.Close()
}

// A tiny wrapper around DB.ExecContext.
// Executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if query != "" {
		return db.rwdb.ExecContext(context.Background(), query, args...)
	}
	return nil, nil
}

// A tiny wrapper around DB.QueryContext.
// Executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if query != "" {
		ro, _ := db.StmtReadOnly(query)
		if ro {
			return db.rodb.QueryContext(context.Background(), query, args...)
		} else {
			return db.rwdb.QueryContext(context.Background(), query, args...)
		}
	}
	return nil, nil
}

// StmtReadOnly returns whether the given SQL statement is read-only.
// As per https://www.sqlite.org/c3ref/stmt_readonly.html, this function
// may not return 100% correct results, but should cover most scenarios.
func (db *DB) StmtReadOnly(sql string) (bool, error) {
	conn, err := db.rodb.Conn(context.Background())
	if err != nil {
		return false, err
	}
	defer conn.Close()

	return db.StmtReadOnlyWithConn(sql, conn)
}

// StmtReadOnlyWithConn returns whether the given SQL statement is read-only, using
// the given connection.
func (db *DB) StmtReadOnlyWithConn(sql string, conn *sql.Conn) (bool, error) {
	var readOnly bool
	f := func(driverConn interface{}) error {
		c := driverConn.(*sqlite3.SQLiteConn)
		drvStmt, err := c.Prepare(sql)
		if err != nil {
			return err
		}
		defer drvStmt.Close()
		sqliteStmt := drvStmt.(*sqlite3.SQLiteStmt)
		readOnly = sqliteStmt.Readonly()
		return nil
	}

	if err := conn.Raw(f); err != nil {
		return false, err
	}

	return readOnly, nil
}

// Checks if statemt must return rows.
func (db *DB) StmtReturnsRows(stmt Statement) bool {
	if stmt.ReturnsRows || stmt.CmdType == CMD_SELECT {
		return true
	}

	return false
}

// Handle transaction commands separately from the other common SQL commands.
func (qcontext *ExecuteQueryContext) handleTransaction(stmt Statement) (bool, error) {
	var err error
	var handled bool

	switch stmt.CmdType {
	case CMD_BEGIN: // Check for transaction start
		handled = true
		if qcontext.tx != nil {
			return handled, fmt.Errorf("Syntax error, an active transaction is already present")
		}
		if qcontext.tx, err = qcontext.db.rwdb.BeginTx(qcontext.ctx, nil); err != nil {
			return handled, err
		}
		break

	case CMD_COMMIT: // Check for transaction end
		handled = true
		if qcontext.tx == nil {
			return handled, fmt.Errorf("No active transaction to COMMIT")
		}

		if err := qcontext.tx.Commit(); err != nil {
			return handled, err
		}
		qcontext.tx = nil
		break

	case CMD_ROLLBACK: // Check for transaction abort/rollback
		handled = true
		if qcontext.tx == nil {
			return handled, fmt.Errorf("No active transaction to ROLLBACK")
		}
		if err := qcontext.tx.Rollback(); err != nil {
			return handled, err
		}
		qcontext.tx = nil
		break
	}

	return handled, err
}

// Get a ExecuteQueryContext for this database.
func (db *DB) CreateContext(ctx context.Context) *ExecuteQueryContext {
	return &ExecuteQueryContext{
		db:  db,
		ctx: ctx,
		tx:  nil,
	}
}

// Request execution of on or more SQL statements that can contain both executes(transactions) and queries returning rows.
func (qcontext *ExecuteQueryContext) Request(statements []Statement) ([]ExecuteQueryResponse, error) {
	var err error
	var response []ExecuteQueryResponse

	// abortOnError indicates whether the caller should continue
	// processing or break.
	abortOnError := func(err error) bool {
		if err != nil && qcontext.tx != nil {
			qcontext.tx.Rollback()
			qcontext.tx = nil
			return true
		}
		return false
	}

	// point executor to default DB connection
	eq := execerQueryer(qcontext.db)

	if qcontext.tx != nil {
		eq = qcontext.tx
	}

	for _, stmt := range statements {
		if handled, err := qcontext.handleTransaction(stmt); handled || err != nil {
			response = append(response, ExecuteQueryResponse{
				Error:      err,
				CommandTag: fmt.Sprintf("%s", string(stmt.CmdType)),
			})
			return response, err
		}

		// Check if statement must return rows
		returnsRows := qcontext.db.StmtReturnsRows(stmt)

		if returnsRows {
			rows, err := eq.QueryContext(qcontext.ctx, stmt.Query, stmt.Parameters...)
			if err != nil {
				fmt.Printf("Error from query %s", err.Error())
			}

			response = append(response, ExecuteQueryResponse{
				Error:      err,
				Rows:       rows,
				CommandTag: fmt.Sprintf("%s 1", stmt.Query),
			})

			if abortOnError(err) {
				break
			}
		} else {
			result, err := eq.ExecContext(qcontext.ctx, stmt.Query, stmt.Parameters...)

			if err != nil {
				fmt.Printf("Error from query %s", err.Error())
				// TODO SQLITE_CONSTRAINT_UNIQUE
				if stmt.CmdType == CMD_INSERT {
					err = pgerror.New(pgerrcode.UniqueViolation, err.Error())
				}
			}

			var rowsAffected int64
			if result != nil {
				rowsAffected, _ = result.RowsAffected()
			}

			response = append(response, ExecuteQueryResponse{
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
