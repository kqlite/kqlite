package parser_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kqlite/kqlite/pkg/parser"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

type OnVisitFn func(*pg_query.Node) (v parser.Visitor, err error)
type OnVisitEndFn func(*pg_query.Node) error

type TestWalker struct {
	VisitFn    OnVisitFn
	VisitEndFn OnVisitEndFn
}

func (tw *TestWalker) Visit(node *pg_query.Node) (v parser.Visitor, err error) {
	if tw.VisitFn != nil {
		return tw.VisitFn(node)
	}
	return tw, nil
}

func (tw *TestWalker) VisitEnd(node *pg_query.Node) error {
	if tw.VisitEndFn != nil {
		return tw.VisitEndFn(node)
	}
	return nil
}

func TestParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parser Suite")
}
