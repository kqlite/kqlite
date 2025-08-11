package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/kqlite/kqlite/pkg/cluster"
	"github.com/kqlite/kqlite/pkg/connpool"
	"github.com/kqlite/kqlite/pkg/db"
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
func sendJoinRequest(localAddr, joinAddr, joinDb string) error {
	// Get connection params.
	host, port, err := getHostPort(joinAddr)
	if err != nil {
		return err
	}
	// Connect to primary node's system database/catalog and send join request.
	conn, err := pgx.Connect(context.Background(),
		fmt.Sprintf("postgres://%s:%s/%s?prefer_simple_protocol=true?sslmode=disable",
			host,
			port,
			ReplicaDB))
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	var joinId int
	hostname, _ := os.Hostname()
	_, localPort, err := getHostPort(localAddr)
	if err != nil {
		return err
	}
	hostAddr := fmt.Sprintf("%s:%s", hostname, localPort)
	err = conn.QueryRow(context.Background(),
		fmt.Sprintf("SELECT id FROM replicas WHERE addr='%s' and db='%s'", hostAddr, joinDb)).Scan(&joinId)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else if errors.Is(err, pgx.ErrNoRows) {
		// Send join request as specified.
		if _, err = conn.Exec(context.Background(),
			fmt.Sprintf("INSERT INTO replicas (addr, db) VALUES('%s', '%s')", hostAddr, joinDb)); err != nil {
			return err
		}
	}
	// Open a replication link to primary.
	log.Printf("Openning bi-directional replication link")
	if err := connpool.NewConnection(context.Background(), host, port, joinDb); err != nil {
		return err
	}
	return nil
}

// Check for pending join request and open replication link to target node:database.
func checkForJoinRequest(conndb *db.Database) (bool, error) {
	var addr string
	var dbname string
	if err := conndb.QueryRow("SELECT addr, db FROM replicas").Scan(&addr, &dbname); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if addr == "" && dbname == "" {
		return false, nil
	}

	host, port, err := getHostPort(addr)
	if err != nil {
		return false, err
	}
	// TODO logging
	log.Printf("Openning replication link")
	// Open a replication link.
	if err := connpool.NewConnection(context.Background(), host, port, dbname); err != nil {
		return true, err
	}
	return true, nil
}

// Install a watch cb function to be notified when join request comes in.
func watchForJoinRequest() error {
	replicasdb, err := db.Open(filepath.Join(os.Getenv("DATA_DIR"), ReplicaDBFile), false, false, false)
	if err != nil {
		return err
	}
	defer replicasdb.Close()

	watchJoinHook := func(ev *command.UpdateHookEvent) error {
		if ev.Table == "replicas" && ev.Op == command.UpdateHookEvent_INSERT {
			conndb, err := db.Open(filepath.Join(os.Getenv("DATA_DIR"), ReplicaDBFile), false, false, false)
			if err != nil {
				return err
			}
			defer conndb.Close()
			if _, err := checkForJoinRequest(conndb); err != nil {
				return err
			}
		}
		return nil
	}

	if err := replicasdb.RegisterUpdateHook(watchJoinHook); err != nil {
		return err
	}
	return nil
}

// Bootstraps the DataStore with initial configuration.
// DataStore will be either in primary or secondary/replica mode.
func Bootstrap(localAddr, joinAddr, joinDb string) error {
	// Open connection to the system database and create system catalog if not created.
	replicasdb, err := db.Open(filepath.Join(os.Getenv("DATA_DIR"), ReplicaDBFile), false, false, false)
	if err != nil {
		return err
	}
	defer replicasdb.Close()

	// Create initial system database/catalog schema.
	if _, err = replicasdb.Exec(ReplicasSchema); err != nil {
		return err
	}
	// Add join candidate to the system catalog and set cluster role.
	if joinAddr != "" {
		cluster.SetSecondary()
		if err := sendJoinRequest(localAddr, joinAddr, joinDb); err != nil {
			return err
		}
		// Connect to remote after sending join request.
		// Open a replication link.
		remoteHost, remotePort, _ := getHostPort(joinAddr)
		if err := connpool.NewConnection(context.Background(), remoteHost, remotePort, joinDb); err != nil {
			return err
		}
	} else {
		cluster.SetPrimary()
		// Check for joined replica node.
		found, err := checkForJoinRequest(replicasdb)
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
