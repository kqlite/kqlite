package pgwire

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	//"strings"
	//"time"

	//"github.com/kqlite/kqlite/pkg/util/pgerror"

	"golang.org/x/sync/errgroup"
	//"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
)

func writeDataToFile(rd io.Reader) error {
	filename := "dbfile.tmp"

	mode := os.FileMode(0644)
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, rd); err != nil {
		return err
	} else if err := f.Sync(); err != nil {
		return err
	} else if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// Handles CopyFrom statement as a special command used for replication.
// It will receive a byte copy of the target database from the client.
func (conn *ClientConn) handleCopy(ctx context.Context, msg *pgproto3.Query) error {
	defer timer("handleCopy")()

	log.Printf("handleCopy \n")

	if err := writeMessages(conn, &pgproto3.CopyOutResponse{OverallFormat: byte(1), ColumnFormatCodes: []uint16{pgtype.ByteaOID}}); err != nil {
		return err
	}

	// Use a pipe to send data from client to a reader.
	pr, pw := io.Pipe()

	// Copy the database file to the LZ4 writer in a separate goroutine.
	var g errgroup.Group
	g.Go(func() error {
		if err := writeDataToFile(pr); err != nil {
			return err
		}
		return nil
	})

	for {
		msg, err := conn.backend.Receive()
		if err != nil {
			return fmt.Errorf("receive message: %w", err)
		}

		//log.Printf("[%s] [recv] %#v", conn.RemoteAddr().String(), msg)

		switch cpmsg := msg.(type) {
		case *pgproto3.CopyData:
			if _, err := pw.Write(cpmsg.Data); err != nil {
				return nil
			}

		case *pgproto3.CopyDone:
			pw.Close()
			writeMessages(conn,
				&pgproto3.CommandComplete{CommandTag: []byte("COPY")},
				&pgproto3.ReadyForQuery{TxStatus: 'I'})
			return g.Wait()

		default:
			return fmt.Errorf("unexpected message type: %#v", msg)
		}
	}

	// return nil
}
