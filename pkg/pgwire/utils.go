package pgwire

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/kqlite/kqlite/pkg/db"
)

func toRowDescription(cols []*sql.ColumnType) *pgproto3.RowDescription {
	var desc pgproto3.RowDescription
	for _, col := range cols {
		var typeOID uint32
		dbType := col.DatabaseTypeName()
		if pgColType, exists := db.Typemap()[dbType]; exists {
			typeOID = pgColType
		} else {
			typeOID = pgtype.TextOID
		}

		typeSize, ok := col.Length()
		if !ok {
			typeSize = -1
		}

		desc.Fields = append(desc.Fields, pgproto3.FieldDescription{
			Name:                 []byte(col.Name()),
			TableOID:             0,
			TableAttributeNumber: 0,
			DataTypeOID:          typeOID,
			DataTypeSize:         int16(typeSize),
			TypeModifier:         -1,
			Format:               0,
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
		row.Values[i] = []byte(fmt.Sprint(values[i]))
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

func parametersToValues(paramValues [][]byte, paramTypes []uint32) []interface{} {
	if len(paramValues) == 0 || len(paramValues) == 0 {
		return nil
	}

	if len(paramValues) != len(paramTypes) {
		return nil
	}

	binds := make([]interface{}, len(paramValues))
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
			binds[i] = paramValues
			continue

		case pgtype.DateOID, pgtype.TimestampOID:
			binds[i] = string(paramValues[i])
			continue
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
