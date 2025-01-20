package sysdb

import (
	"fmt"

	sqlite3 "github.com/mattn/go-sqlite3"
)

type PGDatabaseModule struct{}

func (m *PGDatabaseModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			oid           INTEGER,
			datname       TEXT,
			datdba        INTEGER,
			encoding      INTEGER,
			datcollate    TEXT,
			datctype      TEXT,
			datistemplate INTEGER,
			datallowconn  INTEGER,
			datconnlimit  INTEGER,
			datlastsysoid INTEGER,
			datfrozenxid  INTEGER,
			datminmxid    INTEGER,
			dattablespace INTEGER,
			datacl        TEXT
		)`, args[0]))
	if err != nil {
		return nil, err
	}
	return &PGDatabaseTable{}, nil
}

func (m *PGDatabaseModule) Connect(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	return m.Create(c, args)
}

func (m *PGDatabaseModule) DestroyModule() {}

type PGDatabaseTable struct{}

func (t *PGDatabaseTable) Open() (sqlite3.VTabCursor, error) {
	return &PGDatabaseCursor{}, nil
}

func (t *PGDatabaseTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *PGDatabaseTable) Disconnect() error { return nil }
func (t *PGDatabaseTable) Destroy() error    { return nil }

type PGDatabaseCursor struct {
	index int
}

func (c *PGDatabaseCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultInt(pgDatabases[c.index].oid)
	case 1:
		sctx.ResultText(pgDatabases[c.index].datname)
	case 2:
		sctx.ResultInt(pgDatabases[c.index].datdba)
	case 3:
		sctx.ResultInt(pgDatabases[c.index].encoding)
	case 4:
		sctx.ResultText(pgDatabases[c.index].datcollate)
	case 5:
		sctx.ResultText(pgDatabases[c.index].datctype)
	case 6:
		sctx.ResultInt(pgDatabases[c.index].datistemplate)
	case 7:
		sctx.ResultInt(pgDatabases[c.index].datallowconn)
	case 8:
		sctx.ResultInt(pgDatabases[c.index].datconnlimit)
	case 9:
		sctx.ResultInt(pgDatabases[c.index].datlastsysoid)
	case 10:
		sctx.ResultInt(pgDatabases[c.index].datfrozenxid)
	case 11:
		sctx.ResultInt(pgDatabases[c.index].datminmxid)
	case 12:
		sctx.ResultInt(pgDatabases[c.index].dattablespace)
	case 13:
		sctx.ResultText(pgDatabases[c.index].datacl)
	}
	return nil
}

func (c *PGDatabaseCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	c.index = 0
	return nil
}

func (c *PGDatabaseCursor) Next() error {
	c.index++
	return nil
}

func (c *PGDatabaseCursor) EOF() bool {
	return c.index >= len(pgDatabases)
}

func (c *PGDatabaseCursor) Rowid() (int64, error) {
	return int64(c.index), nil
}

func (c *PGDatabaseCursor) Close() error {
	return nil
}

type PGDatabase struct {
	oid           int
	datname       string
	datdba        int
	encoding      int
	datcollate    string
	datctype      string
	datistemplate int
	datallowconn  int
	datconnlimit  int
	datlastsysoid int
	datfrozenxid  int
	datminmxid    int
	dattablespace int
	datacl        string
}

var pgDatabases = []PGDatabase{}
