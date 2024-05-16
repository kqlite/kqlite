package parser

import (
	//"encoding/json"
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// Reflect an SQL statement.
type Statement struct {
	Type   string            // Type of statement ex.(SELECT, UPDATE, DELETE)
	Params map[string]string // Statement parameters associated with a source/target table.
}

func getWhereClause(statement string, result *pg_query.ParseResult) *pg_query.Node {
	if statement != "" && result != nil {
		switch statement {
		case "SELECT":
			return result.Stmts[0].Stmt.GetSelectStmt().GetWhereClause()
		case "UPDATE":
			return result.Stmts[0].Stmt.GetUpdateStmt().GetWhereClause()
		case "DELETE":
			return result.Stmts[0].Stmt.GetDeleteStmt().GetWhereClause()
		}
	}
	return nil
}

func QueryToStatement(query string) (Statement, error) {
	stmt := Statement{}

	splits := strings.SplitN(query, " ", 2)
	if splits == nil {
		return stmt, fmt.Errorf("Invalid input argument for query")
	}
	stmt.Type = strings.ToUpper(splits[0])

	query, err := pg_query.Normalize(query)
	if err != nil {
		return stmt, fmt.Errorf("Error normalize query: %s, %w", query, err)
	}
	result, err := pg_query.Parse(query)
	if err != nil {
		return stmt, fmt.Errorf("Error parsing the query: %s, %w", query, err)
	}

	//fmt.Printf("Statemnt: %#v\n", result.Stmts[0].Stmt.GetSelectStmt().GetWhereClause().GetAExpr())
	if len(result.Stmts) != 0 {
		// Extract Where clause and get the statement parameters/arguments.
		if node := getWhereClause(stmt.Type, result); node != nil {
			var fields []*pg_query.Node
			if expr := node.GetAExpr(); expr != nil {
				fields = expr.GetLexpr().GetColumnRef().GetFields()
			} else {
				// Get the expression in the WHERE clause and extract params.
				args := node.GetBoolExpr().GetArgs()
				for idx := range args {
					expr := args[idx].GetAExpr()
					fields = expr.GetLexpr().GetColumnRef().GetFields()
					if len(fields) != 0 {
						s := fields[0].GetString_().GetSval()
						fmt.Printf("Got sval: %s\n", s)
					}
				}
			}
			if len(fields) != 0 {
				s := fields[0].GetString_().GetSval()
				fmt.Printf("Got sval: %s\n", s)
			}
		}
	}
	return stmt, nil
}

func QueryToJson(query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("Invalid input argument for query")
	}

	query, err := pg_query.Normalize(query)
	if err != nil {
		return "", fmt.Errorf("Error normalize query: %s, %w", query, err)
	}
	jsonOut, err := pg_query.ParseToJSON(query)
	if err != nil {
		return "", fmt.Errorf("Error parsing the query: %s, %w", query, err)
	}
	return jsonOut, nil
}

// Return argument count of prepared (SELECT, UPDATE, DELETE) query.
func QueryArgsCount(query string) (int, error) {
	data, err := QueryToJson(query)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Query: %s\n", data)

	_, err = QueryToStatement(query)
	/*
		if err != nil {
			fmt.Printf("Got err %s\n", err.Error())
		}

		var tree interface{}
		err = json.Unmarshal([]byte(data), &tree)
		if err != nil {
			return 0, fmt.Errorf("Error unmarshal json, error: %s", err.Error())
		}

		// JSON object is parsed into a map with string keys
		statements := AnyType{V: tree}.Get("stmts").Array()
		if statements == nil {
			return 0, fmt.Errorf("No statements present in the query string")
		}

		for _, st := range statements {
			stmt := AnyType{V: st}.Get("stmt").Object()
			if stmt == nil {
				return 0, fmt.Errorf("Missing 'stmt' element in json")
			}
			for k, _ := range stmt {
				whereClause := AnyType{V: stmt[k]}.Get("whereClause").Object()
				if whereClause == nil {
					return 0, fmt.Errorf("Missing 'whereClause'")
				}
				for k, _ := range whereClause {
					if k == "A_Expr" {
						return 1, nil
					}
					elem := AnyType{V: whereClause[k]}.Object()
					for k, _ := range elem {
						if k == "args" {
							return len(AnyType{V: elem[k]}.Array()), nil
						}
					}
				}
			}
		}*/
	return 0, nil
}
