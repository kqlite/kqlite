package replication

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/benbjohnson/litestream"
	lssftp "github.com/benbjohnson/litestream/sftp"
)

// Options are the arguments for creating a new litestream replication connection.
type LiteStreamOptions struct {
	// Target host to replicate the database to, could be in the form host:port.
	Host string

	// Destination path on the SFTP server to store replicated data.
	DestinationPath string

	// Local data source to replicate from.
	DataSourceName string

	// Secret that will be used to connect to the replication SFTP service.
	Secret string
}

// Replication destination endpoint.
type LiteStream struct {
	// Litestream db instance.
	db *litestream.DB
}

// Create and configure the SFTP litestream replication endpoint.
func NewLiteStream(opts LiteStreamOptions) (*LiteStream, error) {
	if opts.DataSourceName == "" {
		return nil, errors.New("Empty data source")
	}

	if opts.Host == "" {
		opts.Host = "localhost:2022"
	}

	if opts.DestinationPath == "" {
		opts.DestinationPath = "/var/kqlite"
	}

	if opts.Secret == "" {
		opts.Secret = "litestream"
	}

	// Create Litestream DB reference for managing replication.
	lsdb := litestream.NewDB(opts.DataSourceName)

	// Build SFTP replica and attach to database.
	client := lssftp.NewReplicaClient()
	// Configure SFTP client
	client.Host = opts.Host
	client.User = opts.Secret
	client.Path = opts.DestinationPath

	replica := litestream.NewReplica(lsdb, "sftp")
	replica.Client = client
	lsdb.Replicas = append(lsdb.Replicas, replica)

	// Initialize database.
	if err := lsdb.Open(); err != nil {
		return nil, err
	}

	return &LiteStream{db: lsdb}, nil
}

// Sync local database to remote replica.
func (ls *LiteStream) Sync(ctx context.Context) error {
	if err := ls.db.Sync(ctx); err != nil {
		return err
	}

	// Sync litestream with remote SFTP.
	startTime := time.Now()
	if err := ls.db.Replicas[0].Sync(ctx); err != nil {
		return err
	}

	// TODO use log
	fmt.Printf("completed sync, elapsed=%s", time.Since(startTime))

	return nil
}

// Restore local database from remote replica,
// if there is a existing local database with the same name it will get overridden.
func (ls *LiteStream) Restore(ctx context.Context) error {
	replica := ls.db.Replicas[0]

	// Remove local database if already exists.
	localdbPath := replica.DB().Path()
	if _, err := os.Stat(localdbPath); err == nil {
		if err := os.Remove(localdbPath); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// Configure restore to write out to the local datasource path.
	opt := litestream.NewRestoreOptions()
	opt.OutputPath = localdbPath

	// Determine the latest generation to restore from.
	var err error
	if opt.Generation, _, err = replica.CalcRestoreTarget(ctx, opt); err != nil {
		return err
	}

	// Only restore if there is a generation available on the replica.
	// Otherwise we'll let the application create a new database.
	if opt.Generation == "" {
		fmt.Println("no generation found, creating new database")
		return nil
	}

	fmt.Printf("restoring replica for generation %s\n", opt.Generation)
	if err := replica.Restore(ctx, opt); err != nil {
		return err
	}
	fmt.Println("restore complete")

	return nil
}
