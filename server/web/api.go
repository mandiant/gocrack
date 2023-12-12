package web

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/mandiant/gocrack/server/authentication"
	"github.com/mandiant/gocrack/server/filemanager"
	"github.com/mandiant/gocrack/server/storage"
	"github.com/mandiant/gocrack/server/workmgr"
	"github.com/mandiant/gocrack/shared"
)

var (
	serverVersion     string
	serverCompileTime string
)

type Server struct {
	stor storage.Backend
	wmgr *workmgr.WorkerManager
	auth *authentication.AuthWrapper
	netl net.Listener
	rt   *RealtimeServer
	fm   *filemanager.Context

	*http.Server
}

// NewServer creates an HTTP API server
func NewServer(cfg Config, stor storage.Backend, wmgr *workmgr.WorkerManager, auth *authentication.AuthWrapper, fm *filemanager.Context) (*Server, error) {
	var l net.Listener
	var err error

	if cfg.Listener.UseSSL {
		tcfg, err := shared.GetTLSConfig(cfg.Listener.Certificate, cfg.Listener.PrivateKey, cfg.Listener.CACertificate)
		if err != nil {
			return nil, fmt.Errorf("Error creating HTTP TLS listener: %s", err)
		}

		l, err = tls.Listen("tcp", cfg.Listener.Address, tcfg)
	} else {
		l, err = net.Listen("tcp", cfg.Listener.Address)
	}

	if err != nil {
		return nil, err
	}

	svr := &Server{
		stor: stor,
		wmgr: wmgr,
		netl: l,
		auth: auth,
		rt:   NewRealtimeServer(wmgr, stor),
		fm:   fm,
	}

	svr.Server = newHTTPServer(cfg, svr)
	return svr, nil
}

// Start the API server and block
func (s *Server) Start() error {
	return s.Serve(s.netl)
}

// SetVersionInfo sets the server version information
func SetVersionInfo(revision, ctime string) {
	serverVersion = revision
	serverCompileTime = ctime
}

// Stop the HTTP server gracefully
func (s *Server) Stop() error {
	// Stop the Realtime Streaming Server
	s.rt.Stop()
	if err := s.Shutdown(context.Background()); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Address returns the ip and port on which this server is listening on
func (s *Server) Address() string {
	return s.netl.Addr().String()
}
