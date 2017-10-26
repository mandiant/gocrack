package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/fireeye/gocrack/server"
	"github.com/fireeye/gocrack/shared"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configPath string
	debug      bool
)

func init() {
	flag.StringVar(&configPath, "config", "configs/server.yaml", "path to configuration file")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func main() {
	var cfg server.Config

	if err := shared.LoadConfigFile(configPath, &cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration file")
		return
	}

	// Enable the debug flag in the config if false but the user set the cmdline flag
	if !cfg.Debug && debug {
		cfg.Debug = debug
	}

	if cfg.Debug {
		logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
		log.Logger = logger
	}

	svr, err := server.New(&cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize server")
	}

	go func() {
		if err := svr.Start(); err != nil {
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to start server")
			}
		}
	}()

	refresh := make(chan os.Signal, 1)
	signal.Notify(refresh, syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		for {
			select {
			case <-refresh:
				log.Warn().Msg("Refreshing server...")
				svr.Refresh()
			}
		}
	}()

	// Normal Mode
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		log.Warn().Msg("Caught signal, shutting down")
		svr.Stop()
	}
}
