package parser

import (
	//"fmt"
	"slices"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

type parserStmtWalker struct {
	result ParserStmtResult
	// For SELECT, DELETE and UPDATE statements arguments are extracted from the SQL query expressions.
	exprLocation  int      // Unique Location of the expression found in the SQL statement.
	exprColumns   []string // Expression condition columns extracted.
	boolExprFound bool
	insertStmt    bool     // INSERT statement node located.
	updateStmt    bool     // UPDATE statement node located.
	targetColumns []string // INSERT,UPDATE statement target expression columns extracted.
}

type ParserStmtResult struct {
	Tables      []string                     // Tables referenced in the statement.
	Sql         string                       // Statement SQL text.
	ArgColumns  []string                     // Statement referenced columns as params/arguments names.
	SqlCmd      pg_query.CmdType             // SQL command. (Ex. SELECT, INSERT, UPDATE, DELETE ...)
	TxCmd       pg_query.TransactionStmtKind // Transaction command (Ex. BEGIN, COMMIT, ROLLBACK...)
	ReturnsRows bool                         // Indicates whether statement returns rows (ex. RETURNING clause in INSERT, UPDATE, DELETE ..)
}

func (walker *parserStmtWalker) getTableName(rangevar *pg_query.RangeVar) {
	if rangevar != nil {
		relname := rangevar.GetRelname()
		if relname != "" {
			if !slices.Contains(walker.result.Tables, relname) {
				walker.result.Tables = append(walker.result.Tables, relname)
			}
		}
	}
}

// Set the corresponding SQL statement type/command.
func (walker *parserStmtWalker) setCommand(cmd pg_query.CmdType) {
	if walker.result.SqlCmd == pg_query.CmdType_CMD_TYPE_UNDEFINED {
		walker.result.SqlCmd = cmd
		walker.result.TxCmd = pg_query.TransactionStmtKind_TRANSACTION_STMT_KIND_UNDEFINED
	}
}

func (walker *parserStmtWalker) Visit(node *pg_query.Node) (v Visitor, err error) {
	switch n := node.Node.(type) {
	case *pg_query.Node_TransactionStmt:
		switch n.TransactionStmt.Kind {
		case pg_query.TransactionStmtKind_TRANS_STMT_BEGIN:
			walker.result.TxCmd = pg_query.TransactionStmtKind_TRANS_STMT_BEGIN
			walker.result.SqlCmd = pg_query.CmdType_CMD_TYPE_UNDEFINED
			break

		case pg_query.TransactionStmtKind_TRANS_STMT_COMMIT:
			walker.result.TxCmd = pg_query.TransactionStmtKind_TRANS_STMT_COMMIT
			walker.result.SqlCmd = pg_query.CmdType_CMD_TYPE_UNDEFINED
			break

		case pg_query.TransactionStmtKind_TRANS_STMT_ROLLBACK:
			walker.result.TxCmd = pg_query.TransactionStmtKind_TRANS_STMT_ROLLBACK
			walker.result.SqlCmd = pg_query.CmdType_CMD_TYPE_UNDEFINED
			break
		}
		break

	case *pg_query.Node_SelectStmt:
		walker.setCommand(pg_query.CmdType_CMD_SELECT)
		break

	case *pg_query.Node_InsertStmt:
		walker.insertStmt = true
		walker.getTableName(n.InsertStmt.GetRelation())
		walker.setCommand(pg_query.CmdType_CMD_INSERT)
		walker.result.ReturnsRows = (len(n.InsertStmt.GetReturningList()) != 0)
		break

	case *pg_query.Node_DeleteStmt:
		walker.getTableName(n.DeleteStmt.GetRelation())
		walker.setCommand(pg_query.CmdType_CMD_DELETE)
		walker.result.ReturnsRows = (len(n.DeleteStmt.GetReturningList()) != 0)
		break

	case *pg_query.Node_UpdateStmt:
		walker.updateStmt = true
		walker.getTableName(n.UpdateStmt.GetRelation())
		walker.setCommand(pg_query.CmdType_CMD_UPDATE)
		walker.result.ReturnsRows = (len(n.UpdateStmt.GetReturningList()) != 0)
		break

	case *pg_query.Node_RangeVar:
		walker.getTableName(n.RangeVar)
		break

	case *pg_query.Node_AExpr:
		// Found expression in the SQL query, init relevant fields.
		if walker.exprLocation == 0 {
			walker.exprLocation = int(n.AExpr.GetLocation())
		}
		break

	case *pg_query.Node_BoolExpr:
		walker.boolExprFound = true
		break

	case *pg_query.Node_ColumnRef:
		if walker.exprLocation != 0 {
			// Collect referenced columns/fields from current expression.
			fields := n.ColumnRef.GetFields()
			for _, fieldn := range fields {
				if fieldn == nil || fieldn.Node == nil {
					continue
				}
				field, ok := fieldn.Node.(*pg_query.Node_String_)
				if ok && field != nil {
					walker.exprColumns = append(walker.exprColumns, field.String_.GetSval())
				}
			}
			break
		}

	case *pg_query.Node_ParamRef:
		if walker.exprLocation != 0 && len(walker.exprColumns) != 0 {
			walker.result.ArgColumns = append(walker.result.ArgColumns, walker.exprColumns[len(walker.exprColumns)-1])
			break
		}

		if (walker.insertStmt || walker.updateStmt) && len(walker.targetColumns) != 0 {
			paramRef := n.ParamRef.GetNumber()
			// Check if column has a corresponding parameter entry in expression.
			if len(walker.targetColumns) >= int(paramRef) {
				walker.result.ArgColumns = append(walker.result.ArgColumns, walker.targetColumns[paramRef-1])
			}
			break
		}

		if walker.boolExprFound {
			walker.result.ArgColumns = append(walker.result.ArgColumns, "boolean")
		} else {
			walker.result.ArgColumns = append(walker.result.ArgColumns, "blob")
		}

	case *pg_query.Node_ResTarget:
		if walker.insertStmt || walker.updateStmt {
			var name string
			// Check for column indirection in the ResTarget expression,
			// (ex. table_name.column_name instead of just column_name)
			indirection := n.ResTarget.GetIndirection()
			if len(indirection) != 0 {
				n := indirection[len(indirection)-1]

				field, ok := n.Node.(*pg_query.Node_String_)
				if ok && field != nil {
					name = field.String_.GetSval()
				}
			}
			// Fallback if no indirection or no name found in the indirection.
			if name == "" {
				name = n.ResTarget.GetName()
			}

			if name != "" {
				walker.targetColumns = append(walker.targetColumns, name)
			}
			break
		}
	}
	return walker, err
}

func (walker *parserStmtWalker) VisitEnd(node *pg_query.Node) error {
	switch n := node.Node.(type) {
	case *pg_query.Node_AExpr:
		// Clear expression info on node exit.
		if walker.exprLocation == int(n.AExpr.GetLocation()) {
			walker.exprLocation = 0
			walker.exprColumns = []string{}
		}
		break

	case *pg_query.Node_BoolExpr:
		walker.boolExprFound = false
		break

	}
	return nil
}

// Extract single statements from SQL query text.
func (walker *parserStmtWalker) getStatement(sql string, stmt *pg_query.RawStmt) {
	if sql != "" {
		stmloc := stmt.GetStmtLocation()
		stmlen := stmt.GetStmtLen()

		// unknown start location.
		if stmloc == -1 {
			walker.result.Sql = sql
			return
		}

		// length in bytes; 0 means "rest of string"
		if stmlen == 0 {
			stmlen = (int32(len(sql)) - stmloc)
		}

		// validate position and length against the SQL text.
		if int(stmloc+stmlen) <= len(sql) {
			walker.result.Sql = strings.TrimSpace(sql[stmloc : stmloc+stmlen])
		}
	}
}

// Parser supports only the most common queries used for manipulating data.
// Special queries are not handled at all.
func isSpecialQuery(sql string) bool {
	q := strings.ToUpper(strings.TrimSpace(sql))
	if strings.HasPrefix(q, "SELECT") ||
		strings.HasPrefix(q, "INSERT") ||
		strings.HasPrefix(q, "UPDATE") ||
		strings.HasPrefix(q, "DELETE") ||
		strings.HasPrefix(q, "WITH") ||
		strings.HasPrefix(q, "BEGIN") ||
		strings.HasPrefix(q, "COMMIT") ||
		strings.HasPrefix(q, "END") ||
		strings.HasPrefix(q, "ROLLBACK") {
		return false
	}

	return true
}

// Parse a SQL query string, can have multiple statements.
func Parse(sql string) ([]ParserStmtResult, error) {
	var result []ParserStmtResult
	if sql == "" {
		return result, nil
	}

	if isSpecialQuery(sql) {
		return result, nil
	}

	tree, err := pg_query.Parse(sql)
	if err != nil {
		return result, err
	}

	for _, raw := range tree.Stmts {
		if st := raw.GetStmt(); st != nil {
			walker := &parserStmtWalker{}
			if err := Walk(walker, st); err != nil {
				return result, err
			}
			walker.getStatement(sql, raw)
			result = append(result, walker.result)
		}
	}
	return result, nil
}
