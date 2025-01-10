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

	"github.com/jackc/pgproto3/v2"
	"golang.org/x/sync/errgroup"

	"github.com/kqlite/kqlite/pkg/db"
	"github.com/kqlite/kqlite/pkg/parser"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// Postgres settings.
const (
	ServerVersion = "14.0.0"
)

// Represents the database server to serve client connections.
type DBServer struct {
	listener    net.Listener   // Network listener
	store       sync.Map       // Database storage access
	connections sync.Map       // Client network connections
	group       errgroup.Group // Global goroutine group

	ctx    context.Context // Global server context
	cancel func()

	Address string // Bind address to listen to Postgres wire protocol.
	DataDir string // Directory that holds SQLite databases.
}

type ClientConn struct {
	net.Conn
	backend *pgproto3.Backend
	db      *db.DB
	exeqc   *db.ExecuteQueryContext
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

	// Create a query execution context for this DB connection.
	conn.exeqc = conn.db.CreateContext(ctx)

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

		case *pgproto3.Close:
			return conn.Close()

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

	// Extract all statements present in the SQL query and do a syntax validation.
	parserResult, err := parser.Parse(msg.String)
	if err != nil {
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'},
		)
	}

	// Convert parser result to database statements.
	var statements []db.Statement
	for _, result := range parserResult {
		statements = append(statements, db.Statement{
			Query:   result.Sql,
			CmdType: convertToStmtCmd(result),
		})
	}

	// Execute all statements part of the SQL query.
	response, err := conn.exeqc.Request(statements)
	if err != nil {
		return writeMessages(conn,
			&pgproto3.ErrorResponse{Message: err.Error()},
			&pgproto3.ReadyForQuery{TxStatus: 'I'},
		)
	}

	// Send response to client for every statement present in the SQL query.
	for _, resp := range response {
		// Handle error from a single statement execution.
		if resp.Error != nil {
			err := writeMessages(conn,
				&pgproto3.ErrorResponse{Message: resp.Error.Error()},
				&pgproto3.ReadyForQuery{TxStatus: 'I'})
			if err != nil {
				return err
			}
		}

		var buf []byte
		if resp.Rows != nil {
			defer resp.Rows.Close()
			// Encode result rows to PG wire data rows.
			buf, err := encodeRows(resp.Rows)
			if err != nil {
				return err
			}

			// Mark command complete and ready for next query.
			buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
			buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
			if _, err := conn.Write(buf); err != nil {
				return err
			}
		} else {
			buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte(resp.CommandTag)}).Encode(buf)
			buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
			if _, err := conn.Write(buf); err != nil {
				return err
			}
		}
	}

	return nil
}

