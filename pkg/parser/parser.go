package parser

import (
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

type parserStmtWalker struct {
	result ParserStmtResult
	// For SELECT, DELETE and UPDATE statements arguments are extracted from the SQL query expressions.
	exprLocation  int      // Unique Location of the expression found in the SQL statement.
	exprColumns   []string // Expression columns extracted.
	insertStmt    bool     // INSERT statement node located.
	insertColumns []string // INSERT statement columns extracted.
}

type ParserStmtResult struct {
	Args   []string // Statement params/arguments.
	Tables []string // Tables referenced in the statement.
}

func (walker *parserStmtWalker) getTableName(rangevar *pg_query.RangeVar) {
	if rangevar != nil {
		relname := rangevar.GetRelname()
		if relname != "" {
			walker.result.Tables = append(walker.result.Tables, relname)
		}
	}
}

func (walker *parserStmtWalker) Visit(node *pg_query.Node) (v Visitor, err error) {
	switch n := node.Node.(type) {
	case *pg_query.Node_InsertStmt:
		walker.insertStmt = true
		walker.getTableName(n.InsertStmt.GetRelation())
		break
	case *pg_query.Node_DeleteStmt:
		walker.getTableName(n.DeleteStmt.GetRelation())
		break
	case *pg_query.Node_UpdateStmt:
		walker.getTableName(n.UpdateStmt.GetRelation())
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
			walker.result.Args = append(walker.result.Args, walker.exprColumns[len(walker.exprColumns)-1])
			break
		}
		if walker.insertStmt && len(walker.insertColumns) != 0 {
			// Check if column has a corresponding parameter entry in expression.
			number := n.ParamRef.GetNumber()
			if len(walker.insertColumns) >= int(number) {
				walker.result.Args = append(walker.result.Args, walker.insertColumns[number-1])
			}
			break
		}
	case *pg_query.Node_ResTarget:
		if walker.insertStmt {
			name := n.ResTarget.GetName()
			if name != "" {
				walker.insertColumns = append(walker.insertColumns, name)
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
	case *pg_query.Node_InsertStmt:
		// Clear INSERT data in case of a subsequent inserts.
		walker.insertStmt = false
		walker.insertColumns = []string{}
	}
	return nil
}

// Parse a SQL query string, can have multiple statements.
func Parse(sql string) ([]ParserStmtResult, error) {
	var result []ParserStmtResult
	if sql == "" {
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
			result = append(result, walker.result)
		}
	}
	return result, nil
}
