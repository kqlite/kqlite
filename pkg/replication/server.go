package replication

import (
	"net"
	"net/http"

	"github.com/go-logr/logr"
	// "github.com/kqlite/kqlite/pkg/utils"
)

type StreamServer struct {
	httpServer http.Server
	address    string       // Bind address of the HTTP service.
	listener   net.Listener // Service listener

	CertFile     string // Path to server's own x509 certificate.
	KeyFile      string // Path to server's own x509 private key.
	ClientVerify bool   // Whether client certificates should verified.
	Log          logr.Logger
}

// Returns an uninitialized replication server. If credentials is nil, then
// the server performs no authentication and authorization checks.
func New(address string) *StreamServer {
	return &StreamServer{
		address: address,
		//logger:  log.New(os.Stderr, "[replication] ", log.LstdFlags),
	}
}

func (s *StreamServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// Starts the service.
func (s *StreamServer) Start() error {
	s.httpServer = http.Server{
		Handler: s,
	}

	var listener net.Listener
	var err error

	if s.CertFile == "" || s.KeyFile == "" {
		listener, err = net.Listen("tcp", s.address)
		if err != nil {
			return err
		}
	}

	s.listener = listener

	go func() {
		err := s.httpServer.Serve(s.listener)
		if err != nil {
			s.Log.Info("HTTP service on", s.listener.Addr().String(), "stopped:", err.Error())
		}
	}()
	s.Log.Info("service listening on", s.listener.Addr())

	return nil
}
