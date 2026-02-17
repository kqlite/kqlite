package catalog

import (
	"fmt"
	"os"
	"strings"

	sqlite3 "github.com/mattn/go-sqlite3"
)

type PGDatabaseModule struct{}

func (m *PGDatabaseModule) Create(c *sqlite3.SQLiteConn, args []string) (sqlite3.VTab, error) {
	err := c.DeclareVTab(fmt.Sprintf(`
		CREATE TABLE %s (
			oid           INTEGER PRIMARY KEY,
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
	datadir := os.Getenv("DATA_DIR")

	files, err := os.ReadDir(datadir)
	if err != nil {
		return &PGDatabaseCursor{}, err
	}

	dbs := []PGDatabase{}
	for _, file := range files {
		if dbname, found := strings.CutSuffix(file.Name(), ".db"); found {
			dbs = append(dbs, PGDatabase{
				datname:      dbname,
				encoding:     6,
				datcollate:   "en_US.UTF-8",
				datctype:     "en_US.UTF-8",
				datconnlimit: -1,
				datminmxid:   1,
			})
		}
	}

	return &PGDatabaseCursor{0, dbs}, nil
}

func (t *PGDatabaseTable) BestIndex(cst []sqlite3.InfoConstraint, ob []sqlite3.InfoOrderBy) (*sqlite3.IndexResult, error) {
	return &sqlite3.IndexResult{Used: make([]bool, len(cst))}, nil
}

func (t *PGDatabaseTable) Disconnect() error { return nil }
func (t *PGDatabaseTable) Destroy() error    { return nil }

type PGDatabaseCursor struct {
	index int
	dbs   []PGDatabase
}

func (cursor *PGDatabaseCursor) Column(sctx *sqlite3.SQLiteContext, col int) error {
	switch col {
	case 0:
		sctx.ResultInt(cursor.dbs[cursor.index].oid)
	case 1:
		sctx.ResultText(cursor.dbs[cursor.index].datname)
	case 2:
		sctx.ResultInt(cursor.dbs[cursor.index].datdba)
	case 3:
		sctx.ResultInt(cursor.dbs[cursor.index].encoding)
	case 4:
		sctx.ResultText(cursor.dbs[cursor.index].datcollate)
	case 5:
		sctx.ResultText(cursor.dbs[cursor.index].datctype)
	case 6:
		sctx.ResultInt(cursor.dbs[cursor.index].datistemplate)
	case 7:
		sctx.ResultInt(cursor.dbs[cursor.index].datallowconn)
	case 8:
		sctx.ResultInt(cursor.dbs[cursor.index].datconnlimit)
	case 9:
		sctx.ResultInt(cursor.dbs[cursor.index].datlastsysoid)
	case 10:
		sctx.ResultInt(cursor.dbs[cursor.index].datfrozenxid)
	case 11:
		sctx.ResultInt(cursor.dbs[cursor.index].datminmxid)
	case 12:
		sctx.ResultInt(cursor.dbs[cursor.index].dattablespace)
	case 13:
		sctx.ResultText(cursor.dbs[cursor.index].datacl)
	}
	return nil
}

func (cursor *PGDatabaseCursor) Filter(idxNum int, idxStr string, vals []interface{}) error {
	cursor.index = 0
	return nil
}

func (cursor *PGDatabaseCursor) Next() error {
	cursor.index++
	return nil
}

func (cursor *PGDatabaseCursor) EOF() bool {
	return cursor.index >= len(cursor.dbs)
}

func (cursor *PGDatabaseCursor) Rowid() (int64, error) {
	return int64(cursor.index), nil
}

func (cursor *PGDatabaseCursor) Close() error {
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
