package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// SQLIte to PostgreSQL type mapping.
func Typemap() map[string]uint32 {
	return map[string]uint32{
		// Integer
		"INT":              pgtype.Int8OID,
		"INTEGER":          pgtype.Int8OID,
		"TINYINT":          pgtype.Int2OID,
		"SMALLINT":         pgtype.Int4OID,
		"MEDIUMINT":        pgtype.Int4OID,
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
		// Boolean
		"BOOLEAN": pgtype.BoolOID,
		// Date/timestamp
		"DATE":      pgtype.DateOID,
		"TIMESTAMP": pgtype.TimestampOID,
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

	// Get column name with corresponding type from the row result.
	columnDBInfo := map[string]string{}
	for rows.Next() {
		var colName, colType string
		if err := rows.Scan(&colName, &colType); err != nil {
			return columnTypes, nil
		}
		columnDBInfo[colName] = colType
	}

	// match column name and type with provided column arguments.
	for _, colName := range columns {
		if colType, found := columnDBInfo[colName]; found {
			if pgColtype, exists := Typemap()[colType]; exists {
				columnTypes = append(columnTypes, pgColtype)
			} else {
				// Set TextOID as default if can't lookup type
				// TODO log warning
				columnTypes = append(columnTypes, pgtype.TextOID)
			}
		} else {
			// Anonymous pameter present.
			if colName == "boolean" {
				columnTypes = append(columnTypes, pgtype.BoolOID)
			}

			if colName == "blob" {
				columnTypes = append(columnTypes, pgtype.ByteaOID)
			}
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

func ValueToOID(value any) uint32 {
	var oid uint32

	switch value.(type) {
	case int, int64:
		oid = pgtype.Int8OID
		break

	case int16:
		oid = pgtype.Int2OID
		break

	case int32:
		oid = pgtype.Int4OID
		break

	case float32:
		oid = pgtype.Float4OID
		break

	case float64:
		oid = pgtype.Float8OID
		break

	case bool:
		oid = pgtype.BoolOID
		break

	case string:
		oid = pgtype.TextOID
		break

	case []byte:
		oid = pgtype.ByteaOID
		break

	case time.Time:
		oid = pgtype.TimestampOID
		break

	case nil:
		oid = pgtype.UnknownOID
		break

	default: // string
		oid = pgtype.TextOID
	}

	return oid
}

// BytesToValues decodes values from raw bytes based on the PostgreSQL OID type.
/*
func BytesToValues(byteValues [][]byte, oidTypes []uint32) []any {
	valuesLen := len(byteValues)
	oidsLen := len(oidTypes)

	// Validate inputs otherwise not safe proceed.
	if valuesLen == 0 || oidsLen == 0 || (valuesLen != oidsLen) {
		return nil
	}

	values := make([]any, valuesLen)
	for i := range oidTypes {
		switch oidTypes[i] {
		case pgtype.Int2OID, pgtype.Int4OID, pgtype.Int8OID:
			values[i] = int64(big.NewInt(0).SetBytes(byteValues[i]).Uint64())
			continue

		case pgtype.Float8OID, pgtype.NumericOID:
			bits := binary.LittleEndian.Uint64(byteValues[i])
			values[i] = math.Float64frombits(bits)
			continue

		case pgtype.TextOID, pgtype.VarcharOID:
			binds[i] = string(paramValues[i])
			continue

		case pgtype.BoolOID:
			if bytes.Compare(paramValues[i], []byte{0x0}) > 0 {
				binds[i] = bool(true)
			} else {
				binds[i] = bool(false)
			}
			continue

		case pgtype.ByteaOID:
			binds[i] = paramValues[i]
			continue

		case pgtype.DateOID, pgtype.TimestampOID:
			binds[i] = string(paramValues[i])
			continue

		default:
			binds[i] = string(paramValues[i])
		}
	}

	return binds
}

func ValuesToBytes(values []any) [][]byte {
	return nil
}

func ValuesToTypes(values []any) []uint32 {
	return nil
}
*/
