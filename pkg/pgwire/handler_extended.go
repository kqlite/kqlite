package pgwire

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgproto3/v2"

	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/parser"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// Handle the Extended Query protocol.
func (server *DBServer) handleParseMessage(ctx context.Context, conn *ClientConn, pmsg *pgproto3.Parse) error {
	// Rewrite system-information queries so they're tolerable by SQLite.
	query := parser.RewriteQuery(pmsg.Query)

	if pmsg.Query != query {
		log.Printf("query rewrite: %s", query)
	}

	// Validate syntax and extract statements parameter names.
	parserResult, err := parser.Parse(query)
	if err != nil {
		return err
	}

	// Convert parser result to database statements and execute every statement.
	// The query string contained in a Parse message cannot include more than one SQL statement;
	// else a syntax error is reported.
	statements := []db.Statement{
		{
			Query:   parserResult[0].Sql,
			CmdType: convertToStmtCmd(parserResult[0]),
		},
	}

	// Extract parameters types the statement in the query.
	paramTypes, err := db.LookupTypeInfo(ctx, conn.db, parserResult[0].ArgColumns, parserResult[0].Tables)
	if err != nil {
		return err
	}

	// Parsing complete.
	if err := writeMessages(conn, &pgproto3.ParseComplete{}); err != nil {
		return err
	}

	var response []db.ExecuteQueryResponse
	// message loop.
	for {
		msg, err := conn.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message during parse: %w", err)
		}

		log.Printf("[recv(p)] %#v", msg)

		switch msg := msg.(type) {
		case *pgproto3.Bind:
			// Extract parameters values for the target statement.
			statements[0].Parameters = parametersToValues(msg.Parameters, paramTypes)

			// Send back the response message.
			if err := writeMessages(conn, &pgproto3.BindComplete{}); err != nil {
				return err
			}
			break

		case *pgproto3.Describe:
			writeMessages(conn, &pgproto3.ParameterDescription{ParameterOIDs: paramTypes})
			break

		case *pgproto3.Execute:
			// Execute all statements part of the SQL query.
			var err error
			response, err = conn.exeqc.Request(statements)
			if err != nil {
				return writeMessages(conn,
					&pgproto3.ErrorResponse{Message: err.Error()},
					&pgproto3.ReadyForQuery{TxStatus: 'I'})
			}

			resp := response[0]
			// Handle error from the statement execution.
			if resp.Error != nil {
				err := writeMessages(conn,
					&pgproto3.ErrorResponse{Message: resp.Error.Error()},
					&pgproto3.ReadyForQuery{TxStatus: 'I'})
				if err != nil {
					return err
				}
			}

			var buf []byte
			if resp.Rows != nil {
				defer resp.Rows.Close()

				// Encode result rows to PG wire data rows.
				buf, err = encodeRows(resp.Rows)
				if err != nil {
					return err
				}
				// Create buufer with command complete with resulting rows.
				buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
			} else {
				// Send NoData along the way.
				buf, _ = (&pgproto3.NoData{}).Encode(nil)
				buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte(response[0].CommandTag)}).Encode(buf)
			}
			// Send result back
			if _, err := conn.Write(buf); err != nil {
				return err
			}
			break

		case *pgproto3.Sync:
			err := writeMessages(conn, &pgproto3.ReadyForQuery{TxStatus: 'I'})
			// Statement is executed, done with this extended query session.
			if len(response) != 0 {
				return err
			}
			break

		case *pgproto3.Close:
			// TODO
			return nil
		default:
			return fmt.Errorf("unexpected message type during parse: %#v", msg)
		}
	}
}

// Will convert parser commands types to database statement commands.
func convertToStmtCmd(stmtResult parser.ParserStmtResult) db.SqlCmdType {
	// Transaction commands.
	if stmtResult.TxCmd != pg_query.TransactionStmtKind_TRANSACTION_STMT_KIND_UNDEFINED {
		switch stmtResult.TxCmd {
		case pg_query.TransactionStmtKind_TRANS_STMT_BEGIN:
			return db.CMD_BEGIN

		case pg_query.TransactionStmtKind_TRANS_STMT_COMMIT:
			return db.CMD_COMMIT

		case pg_query.TransactionStmtKind_TRANS_STMT_ROLLBACK:
			return db.CMD_ROLLBACK
		}
	}
	// Common SQL commands.
	if stmtResult.SqlCmd != pg_query.CmdType_CMD_TYPE_UNDEFINED {
		switch stmtResult.SqlCmd {
		case pg_query.CmdType_CMD_SELECT:
			return db.CMD_SELECT

		case pg_query.CmdType_CMD_INSERT:
			return db.CMD_INSERT

		case pg_query.CmdType_CMD_DELETE:
			return db.CMD_DELETE

		case pg_query.CmdType_CMD_UPDATE:
			return db.CMD_UPDATE
		}
	}
	return ""
}
