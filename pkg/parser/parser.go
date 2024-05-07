package parser

import (
	"encoding/json"
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

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

	var tree interface{}
	err := json.Unmarshal([]byte(data), &tree)
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
	}
	return 0, nil
}
