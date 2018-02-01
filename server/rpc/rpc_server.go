package rpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/server/workmgr"
	"github.com/fireeye/gocrack/shared"
	"github.com/fireeye/gocrack/shared/ginlog"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/tankbusta/gzip"
)

type (
	RPCServer struct {
		stor   storage.Backend
		wmgr   *workmgr.WorkerManager
		engine *gin.Engine
		l      net.Listener
		cfg    Config

		*http.Server
	}

	RPCError struct {
		StatusCode int
		Err        error
	}

	RPCAPI func(c *gin.Context) *RPCError
)

func WrapCallError(f RPCAPI) gin.HandlerFunc {
	return func(c *gin.Context) {
		if werr := f(c); werr != nil {
			evt := log.Error().Int("status_code", werr.StatusCode)

			if werr.Err != nil {
				evt.Err(werr.Err)
			}

			evt.Msg("An error occurred while handling an RPC call")
			c.JSON(werr.StatusCode, werr)
		}
	}
}

// NewRPCServer creates an RPC server that workers connect to
func NewRPCServer(cfg Config, stor storage.Backend, wmgr *workmgr.WorkerManager) (*RPCServer, error) {
	var l net.Listener
	var err error

	if cfg.Listener.UseSSL {
		tcfg, err := shared.GetTLSConfig(cfg.Listener.Certificate, cfg.Listener.PrivateKey, cfg.Listener.CACertificate)
		if err != nil {
			return nil, fmt.Errorf("Error creating RPC TLS listener: %s", err)
		}

		// Change a few settings on the TLS config for the listener
		tcfg.ServerName = cfg.Listener.Address
		tcfg.ClientAuth = tls.RequireAnyClientCert
		tcfg.MinVersion = tls.VersionTLS12
		tcfg.PreferServerCipherSuites = true
		l, err = tls.Listen("tcp", cfg.Listener.Address, tcfg)
	} else {
		l, err = net.Listen("tcp", cfg.Listener.Address)
	}

	if err != nil {
		return nil, err
	}

	svr := &RPCServer{
		stor: stor,
		wmgr: wmgr,
		l:    l,
	}
	svr.initRPCEngineAndServer()

	return svr, nil
}

// Start the RPC server and handle requests from workers
func (s *RPCServer) Start() error {
	return s.Serve(s.l)
}

// Address returns the listening address of the RPC server
func (s *RPCServer) Address() string {
	return s.l.Addr().String()
}

// Stop the RPC server gracefully
func (s *RPCServer) Stop() error {
	if err := s.Shutdown(context.Background()); err != nil {
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
	return nil
}

func (s *RPCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.engine.ServeHTTP(w, r)
}

func (s *RPCServer) initRPCEngineAndServer() {
	s.engine = gin.New()
	s.engine.Use(gin.Recovery(), ginlog.LogRequests(), gzip.Gzip(gzip.DefaultCompression))

	routes := s.engine.Group("/rpc/v1").Use(shared.RecordAPIMetrics(requestDuration, requestCounter))
	{
		routes.POST("/beacon", WrapCallError(s.workerBeacon))
		routes.POST("/task/status_change", WrapCallError(s.changeTaskStatus))
		routes.POST("/task/checkpoint", WrapCallError(s.taskRestoreFile))
		routes.GET("/task/checkpoint/:taskid", WrapCallError(s.getCheckpointFile))
		routes.POST("/task/payload", WrapCallError(s.getTaskPayload))
		routes.POST("/task/cracked", WrapCallError(s.saveCrackedPassword))
		routes.POST("/task/status", WrapCallError(s.taskStatusUpdate))
		routes.POST("/file", WrapCallError(s.getTaskFile))
	}

	s.Server = &http.Server{Handler: s.engine}
}
