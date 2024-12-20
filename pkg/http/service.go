// Provides the HTTP server for implementing replication protocol and serving stats.
package http

import (
	"context"
	"log"
	"net"
	"net/http"
	// "net/url"
	"os"
	"strings"
)

// HTTP Service.
type Service struct {
	httpServer http.Server
	closeChan  chan struct{}
	address    string       // Bind address of the HTTP service.
	listener   net.Listener // Service listener

	CertFile     string // Path to server's own x509 certificate.
	KeyFile      string // Path to server's own x509 private key.
	ClientVerify bool   // Whether client certificates should verified.

	logger *log.Logger
}

const (
	// VersionHTTPHeader is the HTTP header key for the version.
	VersionHTTPHeader = "X-KQLITE-VERSION"

	// ServedByHTTPHeader is the HTTP header used to report which
	// node (Source or Replica) actually served the request.
	ServedByHTTPHeader = "X-KQLITE-SERVED-BY"

	// AllowOriginHeader is the HTTP header for allowing CORS compliant access from certain origins
	AllowOriginHeader = "Access-Control-Allow-Origin"

	// AllowMethodsHeader is the HTTP header for supporting the correct methods
	AllowMethodsHeader = "Access-Control-Allow-Methods"

	// AllowHeadersHeader is the HTTP header for supporting the correct request headers
	AllowHeadersHeader = "Access-Control-Allow-Headers"

	// AllowCredentialsHeader is the HTTP header for supporting specifying credentials
	AllowCredentialsHeader = "Access-Control-Allow-Credentials"
)

// Returns an uninitialized HTTP service. If credentials is nil, then
// the service performs no authentication and authorization checks.
func New(address string) *Service {
	return &Service{
		address: address,
		logger:  log.New(os.Stderr, "[http] ", log.LstdFlags),
	}
}

// Starts the service.
func (s *Service) Start() error {
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
	s.closeChan = make(chan struct{})

	go func() {
		err := s.httpServer.Serve(s.listener)
		if err != nil {
			s.logger.Printf("HTTP service on %s stopped: %s", s.listener.Addr().String(), err.Error())
		}
	}()
	s.logger.Println("service listening on", s.listener.Addr())

	return nil
}

// Stops the service.
func (s *Service) Stop() {
	s.logger.Println("closing HTTP service on", s.listener.Addr().String())
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		s.logger.Println("HTTP service shutdown error:", err.Error())
	}
}

// ServeHTTP allows Service to serve HTTP requests, implements the http.Handler interface.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// s.addAllowHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var err error
	// params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch {
	case r.URL.Path == "/" || r.URL.Path == "":
		http.Redirect(w, r, "/status", http.StatusFound)
	case strings.HasPrefix(r.URL.Path, "/db/backup"):
		// s.handleBackup(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/db/replicate"):
		// s.handleReplicate(w, r, params)
	case strings.HasPrefix(r.URL.Path, "/status"):
		// s.handleStatus(w, r, params)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
