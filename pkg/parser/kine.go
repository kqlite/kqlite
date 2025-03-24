package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	paramCharacter = "$"
	numbered       = true

	createTablePrefix  = "CREATE TABLE IF NOT EXISTS kine"
	createTableReplace = `
			CREATE TABLE IF NOT EXISTS kine
				(
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					name TEXT,
					created INTEGER,
					deleted INTEGER,
					create_revision INTEGER,
					prev_revision INTEGER,
					lease INTEGER,
					value BLOB,
					old_value BLOB
				)`

	listQueryIndexPrefix        = "CREATE INDEX IF NOT EXISTS kine_list_query_index"
	listQueryIndexPrefixReplace = "PRAGMA wal_checkpoint(TRUNCATE)"

	listSQLMatch  = `(SELECT MAX(crkv.prev_revision) AS prev_revision FROM kine AS crkv WHERE crkv.name = 'compact_rev_key'),`
	countSQLMatch = `COUNT(c.theid)`

	GetCurrentSQL        = "AND kv.name > $2"
	ListRevisionStartSQL = "AND kv.id <= $2"
	GetRevisionAfterSQL  = "AND kv.name > $2 AND kv.id <= $3"
	CountCurrentSQL      = "AND kv.name > $2"
	CountRevisionSQL     = "AND kv.name > $2 AND kv.id <= $3"

	GetSizeSQL        = "SELECT pg_total_relation_size('kine')"
	GetSizeSQLReplace = "SELECT SUM(pgsize) FROM dbstat"

	CompactSQL        = "DELETE FROM kine AS kv"
	CompactSQLReplace = `
		DELETE FROM kine AS kv
		WHERE
			kv.id IN (
				SELECT kp.prev_revision AS id
				FROM kine AS kp
				WHERE
					kp.name != 'compact_rev_key' AND
					kp.prev_revision != 0 AND
					kp.id <= $1
				UNION
				SELECT kd.id AS id
				FROM kine AS kd
				WHERE
					kd.deleted != 0 AND
					kd.id <= $2
			)`
)

func q(sql, param string, numbered bool) string {
	if param == "?" && !numbered {
		return sql
	}

	regex := regexp.MustCompile(`\?`)
	n := 0
	return regex.ReplaceAllStringFunc(sql, func(string) string {
		if numbered {
			n++
			return param + strconv.Itoa(n)
		}
		return param
	})
}

var (
	columns = "kv.id AS theid, kv.name AS thename, kv.created, kv.deleted, kv.create_revision, kv.prev_revision, kv.lease, kv.value, kv.old_value"
	revSQL  = `
		SELECT MAX(rkv.id) AS id
		FROM kine AS rkv`

	compactRevSQL = `
		SELECT MAX(crkv.prev_revision) AS prev_revision
		FROM kine AS crkv
		WHERE crkv.name = 'compact_rev_key'`

	listSQL = fmt.Sprintf(`
		SELECT *
		FROM (
			SELECT (%s), (%s), %s
			FROM kine AS kv
			JOIN (
				SELECT MAX(mkv.id) AS id
				FROM kine AS mkv
				WHERE
					mkv.name LIKE ?
					%%s
				GROUP BY mkv.name) AS maxkv
				ON maxkv.id = kv.id
			WHERE
				kv.deleted = 0 OR
				?
		) AS lkv
		ORDER BY lkv.thename ASC
		`, revSQL, compactRevSQL, columns)

	GetCurrentSQLReplace        = q(fmt.Sprintf(listSQL, "AND mkv.name > ?"), paramCharacter, numbered)
	ListRevisionStartSQLReplace = q(fmt.Sprintf(listSQL, "AND mkv.id <= ?"), paramCharacter, numbered)
	GetRevisionAfterSQLReplace  = q(fmt.Sprintf(listSQL, "AND mkv.name > ? AND mkv.id <= ?"), paramCharacter, numbered)
	CountCurrentSQLReplace      = q(fmt.Sprintf(`
			SELECT (%s), COUNT(c.theid)
			FROM (
				%s
			) c`, revSQL, fmt.Sprintf(listSQL, "AND mkv.name > ?")), paramCharacter, numbered)

	CountRevisionSQLReplace = q(fmt.Sprintf(`
			SELECT (%s), COUNT(c.theid)
			FROM (
				%s
			) c`, revSQL, fmt.Sprintf(listSQL, "AND mkv.name > ? AND mkv.id <= ?")), paramCharacter, numbered)
)

func rewriteKine(q string) string {
	if strings.HasPrefix(q, createTablePrefix) {
		return createTableReplace
	}

	if strings.HasPrefix(q, listQueryIndexPrefix) {
		return listQueryIndexPrefixReplace
	}

	// GetCurrentSQL match
	if strings.Contains(q, listSQLMatch) && strings.Contains(q, GetCurrentSQL) && !strings.Contains(q, GetRevisionAfterSQL) {
		return GetCurrentSQLReplace
	}

	// ListRevisionStartSQL match
	if strings.Contains(q, listSQLMatch) && strings.Contains(q, ListRevisionStartSQL) {
		return ListRevisionStartSQLReplace
	}

	// GetRevisionAfterSQL match
	if strings.Contains(q, listSQLMatch) && strings.Contains(q, GetRevisionAfterSQL) {
		return GetRevisionAfterSQLReplace
	}

	// CountCurrentSQL match
	if strings.Contains(q, countSQLMatch) && strings.Contains(q, CountCurrentSQL) && !strings.Contains(q, CountRevisionSQL) {
		return CountCurrentSQLReplace
	}

	// CountRevisionSQL match
	if strings.Contains(q, countSQLMatch) && strings.Contains(q, CountRevisionSQL) {
		return CountRevisionSQLReplace
	}

	// CompactSQL match
	if strings.Contains(q, CompactSQL) {
		return CompactSQLReplace
	}

	return q
}
