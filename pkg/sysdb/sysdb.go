package sysdb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sqlite3 "github.com/mattn/go-sqlite3"
)

const DriverName = "kqlite-sqlite3"

// Kqlite system database/catalog name and schema definitions.
const (
	// Catalog name, for remote DSN.
	Catalog = "kqlite"

	// Catalog filename, for local DSN.
	CatalogFile = "kqlite.db"

	// Catalog schema/table definitions.
	// addr is in the form host:port or :port
	CatalogSchema = `
CREATE TABLE IF NOT EXISTS replicas
	(
		id		INTEGER PRIMARY KEY AUTOINCREMENT,
		addr	TEXT,
		db		TEXT
	)
	`
)

func init() {
	sql.Register(DriverName, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("current_catalog", current_catalog, true); err != nil {
				return fmt.Errorf("cannot register current_catalog() function")
			}

			if err := conn.RegisterFunc("current_schema", currentSchema, true); err != nil {
				return fmt.Errorf("cannot register current_schema() function")
			}

			if err := conn.RegisterFunc("current_user", currentUser, true); err != nil {
				return fmt.Errorf("cannot register current_user() function")
			}

			if err := conn.RegisterFunc("session_user", sessionUser, true); err != nil {
				return fmt.Errorf("cannot register session_user() function")
			}

			if err := conn.RegisterFunc("user", user, true); err != nil {
				return fmt.Errorf("cannot register user() function")
			}

			if err := conn.RegisterFunc("show", show, true); err != nil {
				return fmt.Errorf("cannot register show() function")
			}

			if err := conn.RegisterFunc("format_type", formatType, true); err != nil {
				return fmt.Errorf("cannot register format_type() function")
			}

			if err := conn.RegisterFunc("version", version, true); err != nil {
				return fmt.Errorf("cannot register version() function")
			}

			if err := conn.RegisterFunc("pg_total_relation_size", pg_total_relation_size, true); err != nil {
				return fmt.Errorf("cannot register pg_total_relation_size() function")
			}

			if err := conn.CreateModule("pg_database_module", &PGDatabaseModule{}); err != nil {
				return fmt.Errorf("cannot register pg_database module")
			}

			if err := conn.CreateModule("pg_type_module", &pgTypeModule{}); err != nil {
				return fmt.Errorf("cannot register pg_type module")
			}

			if err := conn.CreateModule("pg_settings_module", &pgSettingsModule{}); err != nil {
				return fmt.Errorf("cannot register pg_settings module")
			}

			if err := conn.CreateModule("pg_range_module", &pgRangeModule{}); err != nil {
				return fmt.Errorf("cannot register pg_range module")
			}

			if err := conn.CreateModule("pg_namespace_module", &pgNamespaceModule{}); err != nil {
				return fmt.Errorf("cannot register pg_namespace module")
			}

			if err := conn.CreateModule("pg_description_module", &pgDescriptionModule{}); err != nil {
				return fmt.Errorf("cannot register pg_description module")
			}

			if err := conn.CreateModule("pg_class_module", &pgClassModule{}); err != nil {
				return fmt.Errorf("cannot register pg_class module")
			}

			return nil
		},
	})
}

func current_catalog() string {
	return "public" /*"kqlite v0.0.0"*/
}

func currentSchema() string { return "public" }

func currentUser() string { return "sqlite3" }
func sessionUser() string { return "sqlite3" }
func user() string        { return "sqlite3" }

func version() string { return "cockroachdb" /*"kqlite v0.0.0"*/ }

func formatType(type_oid, typemod string) string { return "" }

func show(name string) string { return "" }

// Returns Total disk space used by the specified table, including all indexes and TOAST data.
func pg_total_relation_size(name string) int64 {
	if finfo, err := os.Stat(filepath.Join(os.Getenv("DATA_DIR"), name+".db")); err != nil {
		return -1
	} else {
		return finfo.Size()
	}
}

func DatabaseTypeConvSqlite(t string) int {
	if strings.Contains(t, "INT") {
		return sqlite3.SQLITE_INTEGER
	}
	if t == "CLOB" || t == "TEXT" ||
		strings.Contains(t, "CHAR") {
		return sqlite3.SQLITE_TEXT
	}
	if t == "BLOB" {
		return sqlite3.SQLITE_BLOB
	}
	if t == "REAL" || t == "FLOAT" ||
		strings.Contains(t, "DOUBLE") {
		return sqlite3.SQLITE_REAL
	}
	if t == "DATE" || t == "DATETIME" ||
		t == "TIMESTAMP" {
		return sqlite3.SQLITE_TIME
	}
	if t == "NUMERIC" ||
		strings.Contains(t, "DECIMAL") {
		return sqlite3.SQLITE_NUMERIC
	}
	if t == "BOOLEAN" {
		return sqlite3.SQLITE_BOOL
	}

	return sqlite3.SQLITE_NULL
}
