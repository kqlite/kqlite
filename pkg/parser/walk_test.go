package parser_test

import (
	"github.com/kqlite/kqlite/pkg/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

var _ = Describe("Walker tests", Ordered, func() {

	It("Walk SELECT statement", func() {
		walker := &TestWalker{}
		var foundSelect, foundSourceTable bool
		var foundExpr, foundResultTarget int
		sql := `SELECT name, age FROM employees WHERE income = $1`

		tree, err := pg_query.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(tree).NotTo(BeNil())

		walker.VisitFn = func(node *pg_query.Node) (v parser.Visitor, err error) {
			switch n := node.Node.(type) {
			case *pg_query.Node_SelectStmt:
				if n != nil {
					foundSelect = true
				}
			case *pg_query.Node_AExpr:
				foundExpr++
			case *pg_query.Node_ResTarget:
				foundResultTarget++
			case *pg_query.Node_RangeVar:
				foundSourceTable = true
			}
			return walker, nil
		}
		for _, raw := range tree.Stmts {
			st := raw.GetStmt()
			Expect(st).NotTo(BeNil())
			Expect(parser.Walk(walker, st)).NotTo(HaveOccurred())
		}
		Expect(foundSelect).To(BeTrue())
		Expect(foundSourceTable).To(BeTrue())
		Expect(foundExpr).To(Equal(1))
		Expect(foundResultTarget).To(Equal(2))
	})

	It("Walk INSERT statement", func() {
		walker := &TestWalker{}
		var foundInsert bool
		var foundColumnTarget int
		sql := `INSERT INTO kine(name, created, deleted, create_revision, prev_revision, lease, value, old_value)
				values("one", "two", "three", "four", "five", "six", "seven", "eight")`

		tree, err := pg_query.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(tree).NotTo(BeNil())

		walker.VisitFn = func(node *pg_query.Node) (v parser.Visitor, err error) {
			switch n := node.Node.(type) {
			case *pg_query.Node_SelectStmt:
				if n != nil {
					foundInsert = true
				}
			case *pg_query.Node_ResTarget:
				foundColumnTarget++
			}
			return walker, nil
		}

		for _, raw := range tree.Stmts {
			st := raw.GetStmt()
			Expect(st).NotTo(BeNil())
			Expect(parser.Walk(walker, st)).NotTo(HaveOccurred())
		}
		Expect(foundInsert).To(BeTrue())
		Expect(foundColumnTarget).To(Equal(8))

	})

	It("Walk DELETE statement", func() {
		sql := `DELETE FROM kine AS kv
				USING (
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
				) AS ks
				WHERE kv.id = ks.id`
		tree, err := pg_query.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(tree).NotTo(BeNil())

	})

	It("Walk UPDATE statement", func() {
		sql := `UPDATE Persons
				SET Persons.PersonCityName=(SELECT AddressList.PostCode
                FROM AddressList
                WHERE AddressList.PersonId = 10)`
		tree, err := pg_query.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(tree).NotTo(BeNil())
	})
})
