package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // sql driver
)

func main() {
	dataSourceName := "postgres://127.0.0.1:5432/kine.db.back?sslmode=disable"

	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	//status := "up"
	//if err := db.PingContext(ctx); err != nil {
	//	log.Fatal(err)
	//}
	//log.Println(status)

	rows, err := db.QueryContext(ctx, "SELECT brand, model, year FROM cars WHERE brand = ? AND year = ?", "skoda", "2021")

	//rows, err := db.QueryContext(ctx, "SELECT * FROM cars")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var brand, year, model string
	for rows.Next() {
		if err := rows.Scan(&brand, &model, &year); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			log.Fatal(err)
		}
	}
	log.Printf("brand: %s, model: %s, year: %s", brand, model, year)
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
