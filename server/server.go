package server

import (
	"net/http"
	"sync"

	"github.com/fireeye/gocrack/server/authentication"
	"github.com/fireeye/gocrack/server/filemanager"
	"github.com/fireeye/gocrack/server/notifications"
	"github.com/fireeye/gocrack/server/rpc"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/server/web"
	"github.com/fireeye/gocrack/server/workmgr"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	// CompileTime is when this was compiled
	CompileTime string
	// CompileRev is the git revision hash (sha1)
	CompileRev string
)

// Server is a GoCrack server instance
type Server struct {
	// unexported or filtered fields below
	auth    *authentication.AuthWrapper
	stor    storage.Backend
	wg      *sync.WaitGroup
	cfg     *Config
	workers *workmgr.WorkerManager
	fm      *filemanager.Context
	stop    chan bool
}

// New returns a gocrack server context. It creates the necessary listeners and opens a connection to the storage and authentication backend
func New(cfg *Config) (*Server, error) {
	var err error
	if err = cfg.validate(); err != nil {
		return nil, err
	}

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	stor, err := storage.Open(cfg.Database)
	if err != nil {
		return nil, err
	}

	ctx := &Server{
		wg:      &sync.WaitGroup{},
		cfg:     cfg,
		workers: workmgr.NewWorkerManager(),
		stop:    make(chan bool, 1),
		stor:    stor,
	}

	if ctx.auth, err = authentication.Open(ctx.stor, cfg.Authentication); err != nil {
		return nil, err
	}

	ctx.fm = filemanager.New(ctx.stor, cfg.FileManager)
	return ctx, nil
}

// registerNotifications starts the notification engine and subscribes to needed topic. If the engine starts successfully
// it returns a closer function to stop & unsubscribe from topics
func (s *Server) registerNotifications() (func(), error) {
	log.Debug().Msg("Starting Notification Engine")
	emailer, err := notifications.New(s.cfg.Notification, s.stor)
	if err != nil {
		return nil, err
	}

	emTaskStatusHndl, err := s.workers.Subscribe(workmgr.TaskStatusTopic, func(payload interface{}) {
		taskStatus, ok := payload.(workmgr.TaskStatusChangeBroadcast)
		if !ok {
			log.Error().Msg("TaskStatusTopic message is not the correct type")
			return
		}

		if err := emailer.TaskStatusChanged(taskStatus.TaskID, taskStatus.Status); err != nil {
			log.Error().Err(err).Msg("Failed to send email regarding task status change")
		}
	})
	if err != nil {
		return nil, err
	}

	emCrackedPwHndl, err := s.workers.Subscribe(workmgr.CrackedTopic, func(payload interface{}) {
		crackedPassword, ok := payload.(workmgr.CrackedPasswordBroadcast)
		if !ok {
			log.Error().Msg("CrackedTopic message is not the correct type")
			return
		}

		if err := emailer.CrackedPassword(crackedPassword.TaskID); err != nil {
			log.Error().Err(err).Msg("Failed to send newly cracked password email")
		}
	})
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("Notification Engine Started")
	return func() {
		log.Debug().Msg("Stopping notification engine")

		s.workers.Unsubscribe(emTaskStatusHndl)
		s.workers.Unsubscribe(emCrackedPwHndl)
		emailer.Stop()
	}, nil
}

// Start spawns the API Server as well as the RPC Server and blocks until stop has been called
func (s *Server) Start() error {
	rs, err := rpc.NewRPCServer(s.cfg.RPC, s.stor, s.workers)
	if err != nil {
		return err
	}

	web, err := web.NewServer(s.cfg.WebServer, s.stor, s.workers, s.auth, s.fm)
	if err != nil {
		return err
	}

	if s.cfg.Notification.Enabled {
		closer, err := s.registerNotifications()
		if err != nil {
			return err
		}
		defer closer()
	}

	// If any of the goroutines that are running a listener fail, we'll send the err on this channel
	errch := make(chan error, 1)
	defer close(errch)

	s.wg.Add(2)
	go func() {
		defer s.wg.Done()

		log.Info().Str("addr", rs.Address()).Msg("RPC Server listening")
		if err := rs.Start(); err != nil {
			if err == http.ErrServerClosed {
				log.Warn().Msg("RPC server has stopped")
				return
			}
			log.Error().Err(err).Msg("Error while serving rpc server")
			errch <- err
		}
	}()

	go func() {
		defer s.wg.Done()

		log.Info().Str("addr", web.Address()).Msg("HTTP Server listening")
		if err := web.Start(); err != nil {
			if err == http.ErrServerClosed {
				log.Warn().Msg("HTTP server has stopped")
				return
			}
			log.Error().Err(err).Msg("Error while serving HTTP server")
			errch <- err
		}
	}()

	// Wait for a fatal error on a listener or a stop message
	select {
	case fatalError := <-errch:
		return fatalError
	case <-s.stop:
		rs.Stop()
		web.Stop()
	}

	return nil
}

// Stop gracefully stops the GoCrack server and waits for all the goroutines to exit
func (s *Server) Stop() error {
	close(s.stop)
	s.wg.Wait()
	// Stop the work manager after the RPC & API server have exited safefully
	s.workers.Stop()
	// Stop the storage backend
	s.stor.Close()
	return nil
}

// Refresh is called on a SIGUSR1 and refreshes the internal server state
func (s *Server) Refresh() error {
	if err := s.fm.Refresh(); err != nil {
		log.Error().Err(err).Msg("Failed to refresh filemanager")
	}
	return nil
}
