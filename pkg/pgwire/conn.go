package pgwire

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/parser"
	"github.com/kqlite/kqlite/pkg/util/pgerror"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

// Represents a database client session.
type ClientConn struct {
	net.Conn
	backend *pgproto3.Backend
	db      *db.DB
	exeqc   *db.ExecuteQueryContext

	// Value types encoding and decoding.
	typeMap *pgtype.Map

	// Map of prepared statements for this client session.
	prepStmts map[string]*PreparedStatement

	// Map of prepared portals for this client session.
	portals map[string]*PreparedPortal

	// Forcing to send data in Text format is required when this is a connection from psql client.
	textDataOnly bool
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		completed := time.Since(start)
		if completed.Milliseconds() > 10 {
			fmt.Printf("%s took %v\n", name, completed)
		}
	}
}

func NewClientConn(conn net.Conn) *ClientConn {
	return &ClientConn{
		Conn:         conn,
		backend:      pgproto3.NewBackend(conn, conn),
		prepStmts:    map[string]*PreparedStatement{},
		portals:      map[string]*PreparedPortal{},
		typeMap:      pgtype.NewMap(),
		textDataOnly: false,
	}
}

// Respond to ping queries.
func (conn *ClientConn) handlePing(msg *pgproto3.Query) (bool, error) {
	if strings.HasPrefix(msg.String, "--") && strings.HasSuffix(msg.String, "ping") {
		return true, writeMessages(conn,
			&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
			&pgproto3.ReadyForQuery{TxStatus: 'I'})
	}

	return false, nil
}

// Respond to create database queries.
func (conn *ClientConn) handleCreateDB(msg *pgproto3.Query) (bool, error) {
	if strings.HasPrefix(msg.String, "CREATE DATABASE") {
		return true, writeMessages(conn,
			&pgproto3.CommandComplete{CommandTag: []byte("CREATE DATABASE")},
			&pgproto3.ReadyForQuery{TxStatus: 'I'})
	}

	return false, nil
}

