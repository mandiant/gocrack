package parent

import (
	"errors"
	"os"
	"sync"

	"github.com/fireeye/gocrack/server/rpc"
	"github.com/fireeye/gocrack/shared"
	"github.com/fireeye/gocrack/worker"

	"github.com/rs/zerolog/log"
)

var errNoOpenCLPlatforms = errors.New("worker: no opencl platforms")

// Worker is a gocrack process that beacons to a server and accepts new processing requests
type Worker struct {
	// unexported or filtered fields below
	cfg        *worker.Config
	stop       chan bool
	wg         *sync.WaitGroup
	rc         rpc.GoCrackRPC
	procs      ProcessesByTask
	devices    shared.DeviceMap
	edebugmsgs RemoteOutput
}

// New creates a new parent worker
func New(cfg *worker.Config) *Worker {
	ctx := &Worker{
		cfg:  cfg,
		stop: make(chan bool, 1),
		wg:   &sync.WaitGroup{},
	}

	if cfg.EngineDebug {
		ctx.edebugmsgs = make(RemoteOutput, 100)
	}
	return ctx
}

// engineDebugger will print out messages from our child processes if cfg.EngineDebug is true
func (s *Worker) engineDebugger() {
	defer s.wg.Done()
	defer close(s.edebugmsgs)

DebugLoop:
	for {
		select {
		case <-s.stop:
			break DebugLoop
		case msg := <-s.edebugmsgs:
			log.Info().
				Bool("from_pipe", true).
				Int("from_pid", msg.Pid).
				Bool("from_err", msg.IsStdErr).Msg(msg.Data)
		}
	}
}

// Start establishes a connection to the RPC server and starts the beaconing process. This blocks until Stop is called
func (s *Worker) Start() error {
	log.Info().Msg("Obtaining information about OpenCL devices")
	devs, err := worker.GetAvailableDevices()
	if err != nil {
		return err
	}

	if e := log.Debug(); e.Enabled() {
		for _, device := range devs {
			log.Debug().Str("name", device.Name).Int("id", device.ID).Msg("Found Device")
		}
	}

	s.devices = devs
	s.procs = NewProcessesByTask()

	client, err := worker.InitRPCChannel(*s.cfg)
	if err != nil {
		return err
	}
	s.rc = client

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	if s.cfg.EngineDebug {
		s.wg.Add(1)
		go s.engineDebugger()
	}

	s.wg.Add(1)
	go s.beacon(hostname)

	s.wg.Wait()
	return nil
}

// Stop the server by closing channels, waiting for goroutines to exit, and closing the connections out
func (s *Worker) Stop() error {
	for taskid, proc := range s.procs.data {
		log.Warn().
			Int("pid", proc.pid).
			Str("taskid", taskid).
			Msg("Sending SIGTERM to Child Process")
		proc.StopTerm()
		proc.Wait()
	}

	close(s.stop)

	log.Warn().Msg("Waiting for everything to exit...")
	s.wg.Wait()

	return nil
}
