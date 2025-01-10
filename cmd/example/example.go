package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // sql driver
)

func pingDB(ctx context.Context, db *sql.DB) {
	log.Println("Ping DB")

	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	status := "up"
	if err := conn.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println(status)
}

func simpleSelectQuery(ctx context.Context, db *sql.DB) {
	var lang_id int
	var lang, date string

	log.Println("Simple Select query")

	langid := 2
	rows, err := db.QueryContext(ctx, "SELECT * FROM language WHERE language_id=?", langid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&lang_id, &lang, &date); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Fatal(err)
		}
		log.Printf("lang_id: %d, lang: %s, date: %s", lang_id, lang, date)
	}
}

func simpleSelectPreparedQuery(ctx context.Context, db *sql.DB) {
	var lang_id int
	var lang, date string

	log.Println("Simple prepared Select query")

	stmt, err := db.PrepareContext(ctx, "SELECT * FROM language WHERE name=? AND last_update=?")
	//stmt, err := db.PrepareContext(ctx, "SELECT * FROM language WHERE language_id=?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	//lang_id_sel := 4
	rows, err := stmt.QueryContext(ctx, "Mandarin", "2020-12-23 07:12:12")
	//rows, err := stmt.QueryContext(ctx, lang_id_sel)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&lang_id, &lang, &date); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Fatal(err)
		}
		log.Printf("lang_id: %d, lang: %s, date: %s", lang_id, lang, date)
	}
}

func testTransaction(ctx context.Context, db *sql.DB) {
	log.Println("Test transaction")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	seq := 728
	id := 1
	_, execErr := tx.Exec("UPDATE _litestream_seq SET seq = ? WHERE id = ?", seq, id)

	//_, execErr := tx.Exec("UPDATE _litestream_seq SET seq = 777 WHERE id = 1")
	if execErr != nil {
		_ = tx.Rollback()
		log.Fatal(execErr)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func simpleUpdateQuery(ctx context.Context, db *sql.DB) {
	_, err := db.ExecContext(ctx, "UPDATE _litestream_seq SET seq = 727 WHERE id = 1")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	dataSourceName := "postgres://127.0.0.1:5432/sakila_master.db?sslmode=disable"
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pingDB(ctx, db)

	testTransaction(ctx, db)
	simpleSelectPreparedQuery(ctx, db)
	simpleUpdateQuery(ctx, db)
	simpleSelectQuery(ctx, db)
}

/*
import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dataSourceName := "postgres://127.0.0.1:5432/kine.db?sslmode=disable"
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	var greeting string
	err = db.QueryRow("select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
}
*/
