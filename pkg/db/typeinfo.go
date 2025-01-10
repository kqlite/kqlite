package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgtype"
)

// SQLIte to PostgreSQL type mapping.
func Typemap() map[string]uint32 {
	return map[string]uint32{
		// Integer
		"INT":              pgtype.Int4OID,
		"INTEGER":          pgtype.Int8OID,
		"TINYINT":          pgtype.Int8OID,
		"SMALLINT":         pgtype.Int8OID,
		"MEDIUMINT":        pgtype.Int8OID,
		"BIGINT":           pgtype.Int8OID,
		"UNSIGNED BIG INT": pgtype.Int8OID,
		"INT2":             pgtype.Int2OID,
		"INT8":             pgtype.Int8OID,
		// String
		"CHARACTER(20)":          pgtype.TextOID,
		"VARCHAR(255)":           pgtype.VarcharOID,
		"VARYING CHARACTER(255)": pgtype.VarcharOID,
		"NCHAR(55)":              pgtype.TextOID,
		"NATIVE CHARACTER(70)":   pgtype.TextOID,
		"NVARCHAR(100)":          pgtype.TextOID,
		"TEXT":                   pgtype.TextOID,
		"CLOB":                   pgtype.TextOID,
		// Binary
		"BLOB": pgtype.ByteaOID,
		// Floating point
		"REAL":             pgtype.Float8OID,
		"DOUBLE":           pgtype.Float8OID,
		"DOUBLE PRECISION": pgtype.Float8OID,
		"FLOAT":            pgtype.Float8OID,
		// Numeric
		"NUMERIC":       pgtype.NumericOID,
		"DECIMAL(10,5)": pgtype.NumericOID,
		"BOOLEAN":       pgtype.BoolOID,
		// Date/timestamp
		"DATE":      pgtype.TextOID, //pgtype.DateOID,
		"TIMESTAMP": pgtype.TextOID, //pgtype.TimestampOID,
		"DATETIME":  pgtype.TextOID,
	}
}

func joinElemNames(elems []string) string {
	var result string

	elemsLen := len(elems)
	if elemsLen == 0 {
		return result
	}
	for idx := range elems {
		if idx < (elemsLen - 1) {
			result += fmt.Sprintf("'%s', ", elems[idx])
		} else {
			result += fmt.Sprintf("'%s'", elems[idx])
		}
	}
	return result
}

// Lookup columns type from SQLite by checking the provided list of tables if provided,
// otherwise check all tables.
// Will return the corresponding PostgreSQL type compatible with the wire protocol.
func LookupTypeInfo(ctx context.Context, db *DB, columns, tables []string) ([]uint32, error) {
	var columnTypes []uint32
	if len(columns) == 0 || db == nil {
		return columnTypes, nil
	}

	sqlText := `WITH tables AS (SELECT name tableName, sql 
			    FROM sqlite_master WHERE type = 'table' `
	// Apply a table filter if a specific set of tables is provided.
	if len(tables) != 0 {
		tableSet := joinElemNames(tables)
		sqlText += fmt.Sprintf("AND tableName IN (%s)) ", tableSet)
	} else {
		sqlText += `AND tableName NOT LIKE 'sqlite_%') `
	}

	fieldSet := joinElemNames(columns)
	sqlText += `SELECT fields.name, fields.type
				FROM tables CROSS JOIN pragma_table_info(tables.tableName) fields WHERE `
	sqlText += fmt.Sprintf("fields.name IN (%s) GROUP BY fields.name;", fieldSet)

	rows, err := db.QueryContext(ctx, sqlText)
	if err != nil {
		return columnTypes, err
	}

	for rows.Next() {
		var colName, colType string
		if err := rows.Scan(&colName, &colType); err != nil {
			return columnTypes, nil
		}
		if pgColtype, exists := Typemap()[colType]; exists {
			columnTypes = append(columnTypes, pgColtype)
		} else {
			// Set TextOID as default if can't lookup type
			// TODO log warning
			columnTypes = append(columnTypes, pgtype.TextOID)
		}
	}
	rerr := rows.Close()
	if rerr != nil {
		log.Fatal(rerr)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	return columnTypes, nil
}
