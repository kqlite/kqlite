package command

import (
	"github.com/kqlite/kqlite/pkg/parser"
	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type SQLCommandType string

// Common SQL commands.
const (
	SELECT   SQLCommandType = "SELECT"
	UPDATE   SQLCommandType = "UPDATE"
	INSERT   SQLCommandType = "INSERT"
	DELETE   SQLCommandType = "DELETE"
	BEGIN    SQLCommandType = "BEGIN"
	COMMIT   SQLCommandType = "COMMIT"
	ROLLBACK SQLCommandType = "ROLLBACK"
	UNKNOWN  SQLCommandType = "UNKNOWN"
)

// Convert parser command types to a common SQL statement command.
func ConvertToStmtCmd(stmtResult parser.ParserStmtResult) SQLCommandType {
	// Transaction commands.
	if stmtResult.TxCmd != pg_query.TransactionStmtKind_TRANSACTION_STMT_KIND_UNDEFINED {
		switch stmtResult.TxCmd {
		case pg_query.TransactionStmtKind_TRANS_STMT_BEGIN:
			return BEGIN

		case pg_query.TransactionStmtKind_TRANS_STMT_COMMIT:
			return COMMIT

		case pg_query.TransactionStmtKind_TRANS_STMT_ROLLBACK:
			return ROLLBACK
		}
	}
	// Common SQL commands.
	if stmtResult.SqlCmd != pg_query.CmdType_CMD_TYPE_UNDEFINED {
		switch stmtResult.SqlCmd {
		case pg_query.CmdType_CMD_SELECT:
			return SELECT

		case pg_query.CmdType_CMD_INSERT:
			return INSERT

		case pg_query.CmdType_CMD_DELETE:
			return DELETE

		case pg_query.CmdType_CMD_UPDATE:
			return UPDATE
		}
	}
	return ""
}
