package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // sql driver
)

func main() {
	dataSourceName := "postgres://127.0.0.1:5432/sakila_master.db?sslmode=disable"

	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	status := "up"
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println(status)

	//rows, err := db.QueryContext(ctx, "SELECT brand, model FROM cars WHERE brand = ?", "skoda")
	//rows, err := db.QueryContext(ctx, "update cars set model = 'rapid' where brand = ?", "skoda")
	//langid := 4
	//rows, err := db.QueryContext(ctx, "SELECT * FROM language WHERE language_id=?", "4")
	stmt, _ := db.PrepareContext(ctx, "SELECT * FROM language WHERE name=? AND last_update=?")
	rows, err := stmt.QueryContext(ctx, "Mandarin", "2020-12-23 07:12:12")
	//rows, err := db.QueryContext(ctx, "DELETE FROM customer WHERE customer_id=?", "17")
	//rows, err := db.QueryContext(ctx, "SELECT * FROM cars")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	//var brand, model string
	var lang_id int
	var lang, date string
	for rows.Next() {
		//if err := rows.Scan(&brand, &model); err != nil {
		if err := rows.Scan(&lang_id, &lang, &date); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Fatal(err)
		}
	}
	//log.Printf("brand: %s, model: %s", brand, model)
	log.Printf("lang_id: %d, lang: %s, date: %s", lang_id, lang, date)
	stmt.Close()
	db.Close()
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
