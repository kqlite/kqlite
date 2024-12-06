package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/kqlite/kqlite/pkg/server"
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

	log.SetFlags(0)

	s := server.NewServer()
	s.Addr = *addr
	s.DataDir = *dataDir
	if err := s.Start(); err != nil {
		return err
	}
	defer s.Stop()

	log.Printf("listening on %s", s.Addr)

	// Wait on signal before shutting down.
	<-ctx.Done()
	log.Printf("SIGINT received, shutting down")

	// Perform clean shutdown.
	if err := s.Stop(); err != nil {
		return err
	}
	log.Printf("kqlite shutdown complete")

	return nil
}
