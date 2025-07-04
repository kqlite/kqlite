package pgwire

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kqlite/kqlite/pkg/db"
)

func toRowDescription(cols []*sql.ColumnType) *pgproto3.RowDescription {
	var desc pgproto3.RowDescription
	for _, col := range cols {
		var typeOID uint32
		format := 1
		dbType := col.DatabaseTypeName()

		if pgColType, exists := db.Typemap()[dbType]; exists {
			typeOID = pgColType
		} else {
			typeOID = pgtype.Int8OID
			//typeOID = pgtype.TextOID
		}

		if typeOID == pgtype.TextOID {
			format = 0
		}

		typeSize, ok := col.Length()
		if !ok {
			typeSize = -1
		}

		colName := col.Name()
		if len(colName) > 20 {
			colName = colName[:20]
		}

		r := strings.NewReplacer(" ", "", "\n", "", "\t", "", ")", "", "(", "", ",", "", ".", "")
		colName = r.Replace(colName)

		desc.Fields = append(desc.Fields, pgproto3.FieldDescription{
			Name:                 []byte(colName),
			TableOID:             0,
			TableAttributeNumber: 0,
			DataTypeOID:          typeOID,
			DataTypeSize:         int16(typeSize),
			TypeModifier:         -1,
			Format:               int16(format),
		})
	}

	return &desc
}

func scanRow(rows *sql.Rows, cols []*sql.ColumnType) (*pgproto3.DataRow, error) {
	refs := make([]interface{}, len(cols))
	values := make([]interface{}, len(cols))
	for i := range refs {
		refs[i] = &values[i]
	}

	// Scan from SQLite database.
	if err := rows.Scan(refs...); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	// Convert to TEXT values to return over Postgres wire protocol.
	row := pgproto3.DataRow{Values: make([][]byte, len(values))}
	for i := range values {
		//row.Values[i] = []byte(fmt.Sprint(values[i]))

		if i == 3 {
			// TEXT
			row.Values[i] = []byte(fmt.Sprint(values[i]))
			continue
		}

		if i == 9 || i == 10 {
			// Byte array
			row.Values[i], _ = values[i].([]byte)
			continue
		}

		// Int64
		buf := []byte{}
		buf = append(buf, 0, 0, 0, 0, 0, 0, 0, 0)
		n, _ := values[i].(int64)
		binary.BigEndian.PutUint64(buf[0:], uint64(n))
		row.Values[i] = buf
	}

	return &row, nil
}

// Convert SQL database rows to PG DataRow's.
func encodeRows(rows *sql.Rows) ([]byte, error) {
	// Encode column header.
	cols, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("column types: %w", err)
	}

	buf, _ := toRowDescription(cols).Encode(nil)

	// Iterate over each row and encode it to the wire protocol.
	for rows.Next() {
		row, err := scanRow(rows, cols)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		buf, _ = row.Encode(buf)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}
	return buf, nil
}

func toRowDescriptionNew(cols []*sql.ColumnType, oids []uint32, textDataOnly bool) *pgproto3.RowDescription {
	if len(cols) != len(oids) {
		return nil
	}

	var desc pgproto3.RowDescription
	for idx, col := range cols {
		format := pgtype.BinaryFormatCode

		typeSize, ok := col.Length()
		if !ok {
			typeSize = -1
		}

		colName := col.Name()
		typeOID := oids[idx]

		if textDataOnly || typeOID == pgtype.TextOID {
			typeOID = pgtype.TextOID
			format = pgtype.TextFormatCode
		}

		desc.Fields = append(desc.Fields, pgproto3.FieldDescription{
			Name:                 []byte(colName),
			TableOID:             0,
			TableAttributeNumber: 0,
			DataTypeOID:          typeOID,
			DataTypeSize:         int16(typeSize),
			TypeModifier:         -1,
			Format:               int16(format),
		})
	}

	return &desc
}

func scanRowNew(rows *sql.Rows, cols []*sql.ColumnType, typeMap *pgtype.Map, oids *[]uint32, textDataOnly bool) (*pgproto3.DataRow, error) {
	refs := make([]interface{}, len(cols))
	values := make([]interface{}, len(cols))
	for i := range refs {
		refs[i] = &values[i]
	}

	// Scan from SQLite database.
	if err := rows.Scan(refs...); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	// Encode values to bytes to return over Postgres wire protocol.
	row := pgproto3.DataRow{Values: make([][]byte, len(values))}
	for i := range values {
		// Populate OID's when scanning rows.
		if len(*oids) < len(cols) {
			*oids = append(*oids, db.ValueToOID(values[i]))
		}

		var err error
		var buf []byte
		if !textDataOnly {
			buf, err = typeMap.Encode((*oids)[i], pgtype.BinaryFormatCode, values[i], nil)
		} else {
			buf, err = typeMap.Encode((*oids)[i], pgtype.TextFormatCode, values[i], nil)
		}

		if err != nil {
			return nil, err
		}

		// TODO
		row.Values[i] = buf
	}

	return &row, nil
}

