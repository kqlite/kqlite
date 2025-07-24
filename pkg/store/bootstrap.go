package store

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/kqlite/kqlite/pkg/cluster"
	"github.com/kqlite/kqlite/pkg/connpool"
	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/sysdb"
	"github.com/kqlite/kqlite/pkg/util/command"
)

// Little helper for extracting host and port from address string.
func getHostPort(addr string) (host, port string, err error) {
	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		return host, port, err
	}
	// Set defaults.
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	return host, port, err
}

// Send join request for this node/replica address to request replication for 'joinDb' database.
// Request is send via PostgreSQL wire protocol using an SQL query and writing to primary node system catalog.
func sendJoinRequest(joinAddr, joinDb string) error {
	// Get connection params.
	host, port, err := getHostPort(joinAddr)
	if err != nil {
		return err
	}
	// Connect to primary node's system database/catalog and send join request.
	conn, err := pgx.Connect(context.Background(),
		fmt.Sprintf("postgres://%s:%s/%s?default_query_exec_mode=simple_protocol?sslmode=disable",
			host,
			port,
			sysdb.Catalog))
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	var joinId int
	if err := conn.QueryRow(context.Background(),
		fmt.Sprintf("SELECT id FROM replicas WHERE addr='%s' and db='%s'", joinAddr, joinDb)).Scan(&joinId); err != nil {
		// Already joined.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	// Send join request as specified.
	if _, err = conn.Exec(context.Background(),
		fmt.Sprintf("INSERT INTO replicas VALUES('%s', '%s')", joinAddr, joinDb)); err != nil {
		return err
	}
	return nil
}

// Check for pending join request and open replication link to target node:database.
func checkForJoinRequest(dbcatalog *db.Database) (bool, error) {
	var addr string
	var dbname string
	err := dbcatalog.QueryRow("SELECT addr, db FROM replicas").Scan(&addr, &dbname)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	host, port, err := getHostPort(addr)
	if err != nil {
		return false, err
	}
	// TODO logging
	// log.Printf("Openning replication link")
	// Open a replication link.
	if err := connpool.NewConnection(context.Background(), host, port, dbname); err != nil {
		return true, err
	}
	return true, nil
}

// Install a watch cb function to be notified when join request comes in.
func watchForJoinRequest() error {
	dbcatalog, err := db.Open(filepath.Join(os.Getenv("DATA_DIR"), sysdb.CatalogFile), false, false)
	if err != nil {
		return err
	}
	defer dbcatalog.Close()

	watchJoinHook := func(ev *command.UpdateHookEvent) error {
		if ev.Table == "replicas" && ev.Op == command.UpdateHookEvent_INSERT {
			connCatalog, err := db.Open(filepath.Join(os.Getenv("DATA_DIR"), sysdb.CatalogFile), false, false)
			if err != nil {
				return err
			}
			defer connCatalog.Close()
			if _, err := checkForJoinRequest(dbcatalog); err != nil {
				return err
			}
		}
		return nil
	}

	if err := dbcatalog.RegisterUpdateHook(watchJoinHook); err != nil {
		return err
	}
	return nil
}

// Bootstraps the DataStore with initial configuration.
// DataStore will be either in primary or secondary/replica mode.
func Bootstrap(joinAddr, joinDb string) error {
	// Open connection to the system database and create system catalog if not created.
	dbcatalog, err := db.Open(filepath.Join(os.Getenv("DATA_DIR"), sysdb.CatalogFile), false, false)
	if err != nil {
		return err
	}
	defer dbcatalog.Close()

	// Create initial system database/catalog schema.
	if _, err = dbcatalog.Exec(sysdb.CatalogSchema); err != nil {
		return err
	}
	// Add join candidate to the system catalog and set cluster role.
	if joinAddr != "" {
		cluster.SetSecondary()
		if err := sendJoinRequest(joinAddr, joinDb); err != nil {
			return err
		}
	} else {
		cluster.SetPrimary()
		// Check for joined replica node.
		found, err := checkForJoinRequest(dbcatalog)
		if err != nil {
			return err
		}
		if !found {
			// No joined replica nodes, setup a request watch.
			return watchForJoinRequest()
		}
	}
	return nil
}
