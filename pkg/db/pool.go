package db

import (
	"database/sql"
	"sync"
)

type Pool struct {
	// Database storage access.
	store sync.Map
}

var dbpool Pool

// Open RW connection to database path.
func openDBforWrite(dbPath string, fkEnabled, wal bool) (*sql.DB, error) {
	if sqldb, found := dbpool.store.Load(dbPath); found && sqldb != nil {
		dbase := sqldb.(*sql.DB)
		return dbase, nil
	} else {
		if sqldb, err := openSQLiteDB(dbPath, false, fkEnabled, wal); err != nil {
			return nil, err
		} else {
			// Store database connection in cache.
			dbpool.store.Store(dbPath, sqldb)

			return sqldb, err
		}
	}
}

// Clear and flush the global DB connection pool
func ClearPool() {
	// Close all database connections stored in the pool.
	dbpool.store.Range(func(key, value any) bool {
		dbase := value.(*sql.DB)
		dbase.Close()
		return true
	})

	// Clear all connections references.
	dbpool.store.Clear()
}
