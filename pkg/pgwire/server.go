package pgwire

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgproto3"
	"golang.org/x/sync/errgroup"

	"github.com/kqlite/kqlite/pkg/db"
)

// Postgres settings.
const (
	ServerVersion = "14.0.0"
	systemDB      = "kqlite.db"
)

// Represents the database server to serve client connections.
type DBServer struct {
	// Network listener.
	listener net.Listener

	// Client network connections.
	connections sync.Map

	// Global goroutine group.
	group errgroup.Group

	// System database, storing system related data.
	systemdb *db.DB

	// Global server context
	ctx    context.Context
	cancel func()

	// Bind address to listen to Postgres wire protocol.
	Address string

	// Directory that holds SQLite databases.
	DataDir string

	connCounter int32
}

func NewServer(address, datadir string) *DBServer {
	server := &DBServer{
		Address: address,
		DataDir: datadir,
	}
	server.ctx, server.cancel = context.WithCancel(context.Background())
	return server
}

// Starts the Database server.
func (server *DBServer) Start() (err error) {
	// Ensure data directory exists.
	if _, err = os.Stat(server.DataDir); err != nil {
		return err
	}

	// Open connection to the system database.
	server.systemdb, err = db.Open(filepath.Join(server.DataDir, systemDB), false, false)
	if err != nil {
		return err
	}

	server.listener, err = net.Listen("tcp", server.Address)
	if err != nil {
		return err
	}

	server.group.Go(func() error {
		if err := server.serve(); server.ctx.Err() != nil {
			return err // return error unless context canceled
		}
		return nil
	})
	return nil
}

// Stops the Database server.
func (server *DBServer) Stop() (err error) {
	if server.listener != nil {
		if e := server.listener.Close(); err == nil {
			err = e
		}
	}
	server.cancel()

	// Track and close all open client connections.
	server.connections.Range(func(key, value any) bool {
		conn := key.(*ClientConn)
		if conn != nil {
			if e := conn.Close(); err == nil {
				err = e
			}
		}
		return true
	})
	server.connections.Clear()

	// Clear gloabal database connections pool.
	db.ClearPool()

	// Wait for goroutine's to finish.
	if err := server.group.Wait(); err != nil {
		return err
	}
	return err
}

func (server *DBServer) serve() error {
	for {
		c, err := server.listener.Accept()
		if err != nil {
			return err
		}

		conn := NewClientConn(c)

		// Track client connections.
		server.connections.Store(conn, nil)

		log.Println("connection accepted: ", conn.RemoteAddr())

		atomic.AddInt32(&server.connCounter, 1)
		fmt.Printf("server.connCounter: %d\n", atomic.LoadInt32(&server.connCounter))

		server.group.Go(func() error {
			defer func() {
				if conn.db != nil {
					conn.db.Close()
				}
				conn.Close()
				server.connections.Delete(conn)

				atomic.AddInt32(&server.connCounter, -1)
				fmt.Printf("server.connCounter: %d\n", atomic.LoadInt32(&server.connCounter))
			}()

			if err := server.serveConn(server.ctx, conn); err != nil && server.ctx.Err() == nil {
				log.Printf("connection error, closing: %s", err)
				return nil
			}

			log.Printf("connection closed: %s", conn.RemoteAddr())
			return nil
		})
	}
}