func (server *DBServer) handleParseMessage(ctx context.Context, conn *ClientConn, pmsg *pgproto3.Parse) error {
	// Rewrite system-information queries so they're tolerable by SQLite.
	query := parser.RewriteQuery(pmsg.Query)

	if pmsg.Query != query {
		log.Printf("query rewrite: %s", query)
	}

	// Validate syntax and extract statements parameter names.
	parserResult, err := parser.Parse(query)
	if err != nil {
		return err
	}

	log.Printf("parserResult %#v\n", parserResult)

	// Convert parser result to database statements and execute every statement.
	var statements []db.Statement
	for _, result := range parserResult {
		statements = append(statements, db.Statement{
			Query:   result.Sql,
			CmdType: convertToStmtCmd(result),
		})
	}

	// Extract parameters types for every statement in the query.
	var paramTypes []uint32
	for idx := range parserResult {
		colTypes, err := db.LookupTypeInfo(ctx, conn.db, parserResult[idx].ArgColumns, parserResult[idx].Tables)
		if err != nil {
			return err
		}
		paramTypes = append(paramTypes, colTypes...)
	}

	var binds []interface{}
	var msgState pgproto3.Describe
	var response []db.ExecuteQueryResponse

	// message loop.
	for {
		msg, err := conn.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message during parse: %w", err)
		}

		log.Printf("[recv(p)] %#v", msg)

		switch msg := msg.(type) {
		case *pgproto3.Bind:
			var prevParamCount int

			// Extract parameters values for all statements.
			binds = parametersToValues(msg.Parameters, paramTypes)

			// Provide the parameters for every statement present in the SQL.
			for idx := range statements {
				paramCount := len(parserResult[idx].ArgColumns)
				// make sure we don't overflow the binds array.
				if len(binds) <= (prevParamCount + paramCount) {
					statements[idx].Parameters = binds[prevParamCount:(prevParamCount + paramCount)]
					prevParamCount = (prevParamCount + paramCount)
				}
			}
			break

		case *pgproto3.Describe:
			msgState = *msg
			break

		case *pgproto3.Execute:
			// Bind received, execute query and create Row description.
			if msgState.ObjectType == 0x50 && len(binds) != 0 {
				// Execute all statements part of the SQL query.
				response, err = conn.exeqc.Request(statements)

				if err != nil {
					return writeMessages(conn,
						&pgproto3.ErrorResponse{Message: err.Error()},
						&pgproto3.ReadyForQuery{TxStatus: 'I'},
					)
				}
			}

			// Send response to client for every statement present in the SQL query.
			for _, resp := range response {
				// Handle error from a single statement execution.
				if resp.Error != nil {
					err := writeMessages(conn,
						&pgproto3.ErrorResponse{Message: resp.Error.Error()},
						&pgproto3.ReadyForQuery{TxStatus: 'I'})
					if err != nil {
						return err
					}
				}

				var buf []byte
				if resp.Rows != nil {
					defer resp.Rows.Close()

					buf, err := encodeRows(resp.Rows)
					if err != nil {
						return err
					}

					// Mark command complete and ready for next query.
					buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(buf)
					buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
					if _, err := conn.Write(buf); err != nil {
						return err
					}
				} else {
					buf, _ = (&pgproto3.CommandComplete{CommandTag: []byte(resp.CommandTag)}).Encode(buf)
					buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
					if _, err := conn.Write(buf); err != nil {
						return err
					}
				}

				if err := writeMessages(conn, &pgproto3.ParseComplete{}); err != nil {
					return err
				}
			}
			msgState = pgproto3.Describe{}
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
		case *pgproto3.Query:
			if strings.HasPrefix(msg.String, "--") && strings.HasSuffix(msg.String, "ping") {
				writeMessages(conn,
					&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
					&pgproto3.ReadyForQuery{TxStatus: 'I'})
				break
			}
		default:
			return fmt.Errorf("unexpected message type during parse: %#v", msg)
		}
	}
}

// Will convert parser commands types to database statement commands.
func convertToStmtCmd(stmtResult parser.ParserStmtResult) db.SqlCmdType {
	// Transaction commands.
	if stmtResult.TxCmd != pg_query.TransactionStmtKind_TRANSACTION_STMT_KIND_UNDEFINED {
		switch stmtResult.TxCmd {
		case pg_query.TransactionStmtKind_TRANS_STMT_BEGIN:
			return db.CMD_BEGIN

		case pg_query.TransactionStmtKind_TRANS_STMT_COMMIT:
			return db.CMD_COMMIT

		case pg_query.TransactionStmtKind_TRANS_STMT_ROLLBACK:
			return db.CMD_ROLLBACK
		}
	}
	// Common SQL commands.
	if stmtResult.SqlCmd != pg_query.CmdType_CMD_TYPE_UNDEFINED {
		switch stmtResult.SqlCmd {
		case pg_query.CmdType_CMD_SELECT:
			return db.CMD_SELECT

		case pg_query.CmdType_CMD_INSERT:
			return db.CMD_INSERT

		case pg_query.CmdType_CMD_DELETE:
			return db.CMD_DELETE

		case pg_query.CmdType_CMD_UPDATE:
			return db.CMD_UPDATE
		}
	}
	return ""
}

func (s *DBServer) execSetQuery(ctx context.Context, conn *ClientConn, query string) error {
	buf, _ := (&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")}).Encode(nil)
	buf, _ = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
	_, err := conn.Write(buf)
	return err
}
