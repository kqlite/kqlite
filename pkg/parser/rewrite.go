package parser

import (
	"regexp"
	"strconv"
	"strings"
)

func isSystemFunc(sql string) bool {
	var re = regexp.MustCompile(`current_catalog|current_schema|current_user|session_user|user`)
	return re.MatchString(sql)
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

func RewriteQueryBlobSerialization(q string) string {
	q = emptyBlobSerializedRegex.ReplaceAllString(q, "'0'")
	return blobSerializedRegex.ReplaceAllString(q, "unhex('$1')")
}

// Basic query rewrite.
func RewriteQuery(q string) string {
	// Ignore SET queries by rewriting them to empty resultsets.
	if strings.HasPrefix(q, "SET ") {
		return `SELECT 'SET'`
	}

	q = rewriteKine(q)

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

	// Replace serialised blob's from https://github.com/jackc/pgx/blob/master/internal/sanitize/sanitize.go.
	// Also kqlite is using the simple-query protocol for replication by sending the write queries along with data in serialised text format.
	// Binary data (BLOB) needs to be serialised and deserialised upon receiving before data enters DB.
	// The deserialization is done from sqlite by using the 'unhex' builtin https://www.sqlite.org/lang_corefunc.html#unhex.
	q = blobSerializedRegex.ReplaceAllString(q, "unhex('$1')")
	q = emptyBlobSerializedRegex.ReplaceAllString(q, "'0'")

	return replaceArgStubs(q)
}

var (
	systemFunctionRegex = regexp.MustCompile(`\b(current_catalog|current_schema|current_user|session_user|user)\b([^\(]|$)`)

	castRegex = regexp.MustCompile(`::(regclass)`)

	pgCatalogRegex = regexp.MustCompile(`\bpg_catalog\.`)

	showRegex = regexp.MustCompile(`^SHOW (\w+)`)

	emptyBlobSerializedRegex = regexp.MustCompile(`\'\\x\'`)

	blobSerializedRegex = regexp.MustCompile(`\'\\x([^\\x,]+)\'`)
)
