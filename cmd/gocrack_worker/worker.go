package main

import (
	"flag"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/mandiant/gocrack/shared"
	"github.com/mandiant/gocrack/worker"
	"github.com/mandiant/gocrack/worker/child"
	"github.com/mandiant/gocrack/worker/parent"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configPath   string
	taskID       string
	devicesToUse string
	isWorker     bool
	profile      bool
	debug        bool
)

func init() {
	flag.StringVar(&configPath, "config", "configs/worker.yaml", "path to configuration file")
	flag.BoolVar(&isWorker, "worker", false, "Spawn an instance of the worker process in a child mode. DO NOT USE THIS")
	flag.StringVar(&taskID, "taskid", "", "The Task ID to request for processing. Should only be used in worker mode")
	flag.StringVar(&devicesToUse, "devices", "", "Which devices to use for the task? Should only be used in worker mode")
	flag.BoolVar(&profile, "profile", false, "enable pprof? should only be used in development")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
		log.Logger = logger
	}
}

func main() {
	var cfg worker.Config
	var impl worker.WorkerImpl

	if err := shared.LoadConfigFile(configPath, &cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration file")
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Failed to validate configuration file")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	if isWorker {
		var tmp []int
		// Parse the devices string
		devices := strings.Split(devicesToUse, ",")
		for _, dev := range devices {
			deviceID, err := strconv.Atoi(dev)
			if err != nil {
				log.Fatal().Str("devices", devicesToUse).Msg("Invalid -devices argument")
			}
			tmp = append(tmp, deviceID)
		}
		impl = child.New(&cfg, taskID, tmp)
	} else {
		impl = parent.New(&cfg)
	}

	if profile {
		go func() {
			plisten, err := net.Listen("tcp", ":0")
			if err != nil {
				log.Error().Err(err).Msg("Failed to create listener for profiler")
				return
			}
			defer plisten.Close()
			log.Warn().Str("address", plisten.Addr().String()).Msg("Profile listener listening")
			http.Serve(plisten, nil)
		}()
	}

	go func() {
		<-c
		log.Warn().Msg("Caught stop signal. Shutting down...")
		impl.Stop()
	}()

	// Start and block until we're told to exit...
	if err := impl.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start worker")
	}
}
