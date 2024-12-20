package core

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"golang.org/x/sync/errgroup"

	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/parser"
)

// Postgres settings.
const (
	ServerVersion = "13.0.0"
)

// Represents the database server to serve client connections.
type DBServer struct {
	// Network listener
	listener net.Listener

	// Database storage access
	store sync.Map

	// Client network connections
	connections sync.Map

	// Global goroutine group
	group errgroup.Group

	// Global server context
	ctx    context.Context
	cancel func()

	// Bind address to listen to Postgres wire protocol.
	Address string

	// Directory that holds SQLite databases.
	DataDir string
}

type ClientConn struct {
	net.Conn
	backend *pgproto3.Backend
	db      *db.DB
}

func NewClientConn(conn net.Conn) *ClientConn {
	return &ClientConn{
		Conn:    conn,
		backend: pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn),
	}
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
	if _, err := os.Stat(server.DataDir); err != nil {
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

		server.group.Go(func() error {
			defer conn.Close()

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

	for {
		msg, err := conn.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message: %w", err)
		}

		log.Printf("[recv] %#v", msg)

		switch msg := msg.(type) {
		case *pgproto3.Query:
			if err := server.handleQueryMessage(ctx, conn, msg); err != nil {
				return fmt.Errorf("query message: %w", err)
			}
			continue

		case *pgproto3.Parse:
			if err := server.handleParseMessage(ctx, conn, msg); err != nil {
				return fmt.Errorf("parse message: %w", err)
			}
			continue

		case *pgproto3.Sync: // ignore
			continue

		case *pgproto3.Terminate:
			return conn.Close()
			// return nil // exit

		default:
			return fmt.Errorf("unexpected message type: %#v", msg)
		}
	}
}

func (server *DBServer) handleConnStartup(ctx context.Context, conn *ClientConn) error {
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
	log.Printf("received startup message: %#v", msg)

	// Validate
	name := getParameter(msg.Parameters, "database")
	if name == "" {
		return writeMessages(conn, &pgproto3.ErrorResponse{Message: "database required"})
	} else if strings.Contains(name, "..") {
		return writeMessages(conn, &pgproto3.ErrorResponse{Message: "invalid database name"})
	}

	// TODO Check if database exists and validate DB !!!
	// TODO implement authentication.

	// Open SQL database.
	conn.db, err = db.Open(filepath.Join(server.DataDir, name), false, false)
	if err != nil {
		return err
	}
	server.store.Store(name, conn.db)

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

func (server *DBServer) handleQueryMessage(ctx context.Context, conn *ClientConn, msg *pgproto3.Query) error {
	log.Printf("received query: %q", msg.String)

	// Respond to ping queries.
	if strings.HasPrefix(msg.String, "--") && strings.HasSuffix(msg.String, "ping") {
		return writeMessages(conn,
			&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
			&pgproto3.ReadyForQuery{TxStatus: 'I'})
	}

	// Execute query against database.
	rows, err := conn.db.QueryContext(ctx, msg.String)
	if err != nil {
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'},
		)
	}
	defer rows.Close()

	// Encode column header.
	cols, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("column types: %w", err)
	}
	buf, _ := toRowDescription(cols).Encode(nil)

	// Iterate over each row and encode it to the wire protocol.
	for rows.Next() {
		row, err := scanRow(rows, cols)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		buf, _ = row.Encode(buf)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows: %w", err)
	}

	// Mark command complete and ready for next query.
	buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
	buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)

	_, err = conn.Write(buf)
	return err
}

func toRowDescription(cols []*sql.ColumnType) *pgproto3.RowDescription {
	var desc pgproto3.RowDescription
	for _, col := range cols {
		var typeOID uint32
		dbType := col.DatabaseTypeName()
		if pgColType, exists := db.Typemap()[dbType]; exists {
			typeOID = pgColType
		} else {
			typeOID = pgtype.TextOID
		}

		typeSize, ok := col.Length()
		if !ok {
			typeSize = -1
		}

		desc.Fields = append(desc.Fields, pgproto3.FieldDescription{
			Name:                 []byte(col.Name()),
			TableOID:             0,
			TableAttributeNumber: 0,
			DataTypeOID:          typeOID,
			DataTypeSize:         int16(typeSize),
			TypeModifier:         -1,
			Format:               0,
		})
	}
	return &desc
}

