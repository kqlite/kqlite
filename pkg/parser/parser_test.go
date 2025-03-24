package parser_test

import (
	"fmt"
	"github.com/kqlite/kqlite/pkg/parser"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser tests", Ordered, func() {

	It("Parse DELETE with SELECT Statement", func() {
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

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).NotTo(BeEmpty())
		Expect(result[0].ArgColumns[0]).To(Equal("id"))
		Expect(result[0].ArgColumns[1]).To(Equal("id"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(1))
		Expect(result[0].Tables[0]).To(Equal("kine"))
	})

	It("Parse SELECT Statement with JOIN", func() {
		sql := `SELECT orders.order_id, suppliers.name   
				FROM suppliers  
				INNER JOIN orders  
				ON suppliers.supplier_id = orders.supplier_id  
				ORDER BY order_id`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).To(BeEmpty())

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(2))
		Expect(result[0].Tables[0]).To(Equal("suppliers"))
		Expect(result[0].Tables[1]).To(Equal("orders"))
	})

	It("Parse SELECT Statement", func() {
		sql := `SELECT first_name, age FROM employees
			    WHERE department_id IN (SELECT department_id FROM departments WHERE (age + location_id + income) > $1 AND age > $2)`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).NotTo(BeEmpty())
		Expect(result[0].ArgColumns[0]).To(Equal("income"))
		Expect(result[0].ArgColumns[1]).To(Equal("age"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables[0]).To(Equal("employees"))
	})

	It("Parse complex SELECT with Boolean expression.", func() {
		sql := `
		SELECT
		(SELECT MAX(rkv.id) AS id FROM kine AS rkv),
		(SELECT MAX(crkv.prev_revision) AS prev_revision FROM kine AS crkv WHERE crkv.name = 'compact_rev_key'),
		maxkv.*
		FROM (SELECT DISTINCT ON (name)
			kv.id AS theid, kv.name, kv.created, kv.deleted, kv.create_revision, kv.prev_revision, kv.lease, kv.value, kv.old_value
			FROM
			kine AS kv
				WHERE
					kv.name LIKE $1 
					AND kv.name > $2
						ORDER BY kv.name, theid DESC
							) AS maxkv
								WHERE
									maxkv.deleted = 0 OR $3
										ORDER BY maxkv.name, maxkv.theid DESC`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).NotTo(BeEmpty())
		Expect(result[0].ArgColumns).To(HaveLen(3))
		//Expect(result[0].ArgColumns[0]).To(Equal("income"))
		//Expect(result[0].ArgColumns[1]).To(Equal("age"))
		//
		//By("Verify query tables")
		//Expect(result[0].Tables).NotTo(BeEmpty())
		//Expect(result[0].Tables[0]).To(Equal("employees"))
	})

	It("Parse complex SELECT with anonymous parameter present", func() {
		sql := `
		SELECT *
		FROM (
			SELECT (
		SELECT MAX(rkv.id) AS id
		FROM kine AS rkv), (
		SELECT MAX(crkv.prev_revision) AS prev_revision
		FROM kine AS crkv
		WHERE crkv.name = 'compact_rev_key'), kv.id AS theid, kv.name AS thename, kv.created, kv.deleted, kv.create_revision, kv.prev_revision, kv.lease, kv.value, kv.old_value
			FROM kine AS kv
			JOIN (
				SELECT MAX(mkv.id) AS id
				FROM kine AS mkv
				WHERE
					mkv.name LIKE $1
					AND mkv.name > $2
				GROUP BY mkv.name) AS maxkv
				ON maxkv.id = kv.id
			WHERE
				kv.deleted = 0 OR
				$3
		) AS lkv
		ORDER BY lkv.thename ASC`

		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).NotTo(BeEmpty())
		Expect(result[0].ArgColumns).To(HaveLen(3))
		fmt.Printf("result[0].ArgColumns %v\n", result[0].ArgColumns)
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

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).NotTo(BeEmpty())
		Expect(result[0].ArgColumns).To(HaveLen(2))
		Expect(result[0].ArgColumns[0]).To(Equal("type"))
		Expect(result[0].ArgColumns[1]).To(Equal("tablename"))

		// Parser considers virtual tables like 'tables' as real table reference.
		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(2))
		Expect(result[0].Tables[0]).To(Equal("tables"))
		Expect(result[0].Tables[1]).To(Equal("sqlite_master"))
	})

	It("Parse INSERT Statement", func() {
		sql := `INSERT INTO kine(name, created, deleted, create_revision, prev_revision, lease, value, old_value)
				values($1, $2, $3, $4, $5, $6, $7, $8)`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).NotTo(BeEmpty())
		Expect(result[0].ArgColumns).To(HaveLen(8))
		Expect(result[0].ArgColumns[0]).To(Equal("name"))
		Expect(result[0].ArgColumns[1]).To(Equal("created"))
		Expect(result[0].ArgColumns[2]).To(Equal("deleted"))
		Expect(result[0].ArgColumns[3]).To(Equal("create_revision"))
		Expect(result[0].ArgColumns[4]).To(Equal("prev_revision"))
		Expect(result[0].ArgColumns[5]).To(Equal("lease"))
		Expect(result[0].ArgColumns[6]).To(Equal("value"))
		Expect(result[0].ArgColumns[7]).To(Equal("old_value"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(1))
		Expect(result[0].Tables[0]).To(Equal("kine"))

	})

	It("Parse INSERT with SELECT Statement", func() {
		sql := `INSERT INTO Customers (CustomerName, City, Country)
				SELECT SupplierName, City, Country FROM Suppliers
				WHERE Country=$1`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).NotTo(BeEmpty())
		Expect(result[0].ArgColumns).To(HaveLen(1))
		Expect(result[0].ArgColumns[0]).To(Equal("country"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(2))
		Expect(result[0].Tables[0]).To(Equal("customers"))
		Expect(result[0].Tables[1]).To(Equal("suppliers"))
	})

	It("Parse UPDATE Statement", func() {
		sql := `UPDATE books SET books.primary_author = $1 FROM books INNER JOIN authors
				ON books.author_id = authors.id WHERE books.title = $2`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).To(HaveLen(2))
		Expect(result[0].ArgColumns[0]).To(Equal("primary_author"))
		Expect(result[0].ArgColumns[1]).To(Equal("title"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(2))
		Expect(result[0].Tables[0]).To(Equal("books"))
		Expect(result[0].Tables[1]).To(Equal("authors"))
	})

	It("Parse UPDATE with SELECT Statement", func() {
		sql := `UPDATE Persons
				SET Persons.PersonCityName=(SELECT AddressList.PostCode
                FROM AddressList
                WHERE AddressList.PersonId = $1)`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).NotTo(BeEmpty())
		Expect(result[0].ArgColumns).To(HaveLen(1))
		Expect(result[0].ArgColumns[0]).To(Equal("personid"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(2))
		Expect(result[0].Tables[0]).To(Equal("persons"))
		Expect(result[0].Tables[1]).To(Equal("addresslist"))
	})

	It("Parse simple UPDATE Statement", func() {
		sql := `UPDATE _litestream_seq SET seq = $1 WHERE id = $2`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).To(HaveLen(2))
		Expect(result[0].ArgColumns[0]).To(Equal("seq"))
		Expect(result[0].ArgColumns[1]).To(Equal("id"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(1))
		Expect(result[0].Tables[0]).To(Equal("_litestream_seq"))
	})

	It("Parse UPDATE Statement with multiple columns", func() {
		sql := `UPDATE seq_number SET seq = $1, double = $2, triple = $3 WHERE id = $4`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).To(HaveLen(4))
		Expect(result[0].ArgColumns[0]).To(Equal("seq"))
		Expect(result[0].ArgColumns[1]).To(Equal("double"))
		Expect(result[0].ArgColumns[2]).To(Equal("triple"))
		Expect(result[0].ArgColumns[3]).To(Equal("id"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(1))
		Expect(result[0].Tables[0]).To(Equal("seq_number"))
	})

	It("Parse SELECT with CASE WHEN statement", func() {
		sql := `SELECT
					trackid,
					name,
					CASE
						WHEN milliseconds < $1 THEN
							'short'
						WHEN milliseconds > $2 AND milliseconds < $3 THEN 'medium'
						ELSE
							'long'
						END category
				FROM tracks;`
		result, err := parser.Parse(sql)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(HaveLen(1))

		By("Verify query args/parameters")
		Expect(result[0].ArgColumns).To(HaveLen(3))
		Expect(result[0].ArgColumns[0]).To(Equal("milliseconds"))
		Expect(result[0].ArgColumns[1]).To(Equal("milliseconds"))
		Expect(result[0].ArgColumns[2]).To(Equal("milliseconds"))

		By("Verify query tables")
		Expect(result[0].Tables).NotTo(BeEmpty())
		Expect(result[0].Tables).To(HaveLen(1))
		Expect(result[0].Tables[0]).To(Equal("tracks"))
	})
})
