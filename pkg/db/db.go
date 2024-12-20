package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/kqlite/kqlite/pkg/sqlite"
	"github.com/mattn/go-sqlite3"
)

const (
	// ModeReadOnly is the mode to open a database in read-only mode.
	ModeReadOnly = true
	// ModeReadWrite is the mode to open a database in read-write mode.
	ModeReadWrite = false
)

// DB is the SQL database backend.
type DB struct {
	path string // Path to database file.

	fkEnabled bool // Foreign key constraints enabled
	wal       bool // WAL enabled

	rwdb *sql.DB // Database connection for database reads and writes.
	rodb *sql.DB // Database connection database reads.

	rwdsn string // DSN used for read-write connection
	rodsn string // DSN used for read-only connections
}

// Open opens a file-based database using the default driver.
func Open(dbPath string, fkEnabled, wal bool) (retDB *DB, retErr error) {
	return openSQLiteDB(dbPath, false, false)
}

// Open database implementation for sqlite.
func openSQLiteDB(dbPath string, fkEnabled, wal bool) (*DB, error) {
	// Main RW connection
	rwdsn := makeDSN(dbPath, ModeReadWrite, fkEnabled, wal)
	rwdb, err := sql.Open(sqlite.DriverName, rwdsn)
	if err != nil {
		return nil, err
	}

	if err := rwdb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping on-disk database: %s", err.Error())
	}

	// Read-only connection
	rodsn := makeDSN(dbPath, ModeReadOnly, fkEnabled, wal)
	rodb, err := sql.Open(sqlite.DriverName, rodsn)
	if err != nil {
		return nil, err
	}

	// Set connection pool behaviour.
	rwdb.SetConnMaxLifetime(0)
	// rwdb.SetMaxOpenConns(1) // Key to ensure a new connection doesn't enable checkpointing
	rodb.SetConnMaxIdleTime(30 * time.Second)
	rodb.SetConnMaxLifetime(0)

	return &DB{
		path:      dbPath,
		fkEnabled: fkEnabled,
		wal:       wal,
		rwdb:      rwdb,
		rodb:      rodb,
		rwdsn:     rwdsn,
		rodsn:     rodsn,
	}, nil
}

// makeDSN returns a SQLite DSN(Data source name) for the given path, with the given options.
func makeDSN(path string, readOnly, fkEnabled, walEnabled bool) string {
	opts := url.Values{}

	if readOnly {
		opts.Add("mode", "ro")
	}

	opts.Add("_fk", strconv.FormatBool(fkEnabled))
	opts.Add("_journal", "WAL")

	if !walEnabled {
		opts.Set("_journal", "DELETE")
	}
	opts.Add("_sync", "1")

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

// Backup writes a consistent snapshot of the database to the given file.
// The resultant SQLite database file will be in DELETE mode. This function
// can be called when changes to the database are in flight.
func (db *DB) Backup(path string, vacuum bool) error {
	dstDB, err := Open(path, false, false)
	if err != nil {
		return fmt.Errorf("open: %s", err.Error())
	}

	// clean up when done.
	defer dstDB.Close()

	if err := copyDatabase(db, dstDB); err != nil {
		return fmt.Errorf("backup database: %s", err)
	}

	// Source database might be in WAL mode.
	_, err = dstDB.QueryContext(context.Background(), "PRAGMA journal_mode=DELETE")
	if err != nil {
		return err
	}

	if vacuum {
		if err := dstDB.Vacuum(); err != nil {
			return err
		}
	}

	return dstDB.Close()
}

func copyDatabase(src *DB, dst *DB) error {
	dstConn, err := dst.rwdb.Conn(context.Background())
	if err != nil {
		return err
	}

	// clean up.
	defer dstConn.Close()

	srcConn, err := src.rodb.Conn(context.Background())
	if err != nil {
		return err
	}
	// clean up.
	defer srcConn.Close()

	var dstSQLiteConn *sqlite3.SQLiteConn

	bf := func(driverConn interface{}) error {
		srcSQLiteConn := driverConn.(*sqlite3.SQLiteConn)
		return copyDatabaseConnection(dstSQLiteConn, srcSQLiteConn)
	}
	return dstConn.Raw(
		func(driverConn interface{}) error {
			dstSQLiteConn = driverConn.(*sqlite3.SQLiteConn)
			return srcConn.Raw(bf)
		})
}

func copyDatabaseConnection(dst, src *sqlite3.SQLiteConn) error {
	bk, err := dst.Backup("main", src, "main")
	if err != nil {
		return err
	}

	for {
		done, err := bk.Step(-1)
		if err != nil {
			_ = bk.Finish() // Return the outer error
			return err
		}
		if done {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	return bk.Finish()
}

// Closes the underlying database connection.
func (db *DB) Close() error {
	if err := db.rwdb.Close(); err != nil {
		return err
	}
	return db.rodb.Close()
}

// A tiny wraper around DB.ExecContext.
// Executes a query without returning any rows. The args are for any placeholder parameters in the query.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if query != "" {
		return db.rwdb.ExecContext(ctx, query, args...)
	}
	return nil, nil
}

// A tiny wraper around DB.QueryContext.
// Executes a query that returns rows, typically a SELECT. The args are for any placeholder parameters in the query.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if query != "" {
		return db.rodb.QueryContext(ctx, query, args...)
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
