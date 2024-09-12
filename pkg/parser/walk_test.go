package parser_test

import (
	"fmt"

	"github.com/kqlite/kqlite/pkg/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

var _ = Describe("Walker tests", Ordered, func() {

	It("Walk Simple Select statement", func() {
		qv := &TestWalker{}
		foundSelect := false
		sql := `SELECT first_name, age FROM employees WHERE income = $1`

		tree, err := pg_query.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(tree).NotTo(BeNil())

		By("Verify Select statement is present", func() {
			qv.VisitFn = func(node *pg_query.Node) (v parser.Visitor, err error) {
				switch n := node.Node.(type) {
				case *pg_query.Node_SelectStmt:
					if n != nil {
						foundSelect = true
					}
				}
				return qv, nil
			}

			for _, raw := range tree.Stmts {
				st := raw.GetStmt()
				Expect(st).NotTo(BeNil())
				Expect(parser.Walk(qv, st)).NotTo(HaveOccurred())
			}
			Expect(foundSelect).To(BeTrue())
		})

		By("Extract parameter from a simple query expression", func() {
			exprLoc := 0
			exprName := ""
			foundColumnref := false
			columns := []string{}
			qv.VisitFn = func(node *pg_query.Node) (v parser.Visitor, err error) {
				switch n := node.Node.(type) {
				case *pg_query.Node_SelectStmt:
					foundSelect = true
				case *pg_query.Node_AExpr:
					exprLoc = int(n.AExpr.GetLocation())
				case *pg_query.Node_ColumnRef:
					if exprLoc != 0 && exprName != "" {
						foundColumnref = true
					}
				case *pg_query.Node_String_:
					if exprLoc != 0 && exprName == "" {
						// extract expression name
						exprName = n.String_.GetSval()
					} else if foundColumnref {
						// extract columns from expression
						columns = append(columns, n.String_.GetSval())
					}
				case *pg_query.Node_ParamRef:
					// TODO
				}
				return qv, nil
			}

			for _, raw := range tree.Stmts {
				st := raw.GetStmt()
				Expect(st).NotTo(BeNil())
				Expect(parser.Walk(qv, st)).NotTo(HaveOccurred())
			}
			Expect(foundSelect).To(BeTrue())
			Expect(columns).NotTo(BeEmpty())
			Expect(columns).To(HaveLen(1))
			Expect(exprName).NotTo(Equal(""))
			Expect(exprLoc).NotTo(BeZero())
			fmt.Printf("%s, %d, %+v\n", exprName, exprLoc, columns)
		})
	})
})
