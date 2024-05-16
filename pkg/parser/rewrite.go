package parser

import (
	"regexp"
	"strconv"
	"strings"
)

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
