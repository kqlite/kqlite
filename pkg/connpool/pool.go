package connpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotConnected      = errors.New("Error not connected to database")
	ErrInvalidConnParams = errors.New("Invalid connection arguments provided")
)

// A pool of connections to the primary replication server.
// The pool holds a connection to every replicated database hosted on the primary server.
// Map holds instances of pgx.Pool connections.
type ReplicaPool struct {
	sync.Map
}

var rpool ReplicaPool

const (
	MaxPingRetries             = 3
	MaxPingRetryIntervalMillis = 10
)

func connHealthCheck(conn *pgx.Conn) error {
	if conn == nil {
		return nil
	}

	var err error
	for range MaxPingRetries {
		if err = conn.Ping(context.Background()); err == nil {
			break
		}
		time.Sleep(MaxPingRetryIntervalMillis * time.Millisecond)
	}
	return err
}

// Open new replication connection with target parameters host and database name.
// If connections already exists, a new connection isn't established.
func NewConnection(ctx context.Context, host, port, dbname string) error {
	if host == "" || dbname == "" {
		return ErrInvalidConnParams
	}

	if p, found := rpool.Load(dbname); found && p != nil {
		return nil
	}

	// Probe connection after connect and also send periodic health checks.
	beforeAcquire := func(ctx context.Context, conn *pgx.Conn) bool {
		if err := connHealthCheck(conn); err != nil {
			return false
		}
		return true
	}

	connString := fmt.Sprintf("user=replication password=replication host=%s port=%s dbname=%s sslmode=disable", host, port, dbname)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return err
	}
	// Always use simple query protocol for replication.
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	config.BeforeAcquire = beforeAcquire
	config.MaxConns = 1
	config.HealthCheckPeriod = 5 * time.Second
	config.MaxConnLifetime = 24 * time.Hour

	// Create associated pgxpool resources and connection.
	p, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}
	rpool.Store(dbname, p)
	return nil
}

func getPool(dbname string) (*pgxpool.Pool, error) {
	if dbname == "" {
		return nil, ErrInvalidConnParams
	}

	p, found := rpool.Load(dbname)
	if !found {
		return nil, ErrNotConnected
	}
	return p.(*pgxpool.Pool), nil
}

// Exec acquires a connection from the Pool and executes the given SQL.
// SQL is an SQL string.
// Arguments should be referenced positionally from the SQL string as $1, $2, etc.
// The acquired connection is returned to the pool when the Exec function returns.
func ExecContext(ctx context.Context, dbname string, sql string, args ...any) error {
	p, err := getPool(dbname)
	if err != nil {
		return err
	}
	_, err = p.Exec(ctx, sql, args...)
	return err
}

// Begin acquires a connection from the Pool and starts a transaction.
// Commit or Rollback must be called on the returned transaction to finalize the transaction block.
func Begin(ctx context.Context, dbname string) (pgx.Tx, error) {
	p, err := getPool(dbname)
	if err != nil {
		return nil, err
	}
	return p.BeginTx(ctx, pgx.TxOptions{})
}

// Clear and flush replication pool.
func ClearPool() {
	// Close all database connections stored in the pool.
	rpool.Range(func(key, value any) bool {
		p := value.(*pgxpool.Pool)
		p.Close()
		return true
	})
	// Clear all replication pools.
	rpool.Clear()
}

// Check if pool has a connection to the given database.
func IsConnected(dbname string) bool {
	_, err := getPool(dbname)
	return err == nil
}