func (server *DBServer) serveConn(ctx context.Context, conn *ClientConn) error {
	if err := server.handleConnStartup(ctx, conn); err != nil {
		return fmt.Errorf("startup: %w", err)
	}

	// Create a query execution context for this DB connection.
	conn.exeqc = conn.db.CreateContext(ctx)

	for {
		msg, err := conn.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message: %w", err)
		}

		//log.Printf("[%s] [recv] %#v", conn.RemoteAddr().String(), msg)

		switch msg := msg.(type) {
		case *pgproto3.Query:
			if err := conn.handleQuery(ctx, msg); err != nil {
				fmt.Printf("error query message: %v", err)
			}
			continue

		case *pgproto3.Parse:
			if err := conn.handleParse(ctx, msg); err != nil {
				fmt.Printf("error parse message: %v", err)
			}
			continue

		case *pgproto3.Describe:
			if err := conn.handleDescribe(ctx, msg); err != nil {
				fmt.Printf("error describe message: %v", err)
			}
			continue

		case *pgproto3.Sync:
			err := writeMessages(conn, &pgproto3.ReadyForQuery{TxStatus: 'I'})
			if err != nil {
				return err
			}
			continue

		case *pgproto3.Bind:
			if err := conn.handleBind(ctx, msg); err != nil {
				fmt.Printf("error bind message: %v", err)
			}
			continue

		case *pgproto3.Execute:
			if err := conn.handleExecute(ctx, msg); err != nil {
				fmt.Printf("error execute message: %v", err)
			}
			continue

		case *pgproto3.Terminate:
			return nil

		case *pgproto3.Close:
			if err := conn.handleClose(ctx, msg); err != nil {
				fmt.Printf("error close message: %v", err)
			}
			continue

		case *pgproto3.CancelRequest:
			fmt.Printf("got cancel request message type: %#v", msg)
			return nil

		default:
			return fmt.Errorf("unexpected message type: %#v", msg)
		}
	}
}

func (server *DBServer) handleConnStartup(ctx context.Context, conn *ClientConn) error {
	defer timer("handleConnStartup")()
	for {
		msg, err := conn.backend.ReceiveStartupMessage()
		if err != nil {
			return fmt.Errorf("receive startup message: %w", err)
		}

		switch msg := msg.(type) {
		case *pgproto3.StartupMessage:
			if err := server.handleStartupMessage(ctx, conn, msg); err != nil {
				return fmt.Errorf("startup message: %w", err)
			}
			return nil
		case *pgproto3.SSLRequest:
			if err := server.handleSSLRequestMessage(ctx, conn, msg); err != nil {
				return fmt.Errorf("ssl request message: %w", err)
			}
			continue
		default:
			return fmt.Errorf("unexpected startup message: %#v", msg)
		}
	}
}

func (server *DBServer) handleStartupMessage(ctx context.Context, conn *ClientConn, msg *pgproto3.StartupMessage) (err error) {
	// log.Printf("received startup message: %#v", msg)

	// Validate
	name := getParameter(msg.Parameters, "database")
	if name == "" {
		return writeMessages(conn, &pgproto3.ErrorResponse{Message: "database required"})
	} else if strings.Contains(name, "..") {
		return writeMessages(conn, &pgproto3.ErrorResponse{Message: "invalid database name"})
	}

	// TODO implement authentication.

	walEnabled := true
	fkEnabled := false
	// Open connection to SQL database.
	dbfilename := name + ".db"
	if conn.db, err = db.Open(filepath.Join(server.DataDir, dbfilename), fkEnabled, walEnabled); err != nil {
		return err
	}

	// Initialize postgres catalog virtual tables.
	if err = initCatatog(conn.db); err != nil {
		return err
	}

	return writeMessages(conn,
		&pgproto3.AuthenticationOk{},
		&pgproto3.ParameterStatus{Name: "server_version", Value: ServerVersion},
		&pgproto3.ReadyForQuery{TxStatus: 'I'},
	)
}

func (server *DBServer) handleSSLRequestMessage(ctx context.Context, conn *ClientConn, msg *pgproto3.SSLRequest) error {
	log.Printf("received ssl request message: %#v", msg)
	// SSL mode currently not supported
	if _, err := conn.Write([]byte("N")); err != nil {
		return err
	}
	return nil
}

func (s *DBServer) execSetQuery(ctx context.Context, conn *ClientConn, query string) error {
	buf, _ := (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(nil)
	buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
	_, err := conn.Write(buf)
	return err
}