// Handle the Simple Query protocol.
func (conn *ClientConn) handleQuery(ctx context.Context, msg *pgproto3.Query) error {
	defer timer("handleQuery")()

	if handled, err := conn.handlePing(msg); handled || err != nil {
		return err
	}

	if handled, err := conn.handleCreateDB(msg); handled || err != nil {
		return err
	}

	// Rewrite system-information queries so they're tolerable by SQLite.
	query := parser.RewriteQuery(msg.String)
	if msg.String != query {
		// Debug log the rewriten query.
		// log.Printf("query rewrite: %s", query)
	}

	// Extract all statements present in the SQL query and do a syntax validation.
	parserResult, err := parser.Parse(query)
	if err != nil {
		log.Printf("internal parse query error: %s, err: %s\n", query, err.Error())
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'},
		)
	}

	// Convert parser result to database statements.
	var statements []db.Statement
	if len(parserResult) != 0 {
		for _, result := range parserResult {
			statements = append(statements, db.Statement{
				Query:       result.Sql,
				CmdType:     convertToStmtCmd(result),
				ReturnsRows: result.ReturnsRows,
			})
		}
	} else {
		rows, err := conn.db.QueryContext(ctx, query)
		if err != nil {
			log.Printf("execute query: %s, err: %s\n", query, err.Error())
			return writeMessages(conn,
				&pgproto3.ErrorResponse{Message: err.Error()},
				&pgproto3.ReadyForQuery{TxStatus: 'I'})
		}
		defer rows.Close()

		// Encode result rows to PG wire data rows.
		buf, err := encodeRowsNew(rows, conn.typeMap, conn.textDataOnly)
		if err != nil {
			return err
		}

		buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
		if _, err := conn.Write(buf); err != nil {
			return err
		}

		// Send command complete along with the result data.
		return writeMessages(conn, &pgproto3.ReadyForQuery{TxStatus: 'I'})
	}

	// Execute all statements part of the SQL query.
	response, err := conn.exeqc.Request(statements)
	if err != nil {
		log.Printf("execute query, err: %s\n", err.Error())
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'},
		)
	}

	// Send response to client for every statement present in the SQL query.
	for _, resp := range response {
		// Handle error from a single statement execution.
		if resp.Error != nil {
			log.Printf("query %s, execute stmt error: %s\n", query, resp.Error.Error())
			if err := writeMessages(conn, &pgproto3.ErrorResponse{Message: resp.Error.Error()}); err != nil {
				return err
			}
			continue
		}

		var err error
		var buf []byte
		if resp.Rows != nil {
			defer resp.Rows.Close()
			// Encode result rows to PG wire data rows.
			buf, err = encodeRowsNew(resp.Rows, conn.typeMap, conn.textDataOnly)
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

// Handle the Extended Query protocol Close message.
func (conn *ClientConn) handleClose(ctx context.Context, msg *pgproto3.Close) error {
	defer timer("handleClose")()

	switch msg.ObjectType {
	case PrepareStatementType:
		_, found := conn.prepStmts[msg.Name]
		if !found {
			// The spec says "It is not an error to issue Close against a nonexistent
			// statement or portal name". See
			// https://www.postgresql.org/docs/current/static/protocol-flow.html.
			break
		}
		conn.deletePreparedStmt(msg.Name)
	case PreparePortalType:
		_, found := conn.portals[msg.Name]
		if !found {
			break
		}
		conn.deletePortal(msg.Name)
	default:
		return fmt.Errorf("unknown del type: %v", msg.ObjectType)
	}
	return nil
}

// Handle the Extended Query protocol Execute message.
func (conn *ClientConn) handleExecute(ctx context.Context, msg *pgproto3.Execute) error {
	defer timer("handleExecute")()

	portalName := msg.Portal
	portal, found := conn.portals[portalName]
	if !found {
		return pgerror.New(
			pgerrcode.InvalidCursorName, fmt.Sprintf("unknown portal %q", portalName))
	}

	stmt := *portal.Prepared.Stmt
	stmt.Parameters = portal.Qargs
	statements := []db.Statement{stmt}

	// Execute SQL statement.
	response, err := conn.exeqc.Request(statements)
	if err != nil {
		log.Printf("Error from query %s\n", err.Error())
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'})
	}

	resp := response[0]
	// Handle error from the statement execution.
	if resp.Error != nil {
		log.Printf("Error from statement %s\n", resp.Error.Error())

		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: resp.Error.Error(), Code: pgerror.GetPGCode(err), ConstraintName: "kine_pkey"},
			&pgproto3.ReadyForQuery{TxStatus: 'I'})
	}

	var buf []byte
	if resp.Rows != nil {
		defer resp.Rows.Close()

		// Encode result rows to PG wire data rows.
		// buf, err = encodeRows(resp.Rows)
		buf, err = encodeRowsNew(resp.Rows, conn.typeMap, conn.textDataOnly)
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

	if _, err := conn.Write(buf); err != nil {
		return err
	}

	return nil
}

// Handle the Extended Query protocol Sync message.
func (conn *ClientConn) handleBind(ctx context.Context, msg *pgproto3.Bind) error {
	// Start a timer DEBUG
	defer timer("handleBind")()

	prepared, found := conn.prepStmts[msg.PreparedStatement]
	if !found {
		fmt.Printf("prepared statement %q does not exist\n", msg.PreparedStatement)
		return pgerror.New(
			pgerrcode.InvalidSQLStatementName, fmt.Sprintf("prepared statement %q does not exist", msg.PreparedStatement))
	}

	portalName := msg.DestinationPortal
	if portalName != "" {
		if _, ok := conn.portals[portalName]; ok {
			fmt.Printf("portal %q already exists \n", portalName)
			return pgerror.New(
				pgerrcode.DuplicateCursor, fmt.Sprintf("portal %q already exists", portalName))
		}
	} else {
		// Deallocate the unnamed portal, if it exists.
		conn.deletePortal("")
	}

	// Decode parameters values for the target statement.
	params := parametersToValues(msg.Parameters, prepared.ParamOIDs)

	// Bind portal with statement parameters.
	if err := conn.addPortal(portalName, prepared, params); err != nil {
		return err
	}

	// Send back the response message.
	if err := writeMessages(conn, &pgproto3.BindComplete{}); err != nil {
		return err
	}

	// TODO log
	//if log.V(2) {
	//	log.Infof(ctx, "portal: %q for %q, args %q, formats %q",
	//		portalName, ps.Statement, qargs, columnFormatCodes)
	//}

	return nil
}

// Handle the Extended Query protocol Sync message.
func (conn *ClientConn) handleSync(ctx context.Context, msg *pgproto3.Sync) error {
	defer timer("handleSync")()

	if err := writeMessages(conn, &pgproto3.ReadyForQuery{TxStatus: 'I'}); err != nil {
		return err
	}

	return nil
}

// Send Row descripton if present, used mainly for returning results on Describe message.
func writePreparedRowDescription(conn *ClientConn, prepared *PreparedStatement) error {
	if conn == nil || prepared == nil {
		return nil
	}

	// Send Row description if available
	if prepared.Stmt.ReturnsRows {
		if len(prepared.Fields) != 0 {
			buf, _ := toRowDescription(prepared.Fields).Encode(nil)
			if _, err := conn.Write(buf); err != nil {
				return err
			}
		} else {
			// No information present for the rows send empty row description.
			empty, _ := (&pgproto3.RowDescription{}).Encode(nil)
			if _, err := conn.Write(empty); err != nil {
				return err
			}
		}
	} else {
		// Statement is not returning Rows send NoData.
		if err := writeMessages(conn, &pgproto3.NoData{}); err != nil {
			return err
		}
	}

	return nil
}

// Handle the Extended Query protocol Describe message.
func (conn *ClientConn) handleDescribe(ctx context.Context, msg *pgproto3.Describe) error {
	defer timer("handleDescribe")()

	switch msg.ObjectType {
	case PrepareStatementType:
		prepared, ok := conn.prepStmts[msg.Name]
		if !ok {
			return pgerror.New(
				pgerrcode.InvalidSQLStatementName, fmt.Sprintf("prepared statement %q does not exist", msg.Name))
		}

		// Send parameter description.
		if err := writeMessages(conn, &pgproto3.ParameterDescription{ParameterOIDs: prepared.ParamOIDs}); err != nil {
			return err
		}

		// Send Row descripton if present.
		if err := writePreparedRowDescription(conn, prepared); err != nil {
			return nil
		}

	case PreparePortalType:
		portal, ok := conn.portals[msg.Name]
		if !ok {
			return pgerror.New(
				pgerrcode.InvalidCursorName, fmt.Sprintf("unknown portal %q", msg.Name))
		}

		// Send parameter description, followed by a Row descripton if present.
		if err := writeMessages(conn, &pgproto3.ParameterDescription{ParameterOIDs: portal.Prepared.ParamOIDs}); err != nil {
			return err
		}

		// Send Row descripton if present.
		if err := writePreparedRowDescription(conn, portal.Prepared); err != nil {
			return nil
		}

	default:
		return pgerror.New(
			pgerrcode.ProtocolViolation, fmt.Sprintf("invalid DESCRIBE message subtype %x", msg.ObjectType),
		)
	}

	return nil
}

// Handle the Extended Query protocol Parse message.
func (conn *ClientConn) handleParse(ctx context.Context, msg *pgproto3.Parse) error {
	defer timer("handleParse")()

	// Rewrite system-information queries so they're tolerable by SQLite.
	query := parser.RewriteQuery(msg.Query)

	//if msg.Query != query {
	// Debug log the rewriten query.
	//	log.Printf("query rewrite: %s", query)
	//}

	// Validate syntax and extract statements parameter names.
	parserResult, err := parser.Parse(query)
	if err != nil {
		log.Printf("Error parsing query %s\n", query)
		return err
	}

	// The query string contained in a Parse message cannot include more than one SQL statement;
	// else a syntax error is reported.
	if len(parserResult) != 1 {
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: "Wrong number of prepared statements or invalid statement",
				Code: pgerrcode.InvalidPreparedStatementDefinition})
	}

	// Convert parser result to database statement.
	stmt := &db.Statement{
		Query:       parserResult[0].Sql,
		CmdType:     convertToStmtCmd(parserResult[0]),
		ReturnsRows: parserResult[0].ReturnsRows,
	}

	// Check if Parse message contains any paremter type hints.
	var paramTypes []uint32
	if len(msg.ParameterOIDs) == 0 {
		var err error
		// Extract statement parameters located in the query text.
		paramTypes, err = db.LookupTypeInfo(ctx, conn.db, parserResult[0].ArgColumns, parserResult[0].Tables)
		if err != nil {
			return err
		}
	} else {
		paramTypes = msg.ParameterOIDs
	}

	// fmt.Printf("paramTypes %v\n", paramTypes)

	// Create prepare statement and add it to cache.
	if _, err := conn.addPreparedStatement(msg.Name, stmt, paramTypes); err != nil {
		if _ = writeMessages(conn, &pgproto3.ErrorResponse{
			Message: err.Error(), Code: pgerror.GetPGCode(err)}); err != nil {
			return err
		}
	}

	// Parsing complete.
	return writeMessages(conn, &pgproto3.ParseComplete{})
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
