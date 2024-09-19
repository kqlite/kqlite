package parser_test

import (
	"github.com/kqlite/kqlite/pkg/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser tests", Ordered, func() {

	It("Parse DELETE Statement", func() {
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
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Args[0]).To(Equal("id"))
		Expect(result[0].Args[1]).To(Equal("id"))
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables[0]).To(Equal("kine"))
		Expect(result[0].Tables[1]).To(Equal("kine"))
		Expect(result[0].Tables[2]).To(Equal("kine"))
	})

	It("Parse SELECT Statement", func() {
		sql := `SELECT first_name, age FROM employees
			    WHERE department_id IN (SELECT department_id FROM departments WHERE (age + location_id + income) > $1 AND age > $2)`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Args[0]).To(Equal("income"))
		Expect(result[0].Args[1]).To(Equal("age"))
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables[0]).To(Equal("employees"))
	})

	It("Parse multi SELECT Statement and WITH clause", func() {
		sql := `WITH tables AS (SELECT name tableName, sql 
				FROM sqlite_master WHERE type = $1 AND tableName NOT LIKE $2)
				SELECT fields.name, fields.type, tableName
				FROM tables CROSS JOIN pragma_table_info(tables.tableName) fields`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))
		// Parser condideres virtual tables like 'tables' as real table reference.
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(2))
		Expect(result[0].Args).To(HaveLen(2))
		Expect(result[0].Tables[0]).To(Equal("tables"))
		Expect(result[0].Tables[1]).To(Equal("sqlite_master"))
		Expect(result[0].Args[0]).To(Equal("type"))
		Expect(result[0].Args[1]).To(Equal("tablename"))
	})

	It("Parse INSERT Statement", func() {
		sql := `INSERT INTO kine(name, created, deleted, create_revision, prev_revision, lease, value, old_value)
				values($1, $2, $3, $4, $5, $6, $7, $8)`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(1))
		Expect(result[0].Tables[0]).To(Equal("kine"))
		Expect(result[0].Args).To(HaveLen(8))
		Expect(result[0].Args[0]).To(Equal("name"))
		Expect(result[0].Args[1]).To(Equal("created"))
		Expect(result[0].Args[2]).To(Equal("deleted"))
		Expect(result[0].Args[3]).To(Equal("create_revision"))
		Expect(result[0].Args[4]).To(Equal("prev_revision"))
		Expect(result[0].Args[5]).To(Equal("lease"))
		Expect(result[0].Args[6]).To(Equal("value"))
		Expect(result[0].Args[7]).To(Equal("old_value"))
	})

	It("Parse INSERT wih SELECT Statement", func() {
		sql := `INSERT INTO Customers (CustomerName, City, Country)
				SELECT SupplierName, City, Country FROM Suppliers
				WHERE Country=$1`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Args).To(HaveLen(1))
		Expect(result[0].Tables).To(HaveLen(2))
	})

	It("Parse UPDATE Statement", func() {
		sql := `UPDATE books SET books.primary_author = $1 FROM books INNER JOIN authors
				ON books.author_id = authors.id WHERE books.title = $2`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Args).To(HaveLen(1))
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(3))

		Expect(result[0].Args[0]).To(Equal("title"))
		Expect(result[0].Tables[0]).To(Equal("books"))
		Expect(result[0].Tables[1]).To(Equal("books"))
		Expect(result[0].Tables[2]).To(Equal("authors"))
	})
})
