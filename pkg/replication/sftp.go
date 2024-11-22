package replication

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type OnConnCloseFunc func(remoteAddr string)

type ServerOptions struct {
	// Secret that will grant access to the replication SFTP service.
	Secret string

	// Path to the RSA private key.
	PrivateKeyFile string

	// Bind (Host:Port) address to listen on.
	Address string

	// Server working directory to write and serve from.
	WorkingDir string
}

type ReplicationServer struct {
	// Connection listener.
	listener net.Listener

	// Bind (Host:Port) address to listen on.
	address string

	// Manage group of go routines.
	group errgroup.Group

	// Server context.
	ctx    context.Context
	cancel func()

	sshConfig ssh.ServerConfig

	// Server working directory to write and serve from.
	workDir string

	// Function callback executed when a sftp connection is closed.
	OnConnClose OnConnCloseFunc
}

func NewServer(opts ServerOptions) (*ReplicationServer, error) {
	if opts.PrivateKeyFile == "" {
		return nil, errors.New("No private key file provided")
	}

	if opts.Address == "" {
		opts.Address = "0.0.0.0:2022"
	}

	if opts.WorkingDir == "" {
		opts.WorkingDir = "/var/kqlite"
	}

	// An SSH server is represented by a ServerConfig,
	// which holds certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == opts.Secret {
				return nil, nil
			}
			return nil, fmt.Errorf("Rejected secret %q", c.User())
		},
	}

	privateBytes, err := os.ReadFile(opts.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return err
	}

	config.AddHostKey(private)

	// Configure context.
	ctx, cancel = context.WithCancel(context.Background())

	return &ReplicationServer{
		ctx:       ctx,
		cancel:    cancel,
		sshConfig: config,
		address:   opts.Address,
		workDir:   opts.WorkingDir,
	}, nil
}

// Run the server, ready to accept connections.
func (rs *ReplicationServer) Open() error {
	rs.listener, err := net.Listen("tcp", rs.address)
	if err != nil {
		log.Fatal("failed to listen for connection", err)
		return err
	}
	fmt.Printf("Listening on %v\n", listener.Addr())

	rs.group.Go(func() error {
		if err := rs.serve(); rs.ctx.Err() != nil {
			return err // return error unless context cancelled
		}
		return nil
	})
	return nil
}

func (rs *ReplicationServer) serve() error {
	for {
		newConn, err := rs.listener.Accept()
		if err != nil {
			log.Fatal("failed to accept incoming connection", err)
			return err
		}

		// Before use, a handshake must be performed on the incoming net.Conn.
		_, chans, reqs, err := ssh.NewServerConn(newConn, config)
		if err != nil {
			log.Fatal("failed to handshake", err)
		}
		fmt.Printf("SSH server connection established\n")

		// The incoming Request channel must be serviced.
		go ssh.DiscardRequests(reqs)

		// Handle SFTP session.
		rs.group.Go(func() {
			// Cleanup when connection is closed.
			defer func() {
				remoteAddr := newConn.RemoteAddr().String()
				newConn.Close()
				if rs.OnConnClose != nil {
					rs.OnConnClose(remoteAddr)
				}
			}()

			// Service the incoming Channel channel.
			for newChannel := range chans {
				// Channels have a type, depending on the application level
				// protocol intended. In the case of an SFTP session, this is "subsystem"
				// with a payload string of "<length=4>sftp"
				fmt.Printf("Incoming channel: %s\n", newChannel.ChannelType())
				if newChannel.ChannelType() != "session" {
					newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
					fmt.Printf("Unknown channel type: %s\n", newChannel.ChannelType())
					continue
				}
				channel, requests, err := newChannel.Accept()
				if err != nil {
					log.Fatal("could not accept channel.", err)
				}
				fmt.Printf("Channel accepted\n")

				// Sessions have out-of-band requests such as "shell",
				// "pty-req" and "env".  Here we handle only the
				// "subsystem" request.
				go func(in <-chan *ssh.Request) {
					for req := range in {
						fmt.Printf("Request: %v\n", req.Type)
						ok := false
						switch req.Type {
						case "subsystem":
							fmt.Printf("Subsystem: %s\n", req.Payload[4:])
							if string(req.Payload[4:]) == "sftp" {
								ok = true
							}
						}
						fmt.Printf(" - accepted: %v\n", ok)
						req.Reply(ok, nil)
					}
				}(requests)

				serverOptions := []sftp.ServerOption{
					sftp.WithServerWorkingDirectory(rs.workDir),
				}

				sftpServer, err := sftp.NewServer(
					channel,
					serverOptions...,
				)
				defer sftpServer.Close()

				if err != nil {
					log.Fatal(err)
				}
				if err := sftpServer.Serve(); err != nil {
					if err != io.EOF {
						log.Fatal("sftp server completed with error:", err)
					}
				}

				log.Print("sftp client exited session.")
			}
		})
	}
}
