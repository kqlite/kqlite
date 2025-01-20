package pgwire

import (
	"context"
	"log"
	"strings"

	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/parser"

	"github.com/jackc/pgproto3/v2"
)

// Handle the Simple Query protocol.
func (server *DBServer) handleQueryMessage(ctx context.Context, conn *ClientConn, msg *pgproto3.Query) error {
	log.Printf("received query: %q", msg.String)

	// Respond to ping queries.
	if strings.HasPrefix(msg.String, "--") && strings.HasSuffix(msg.String, "ping") {
		return writeMessages(conn,
			&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
			&pgproto3.ReadyForQuery{TxStatus: 'I'})
	}

	// Extract all statements present in the SQL query and do a syntax validation.
	parserResult, err := parser.Parse(msg.String)
	if err != nil {
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'},
		)
	}

	// Convert parser result to database statements.
	var statements []db.Statement
	for _, result := range parserResult {
		statements = append(statements, db.Statement{
			Query:   result.Sql,
			CmdType: convertToStmtCmd(result),
		})
	}

	// Execute all statements part of the SQL query.
	response, err := conn.exeqc.Request(statements)
	if err != nil {
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'},
		)
	}

	// Send response to client for every statement present in the SQL query.
	for _, resp := range response {
		// Handle error from a single statement execution.
		if resp.Error != nil {
			err := writeMessages(conn,
				&pgproto3.ErrorResponse{Message: resp.Error.Error()},
				&pgproto3.ReadyForQuery{TxStatus: 'I'})
			if err != nil {
				return err
			}
		}

		var err error
		var buf []byte
		if resp.Rows != nil {
			defer resp.Rows.Close()
			// Encode result rows to PG wire data rows.
			buf, err = encodeRows(resp.Rows)
			if err != nil {
				return err
			}
			// Send command complete along with the result data.
			buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
		} else {
			// Send the command tag and complete response.
			buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte(resp.CommandTag)}).Encode(buf)
		}

		if _, err := conn.Write(buf); err != nil {
			return err
		}
	}
	// Complete the response with sending 'Ready for Query'.
	return writeMessages(conn, &pgproto3.ReadyForQuery{TxStatus: 'I'})
}
