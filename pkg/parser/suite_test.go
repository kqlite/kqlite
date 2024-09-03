package parser_test

import (
	"testing"

	"github.com/kqlite/kqlite/pkg/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

type TestWalker struct{}

//func (tw *TestWalker) Do(stmt sql.Statement) (sql.Statement, bool, bool, error) {
//	err := sql.Walk(rw, stmt)
//	if err != nil {
//		return nil, false, false, err
//	}
//	return stmt, rw.randRewritten, rw.returning, nil
//}

func (tw *TestWalker) Visit(node *pg_query.Node) (v parser.Visitor, err error) {
	switch n := node.Node.(type) {
	case *pg_query.Node_SelectStmt:
		if n != nil {
			// TODO
		}
		// TODO
	case *pg_query.Node_InsertStmt:
		// Don't rewrite any further down this branch, as ordering by RANDOM
		// should be left to SQLite itself.
		return nil, nil
	case *pg_query.Node_UpdateStmt:
		return tw, nil
	}
	return tw, nil
}

func (tw *TestWalker) VisitEnd(*pg_query.Node) error {
	return nil
}

func TestParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parser Suite")
}
