package db

import (
	"database/sql"
	"sync"
)

// Central location to store sqlite databases.
type Pool struct {
	sync.Map
}

var dbpool Pool

// Open RW connection to database path.
func openDBforWrite(dbPath string, fkEnabled, wal bool) (*sql.DB, error) {
	if sqldb, found := dbpool.Load(dbPath); found && sqldb != nil {
		dbase := sqldb.(*sql.DB)
		return dbase, nil
	} else {
		if sqldb, err := openSQLiteDB(dbPath, false, fkEnabled, wal); err != nil {
			return nil, err
		} else {
			// Store database connection in cache.
			dbpool.Store(dbPath, sqldb)

			return sqldb, err
		}
	}
}

// Clear and flush the global DB connection pool
func ClearPool() {
	// Close all database connections stored in the pool.
	dbpool.Range(func(key, value any) bool {
		dbase := value.(*sql.DB)
		dbase.Close()
		return true
	})
	// Clear all sqlite database isntances.
	dbpool.Clear()
}
