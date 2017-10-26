package rpc

import (
	"errors"
	"io"

	"github.com/fireeye/gocrack/shared"
)

// Config describes the configuration used by the GoCrack RPC Listener.
type Config struct {
	Listener shared.ServerCfg `yaml:"listener"`
}

// ErrNoCheckpoint is returned when a checkpoint does not exist for the task
var ErrNoCheckpoint = errors.New("rpc: no checkpoint file for task")

// Validate the config and set any default options
func (s *Config) Validate() error {
	if s.Listener.Address == "" {
		s.Listener.Address = ":4014"
	}

	if s.Listener.Certificate == "" || s.Listener.PrivateKey == "" {
		return errors.New("rpc_server.listener.ssl_certificate and rpc_server.listener.ssl_private_key must not be empty")
	}
	return nil
}

// GoCrackRPC describes all the methods that a worker uses to send updates to the server
type GoCrackRPC interface {
	Beacon(BeaconRequest) (*BeaconResponse, error)
	ChangeTaskStatus(ChangeTaskStatusRequest) error
	GetTask(RequestTaskPayload) (*NewTaskPayloadResponse, error)
	GetFile(TaskFileGetRequest) (io.ReadCloser, string, error)
	SavedCrackedPassword(CrackedPasswordRequest) error
	SendTaskStatus(TaskStatusUpdate) error
	GetCheckpointFile(string) ([]byte, error)
	SendCheckpointFile(TaskCheckpointSaveRequest) error
}