func encodeRowsNew(rows *sql.Rows, typeMap *pgtype.Map, textDataOnly bool) ([]byte, error) {
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	// Encode column header.
	cols, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("column types: %w", err)
	}

	var buf []byte
	oids := []uint32{}
	var rowDescr *pgproto3.RowDescription
	// Iterate over each row and encode it to the wire protocol.
	for rows.Next() {
		row, err := scanRowNew(rows, cols, typeMap, &oids, textDataOnly)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		// Generate row description using values and column info.
		if rowDescr == nil {
			rowDescr = toRowDescriptionNew(cols, oids, textDataOnly)
			buf, _ = rowDescr.Encode(nil)
		}

		buf, _ = row.Encode(buf)
	}

	return buf, nil
}

func parametersToValues(paramValues [][]byte, paramTypes []uint32) []any {
	if len(paramValues) == 0 || len(paramTypes) == 0 {
		return nil
	}

	if len(paramValues) != len(paramTypes) {
		return nil
	}

	binds := make([]any, len(paramValues))
	for i := range paramValues {
		switch paramTypes[i] {
		case pgtype.Int2OID, pgtype.Int4OID, pgtype.Int8OID, pgtype.NumericOID:
			binds[i] = int64(big.NewInt(0).SetBytes(paramValues[i]).Uint64())
			continue

		case pgtype.Float8OID:
			bits := binary.LittleEndian.Uint64(paramValues[i])
			binds[i] = math.Float64frombits(bits)
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

// writeMessages writes/packages all messages to a single buffer before sending.
func writeMessages(w io.Writer, msgs ...pgproto3.Message) error {
	var buf []byte
	for _, msg := range msgs {
		buf, _ = msg.Encode(buf)
	}
	_, err := w.Write(buf)
	return err
}

func getParameter(m map[string]string, k string) string {
	if m == nil {
		return ""
	}
	return m[k]
}

// Initialize virtual table catalog.
func initCatatog(ctx context.Context, dbase *db.Database) error {
	// Attach an in-memory database for pg_catalog.
	if _, err := dbase.ExecContext(ctx, `ATTACH ':memory:' AS pg_catalog`); err != nil {
		// Already attached do nothing.
		if err.Error() == "database pg_catalog is already in use" {
			return nil
		}
		return fmt.Errorf("attach pg_catalog: %w", err)
	}

	// Register virtual tables to imitate postgres.
	if _, err := dbase.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_database USING pg_database_module (oid, datname, datdba, encoding, datcollate, datctype, datistemplate, datallowconn, datconnlimit, datlastsysoid, datfrozenxid, datminmxid, dattablespace, datacl)"); err != nil {
		return fmt.Errorf("create pg_database: %w", err)
	}
	if _, err := dbase.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_namespace USING pg_namespace_module (oid, nspname, nspowner, nspacl)"); err != nil {
		return fmt.Errorf("create pg_namespace: %w", err)
	}
	if _, err := dbase.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_description USING pg_description_module (objoid, classoid, objsubid, description)"); err != nil {
		return fmt.Errorf("create pg_description: %w", err)
	}
	if _, err := dbase.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_settings USING pg_settings_module (name, setting, unit, category, short_desc, extra_desc, context, vartype, source, min_val, max_val, enumvals, boot_val, reset_val, sourcefile, sourceline, pending_restart)"); err != nil {
		return fmt.Errorf("create pg_settings: %w", err)
	}
	if _, err := dbase.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_type USING pg_type_module (oid, typname, typnamespace, typowner, typlen, typbyval, typtype, typcategory, typispreferred, typisdefined, typdelim, typrelid, typelem, typarray, typinput, typoutput, typreceive, typsend, typmodin, typmodout, typanalyze, typalign, typstorage, typnotnull, typbasetype, typtypmod, typndims, typcollation, typdefaultbin, typdefault, typacl)"); err != nil {
		return fmt.Errorf("create pg_type: %w", err)
	}
	if _, err := dbase.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_class USING pg_class_module (oid, relname, relnamespace, reltype, reloftype, relowner, relam, relfilenode, reltablespace, relpages, reltuples, relallvisible, reltoastrelid, relhasindex, relisshared, relpersistence, relkind, relnatts, relchecks, relhasrules, relhastriggers, relhassubclass, relrowsecurity, relforcerowsecurity, relispopulated, relreplident, relispartition, relrewrite, relfrozenxid, relminmxid, relacl, reloptions, relpartbound)"); err != nil {
		return fmt.Errorf("create pg_class: %w", err)
	}
	if _, err := dbase.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS pg_catalog.pg_range USING pg_range_module (rngtypid, rngsubtype, rngmultitypid, rngcollation, rngsubopc, rngcanonical, rngsubdiff)"); err != nil {
		return fmt.Errorf("create pg_range: %w", err)
	}

	return nil
}
