package catalog

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	DriverName = "kqlite-sqlite3"

	pg_database_sql = `
		CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_database USING pg_database_module 
		(oid, datname, datdba, encoding, datcollate, datctype, datistemplate, datallowconn, datconnlimit, datlastsysoid, datfrozenxid, datminmxid, dattablespace, datacl);`

	pg_namespace_sql = `
		CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_namespace USING pg_namespace_module (oid, nspname, nspowner, nspacl);`

	pg_description_sql = `
		CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_description USING pg_description_module (objoid, classoid, objsubid, description);`

	pg_settings_sql = `
		CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_settings USING pg_settings_module 
		(name, setting, unit, category, short_desc, extra_desc, context, vartype, source, min_val, max_val, enumvals, boot_val, reset_val, sourcefile, sourceline, pending_restart);`

	pg_type_sql = `
		CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_type USING pg_type_module 
		(oid, typname, typnamespace, typowner, typlen, typbyval, typtype, typcategory, typispreferred, typisdefined, typdelim, typrelid, typelem, typarray, typinput, typoutput, typreceive, typsend, typmodin, typmodout, typanalyze, typalign, typstorage, typnotnull, typbasetype, typtypmod, typndims, typcollation, typdefaultbin, typdefault, typacl);`

	pg_class_sql = `
		CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_class USING pg_class_module 
		(oid, relname, relnamespace, reltype, reloftype, relowner, relam, relfilenode, reltablespace, relpages, reltuples, relallvisible, reltoastrelid, relhasindex, relisshared, relpersistence, relkind, relnatts, relchecks, relhasrules, relhastriggers, relhassubclass, relrowsecurity, relforcerowsecurity, relispopulated, relreplident, relispartition, relrewrite, relfrozenxid, relminmxid, relacl, reloptions, relpartbound);`

	pg_range_sql = `
		CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_range USING pg_range_module (rngtypid, rngsubtype, rngmultitypid, rngcollation, rngsubopc, rngcanonical, rngsubdiff);`
)

// Initialize virtual table catalog.
func initCatatog(conn *sqlite3.SQLiteConn) error {
	// Attach an in-memory database for pg_catalog.
	if _, err := conn.Exec(`ATTACH ':memory:' AS pg_catalog`, nil); err != nil {
		// Already attached, do nothing.
		if err.Error() == "database pg_catalog is already in use" {
			return nil
		}
		return fmt.Errorf("attach pg_catalog: %w", err)
	}

	// Register virtual tables to imitate postgres.
	if _, err := conn.Exec(pg_database_sql, nil); err != nil {
		return fmt.Errorf("create pg_database: %w", err)
	}
	if _, err := conn.Exec(pg_namespace_sql, nil); err != nil {
		return fmt.Errorf("create pg_namespace: %w", err)
	}
	if _, err := conn.Exec(pg_description_sql, nil); err != nil {
		return fmt.Errorf("create pg_description: %w", err)
	}
	if _, err := conn.Exec(pg_settings_sql, nil); err != nil {
		return fmt.Errorf("create pg_settings: %w", err)
	}
	if _, err := conn.Exec(pg_type_sql, nil); err != nil {
		return fmt.Errorf("create pg_type: %w", err)
	}
	if _, err := conn.Exec(pg_class_sql, nil); err != nil {
		return fmt.Errorf("create pg_class: %w", err)
	}
	if _, err := conn.Exec(pg_range_sql, nil); err != nil {
		return fmt.Errorf("create pg_range: %w", err)
	}
	return nil
}

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

			if err := initCatatog(conn); err != nil {
				return err
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
