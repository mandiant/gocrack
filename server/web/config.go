package web

import (
	"errors"
	"time"

	"github.com/fireeye/gocrack/shared"
)

type corsSettings struct {
	AllowedOrigins  []string              `yaml:"allowed_origins"`
	MaxPreflightAge *shared.HumanDuration `yaml:"max_preflight_age,omitempty"`
}

type listener struct {
	Address       string  `yaml:"address"`
	Certificate   string  `yaml:"ssl_certificate"`
	PrivateKey    string  `yaml:"ssl_private_key"`
	CACertificate *string `yaml:"ssl_ca_certificate,omitempty"`
	UseSSL        bool    `yaml:"ssl_enabled"`
}

type uiSettings struct {
	StaticPath  string `yaml:"static_path"`
	CSRFEnabled *bool  `yaml:"csrf_enabled,omitempty"`
	CSRFKey     string `yaml:"csrf_key"`
}

// Config describes the various options available to the API server
type Config struct {
	Listener      listener     `yaml:"listener"`
	CORS          corsSettings `yaml:"cors"`
	UserInterface uiSettings   `yaml:"ui"`
}

var (
	errMissingCORSOrigins   = errors.New("web_server.cors.allowed_origins must contain one or more domains")
	errPreflightAgeNegative = errors.New("web_server.cors.max_preflight_age must be a posititve duration")
	errMissingCSRFKey       = errors.New("web_server.ui.csrf_key must be a secure key")

	defaultPreflightAge    = &shared.HumanDuration{Duration: 24 * time.Hour}
	defaultListenerAddress = ":4013"
)

// Validate the API server configuration
func (s *Config) Validate() error {
	if s.Listener.Address == "" {
		s.Listener.Address = defaultListenerAddress
	}

	if len(s.CORS.AllowedOrigins) == 0 {
		return errMissingCORSOrigins
	}

	if s.UserInterface.CSRFEnabled == nil {
		s.UserInterface.CSRFEnabled = shared.GetBoolPtr(true)
	}

	// Check and see if the key is set if CSRF is Enabled
	if *s.UserInterface.CSRFEnabled && s.UserInterface.CSRFKey == "" {
		return errMissingCSRFKey
	}

	if s.CORS.MaxPreflightAge == nil || s.CORS.MaxPreflightAge.Duration.Nanoseconds() < 0 {
		s.CORS.MaxPreflightAge = defaultPreflightAge
	}

	if s.UserInterface.StaticPath == "" {
		s.UserInterface.StaticPath = "./static"
	}

	return nil
}
