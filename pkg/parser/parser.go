package parser

import (
	//"fmt"
	"regexp"
	"strings"
)

var sqliteOperators = []string{
	"\\||", "->", "->>",
	"\\*", "/", "%", "\\+", "-",
	"\\|", "&", "<<", ">>",
	"<", ">", "<=", ">=", "=", "==", "<>", "!=", "IS", "IS NOT",
	"IN", "MATCH", "LIKE",
	"NOT", "AND", "OR",
}

var (
	exprRegex = regexp.MustCompile(`\((.*?)\)\s*(` + strings.Join(sqliteOperators, "|") + `)\s*\?`)

	paramsRegex = regexp.MustCompile(`\((.*?)\)\s*(` + strings.Join(sqliteOperators, "|") + `)\s*\?`)

	insertIntoValuesRegex = regexp.MustCompile(`^INSERT(\s+)INTO\s+(\S+)\s+\((.*?)\)\s+VALUES`)

	insertIntoRegex = regexp.MustCompile(`^INSERT(\s+)INTO\s+(\S+)\s+VALUES`)
)

func IsSelectStatement(st string) bool {
	return strings.HasPrefix(strings.ToUpper(st), "SELECT")
}

func IsInsertStament(st string) bool {
	return strings.HasPrefix(strings.ToUpper(st), "INSERT")
}

// Get the list of parameter names (table fields) from every prepared statements in a SQL program.
func GetPreparedParams(sql string) ([][]string, error) {
	var err error
	params := [][]string{}

	stms := strings.Split(sql, ";")
	if len(stms) == 0 {
		stms = append(stms, sql)
	}
	for _, st := range stms {
		// parse statement and extract params
		statementParams := []string{}
		if IsInsertStament(st) {
			//statementParams, err = getInsertParams(sql)
		} else {
			statementParams, err = getStatementParams(sql)
		}
		if err != nil {
			return params, err
		}
		params = append(params, statementParams)
	}
	return params, err
}

// Stub func
func QueryArgsCount(query string) (int, error) {
	return 2, nil
}

// Parses an expression part of a SQL statement and returns the list of table fields that are used in the expression.
// Useful when in a prepared statement a parameter is evaluated against an expression.
//func parseExpression(expr string) ([]string, error) {
//}

// Extract a list of parameters (table fields) from a SELECT, UPDATE or DELETE prepared statement.
func getStatementParams(sql string) ([]string, error) {
	var p []string
	if sql == "" {
		return p, nil
	}
	return p, nil
}

// Extract a list of parameters (table fields) from a INSERT prepared statement.
//func getInsertParams(sql string) ([]string, error) {
//}
