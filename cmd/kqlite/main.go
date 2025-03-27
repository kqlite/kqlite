package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/kqlite/kqlite/pkg/pgwire"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	addr := flag.String("addr", ":5432", "postgres protocol bind address")
	dataDir := flag.String("data-dir", "", "data directory")
	flag.Parse()

	if *dataDir == "" {
		return fmt.Errorf("required: -data-dir PATH")
	}

	if err := os.Setenv("DATA_DIR", *dataDir); err != nil {
		return err
	}

	log.SetFlags(0)

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
