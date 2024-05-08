package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
	}
	return 0, nil
}

// Replace query argument stubs like '?' with $n
func replaceArgStubs(sql string) string {
	regex := regexp.MustCompile(`\?`)
	pref := "$"
	n := 0
	return regex.ReplaceAllStringFunc(sql, func(string) string {
		n++
		return pref + strconv.Itoa(n)
	})
}

// Basic query rewrite.
func RewriteQuery(q string) string {
	// Ignore SET queries by rewriting them to empty resultsets.
	if strings.HasPrefix(q, "SET ") {
		return `SELECT 'SET'`
	}

	// Ignore this god forsaken query for pulling keywords.
	if strings.Contains(q, `select string_agg(word, ',') from pg_catalog.pg_get_keywords()`) {
		return `SELECT '' AS "string_agg" WHERE 1 = 2`
	}

	// Rewrite system information variables so they are functions so we can inject them.
	// https://www.postgresql.org/docs/9.1/functions-info.html
	q = systemFunctionRegex.ReplaceAllString(q, "$1()$2")

	// Rewrite double-colon casting by simply removing it.
	// https://www.postgresql.org/docs/7.3/sql-expressions.html#SQL-SYNTAX-TYPE-CASTS
	q = castRegex.ReplaceAllString(q, "")

	// Remove references to the pg_catalog.
	// q = pgCatalogRegex.ReplaceAllString(q, "")

	// Rewrite "SHOW" commands into function calls.
	q = showRegex.ReplaceAllString(q, "SELECT show('$1')")

	return replaceArgStubs(q)
}

var (
	systemFunctionRegex = regexp.MustCompile(`\b(current_catalog|current_schema|current_user|session_user|user)\b([^\(]|$)`)

	castRegex = regexp.MustCompile(`::(regclass)`)

	pgCatalogRegex = regexp.MustCompile(`\bpg_catalog\.`)

	showRegex = regexp.MustCompile(`^SHOW (\w+)`)
)
