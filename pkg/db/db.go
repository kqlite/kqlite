package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kqlite/kqlite/pkg/sysdb"
	"github.com/mattn/go-sqlite3"
)

// DB is the SQL database backend.
type Database struct {
	path      string  // Path to database file.
	fkEnabled bool    // Foreign key constraints enabled.
	wal       bool    // WAL enabled.
	rwdb      *sql.DB // Database connection for database reads and writes.
	rodb      *sql.DB // Database connection for database reads only.
}

// CheckpointMode is the mode in which a checkpoint runs.
type CheckpointMode int

const (
	// CheckpointRestart instructs the checkpoint to run in restart mode.
	CheckpointRestart CheckpointMode = iota
	// CheckpointTruncate instructs the checkpoint to run in truncate mode.
	CheckpointTruncate
)

var (
	checkpointPRAGMAs = map[CheckpointMode]string{
		CheckpointRestart:  "PRAGMA wal_checkpoint(RESTART)",
		CheckpointTruncate: "PRAGMA wal_checkpoint(TRUNCATE)",
	}
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Open opens a file-based database using the default driver and the specified options.
func Open(dbPath string, fkEnabled, wal bool) (*Database, error) {
	var err error
	var rwdb *sql.DB

	rwdb, err = openDBforWrite(dbPath, fkEnabled, wal)
	if err != nil {
		return nil, err
	}

	readOnly := true
	rodb, err := openSQLiteDB(dbPath, readOnly, fkEnabled, wal)
	if err != nil {
		return nil, err
	}

	return &Database{
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

	// Make sure kqlite has full control over the checkpointing process.
	if _, err := rwdb.Exec("PRAGMA wal_autocheckpoint=0"); err != nil {
		return nil, fmt.Errorf("disable autocheckpointing: %s", err.Error())
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
	rwdb.SetMaxOpenConns(1)

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

	opts.Add("_sync", "0")
	opts.Add("cache", "shared")
	opts.Add("_busy_timeout", "3000")

	return fmt.Sprintf("file:%s?%s", path, opts.Encode())
}

// SetBusyTimeout sets the busy timeout for the database. If a timeout is
// is less than zero it is not set.
func (dbase *Database) SetBusyTimeout(rwMs, roMs int) (err error) {
	if rwMs >= 0 {
		_, err := dbase.rwdb.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d", rwMs))
		if err != nil {
			return err
		}
	}
	if roMs >= 0 {
		_, err = dbase.rodb.Exec(fmt.Sprintf("PRAGMA busy_timeout=%d", roMs))
		if err != nil {
			return err
		}
	}

	return nil
}

// BusyTimeout returns the current busy timeout value.
func (dbase *Database) BusyTimeout() (rwMs, roMs int, err error) {
	err = dbase.rwdb.QueryRow("PRAGMA busy_timeout").Scan(&rwMs)
	if err != nil {
		return 0, 0, err
	}
	err = dbase.rodb.QueryRow("PRAGMA busy_timeout").Scan(&roMs)
	if err != nil {
		return 0, 0, err
	}
	return rwMs, roMs, nil
}

// Checkpoint checkpoints the WAL file. If the WAL file is not enabled, this
// function is a no-op.
func (dbase *Database) Checkpoint(mode CheckpointMode) error {
	return dbase.CheckpointWithTimeout(mode, 0)
}

// CheckpointWithTimeout performs a WAL checkpoint. If the checkpoint does not
// run to completion within the given duration, an error is returned. If the
// duration is 0, the busy timeout is not modified before executing the
// checkpoint.
func (dbase *Database) CheckpointWithTimeout(mode CheckpointMode, dur time.Duration) (err error) {
	if dur > 0 {
		rwBt, _, err := dbase.BusyTimeout()
		if err != nil {
			return fmt.Errorf("failed to get busy_timeout on checkpointing connection: %s", err.Error())
		}
		if err := dbase.SetBusyTimeout(int(dur.Milliseconds()), -1); err != nil {
			return fmt.Errorf("failed to set busy_timeout on checkpointing connection: %s", err.Error())
		}
		defer func() {
			// Reset back to default
			if err := dbase.SetBusyTimeout(rwBt, -1); err != nil {
				// TODO Fix logging.
				// db.logger.Printf("failed to reset busy_timeout on checkpointing connection: %s", err.Error())
			}
		}()
	}

	ok, nPages, nMoved, err := checkpointDB(dbase.rwdb, mode)
	if err != nil {
		return fmt.Errorf("error checkpointing WAL: %s", err.Error())
	}
	if ok != 0 {
		return fmt.Errorf("failed to completely checkpoint WAL (%d ok, %d pages, %d moved)", ok, nPages, nMoved)
	}
	return nil
}

func checkpointDB(rwdb *sql.DB, mode CheckpointMode) (ok, pages, moved int, err error) {
	err = rwdb.QueryRow(checkpointPRAGMAs[mode]).Scan(&ok, &pages, &moved)
	return
}

// Vacuum runs a VACUUM on the database.
func (dbase *Database) Vacuum() error {
	_, err := dbase.rwdb.Exec("VACUUM")
	return err
}

// VacuumInto VACUUMs the database into the file at path
func (dbase *Database) VacuumInto(path string) error {
	_, err := dbase.rwdb.Exec(fmt.Sprintf("VACUUM INTO '%s'", path))
	return err
}

// Closes the underlying database connection.
func (dbase *Database) Close() error {
	return dbase.rodb.Close()
}

// A tiny wrapper around sql.Exec.
// Executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (dbase *Database) Exec(query string, args ...any) (sql.Result, error) {
	if query != "" {
		return dbase.rwdb.Exec(query, args...)
	}
	return nil, nil
}

// A tiny wrapper around sql.ExecContext.
// Executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (dbase *Database) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if query != "" {
		return dbase.rwdb.ExecContext(ctx, query, args...)
	}
	return nil, nil
}

// A tiny wrapper around sql.Query.
// Executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (dbase *Database) Query(query string, args ...any) (*sql.Rows, error) {
	if query != "" {
		ro, _ := dbase.StmtReadOnly(query)
		if ro {
			return dbase.rodb.Query(query, args...)
		} else {
			return dbase.rwdb.Query(query, args...)
		}
	}
	return nil, nil
}

// A tiny wrapper around sql.QueryContext.
// Executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (dbase *Database) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if query != "" {
		ro, _ := dbase.StmtReadOnly(query)
		if ro {
			return dbase.rodb.QueryContext(ctx, query, args...)
		} else {
			return dbase.rwdb.QueryContext(ctx, query, args...)
		}
	}
	return nil, nil
}

// StmtReadOnly returns whether the given SQL statement is read-only.
// As per https://www.sqlite.org/c3ref/stmt_readonly.html, this function
// may not return 100% correct results, but should cover most scenarios.
func (dbase *Database) StmtReadOnly(sql string) (bool, error) {
	conn, err := dbase.rodb.Conn(context.Background())
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return dbase.StmtReadOnlyWithConn(sql, conn)
}

// StmtReadOnlyWithConn returns whether the given SQL statement is read-only, using
// the given connection.
func (dbase *Database) StmtReadOnlyWithConn(sql string, conn *sql.Conn) (bool, error) {
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

// A tiny wrapper around sql.BeginTx.
func (dbase *Database) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return dbase.rwdb.BeginTx(ctx, opts)
}

func (dbase *Database) GetName() string {
	_, file := filepath.Split(dbase.path)
	return strings.TrimSuffix(file, ".db")
}
