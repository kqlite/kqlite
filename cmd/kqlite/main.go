package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/kqlite/kqlite/pkg/pgwire"
	"github.com/kqlite/kqlite/pkg/store"
)

const logo = `
 _         _ _ _       
| |       | (_) |      
| | ______| |_| |_ ___ 
| |/ / _  | | | __/ _ \  Lightweight remote high available SQLite.
|   < (_| | | | ||  __/  kqlite.io
|_|\_\__, |_|_|\__\___|
        | |            
        |_|            
`

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	addr := flag.String("addr", ":5432", "server bind address")
	dataDir := flag.String("data-dir", "", "data directory")
	joinAddr := flag.String("join-addr", "", "join/connect to a remote node, for example :5432")
	joinDb := flag.String("join-db", "*", "join/connect to a database for replication, can be a ',' separated list")
	flag.Parse()

	if *dataDir == "" {
		return fmt.Errorf("required: -data-dir PATH")
	}
	if err := os.Setenv("DATA_DIR", *dataDir); err != nil {
		return err
	}
	fmt.Print(logo)
	log.SetFlags(0)

	if *joinAddr != "" && *joinDb == "" {
		return fmt.Errorf("required: -join-db NAME when -join-addr specified")
	}

	if err := store.Bootstrap(*addr, *joinAddr, *joinDb); err != nil {
		return err
	}
	server := pgwire.NewServer(*addr, *dataDir)
	if err := server.Start(); err != nil {
		return err
	}
	log.Printf("listening on %s", server.Address)

	// Wait on signal before shutting down.
	<-ctx.Done()
	log.Printf("SIGINT received, shutting down")

	// Perform clean shutdown.
	if err := server.Stop(); err != nil {
		return err
	}
	log.Printf("kqlite shutdown complete")

	return nil
}