func scanRow(rows *sql.Rows, cols []*sql.ColumnType) (*pgproto3.DataRow, error) {
	refs := make([]interface{}, len(cols))
	values := make([]interface{}, len(cols))
	for i := range refs {
		refs[i] = &values[i]
	}

	// Scan from SQLite database.
	if err := rows.Scan(refs...); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	// Convert to TEXT values to return over Postgres wire protocol.
	row := pgproto3.DataRow{Values: make([][]byte, len(values))}
	for i := range values {
		row.Values[i] = []byte(fmt.Sprint(values[i]))
	}
	return &row, nil
}

func (server *DBServer) handleParseMessage(ctx context.Context, conn *ClientConn, pmsg *pgproto3.Parse) error {
	// Rewrite system-information queries so they're tolerable by SQLite.
	query := parser.RewriteQuery(pmsg.Query)

	if pmsg.Query != query {
		log.Printf("query rewrite: %s", query)
	}

	result, err := parser.Parse(query)
	if err != nil {
		return err
	}
	// Extract query params if any
	var paramTypes []uint32
	for idx := range result {
		colTypes, err := db.LookupTypeInfo(ctx, conn.db, result[idx].Args, result[idx].Tables)
		if err != nil {
			return err
		}
		paramTypes = append(paramTypes, colTypes...)
	}

	// Prepare the query.
	//stmt, err := conn.db.PrepareContext(ctx, pmsg.Query)
	//if err != nil {
	//	return fmt.Errorf("prepare: %w", err)
	//}

	var rows *sql.Rows
	var cols []*sql.ColumnType
	var binds []interface{}
	exec := func() (err error) {
		if rows != nil {
			return nil
		}
		//if rows, err = stmt.QueryContext(ctx, binds...); err != nil {
		//	return fmt.Errorf("query: %w", err)
		//}
		if rows, err = conn.db.QueryContext(ctx, pmsg.Query, binds...); err != nil {
			return fmt.Errorf("query: %w", err)
		}

		if cols, err = rows.ColumnTypes(); err != nil {
			return fmt.Errorf("column types: %w", err)
		}
		return nil
	}

	// LOOP:
	var msgState pgproto3.Describe
	for {
		msg, err := conn.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message during parse: %w", err)
		}

		log.Printf("[recv(p)] %#v", msg)

		switch msg := msg.(type) {
		case *pgproto3.Bind:
			binds = make([]interface{}, len(msg.Parameters))
			for i := range msg.Parameters {
				binds[i] = string(msg.Parameters[i])
			}
		case *pgproto3.Describe:
			msgState = *msg
			break

		case *pgproto3.Execute:
			// Bind received, create Row description.
			if msgState.ObjectType == 0x50 && len(binds) != 0 {
				if err := exec(); err != nil {
					return fmt.Errorf("exec: %w", err)
				}
				buf, _ := toRowDescription(cols).Encode(nil)
				if _, err := conn.Write(buf); err != nil {
					return err
				}
			}

			// TODO: Send pgproto3.ParseComplete?
			var buf []byte
			for rows.Next() {
				row, err := scanRow(rows, cols)
				if err != nil {
					return fmt.Errorf("scan: %w", err)
				}
				buf, _ = row.Encode(buf)
			}
			if err := rows.Err(); err != nil {
				return fmt.Errorf("rows: %w", err)
			}

			// Mark command complete and ready for next query.
			buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
			buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
			_, err := conn.Write(buf)
			msgState = pgproto3.Describe{}

			if rows != nil {
				rows.Close()
			}
			return err

		case *pgproto3.Sync:
			if (msgState != pgproto3.Describe{}) && (msgState.ObjectType == 0x53) {
				writeMessages(conn,
					&pgproto3.ParseComplete{},
					&pgproto3.ParameterDescription{ParameterOIDs: paramTypes},
					//desc,
					&pgproto3.ReadyForQuery{TxStatus: 'I'})
			}
			break
		default:
			return fmt.Errorf("unexpected message type during parse: %#v", msg)
		}
	}
}

func (s *DBServer) execSetQuery(ctx context.Context, conn *ClientConn, query string) error {
	buf, _ := (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(nil)
	buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
	_, err := conn.Write(buf)
	return err
}

func getParameter(m map[string]string, k string) string {
	if m == nil {
		return ""
	}
	return m[k]
}

// writeMessages writes all messages to a single buffer before sending.
func writeMessages(w io.Writer, msgs ...pgproto3.Message) error {
	var buf []byte
	for _, msg := range msgs {
		buf, _ = msg.Encode(buf)
	}
	_, err := w.Write(buf)
	return err
}
