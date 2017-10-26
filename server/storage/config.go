package storage

import "errors"

var (
	noBackend          = errors.New("database.backend must not be empty")
	noConnectionString = errors.New("database.connection_string must not be empty")
)

type Config struct {
	Backend          string `yaml:"backend"`
	ConnectionString string `yaml:"connection_string"`
}

func (s *Config) Validate() error {
	if s.Backend == "" {
		return noBackend
	}

	if s.ConnectionString == "" {
		return noConnectionString
	}
	return nil
}
