package worker

import (
	"errors"
	"time"

	"github.com/fireeye/gocrack/shared"
)

var (
	defBeaconInterval    = &shared.HumanDuration{Duration: time.Second * 30}
	defJobInterval       = &shared.HumanDuration{Duration: time.Second * 5}
	defTermDelayInterval = &shared.HumanDuration{Duration: time.Second * 30}

	// Default Max # of GPUs that will be assigned for a task given it's priority
	defNumGPUHigh   = shared.GetIntPtr(4)
	defNumGPUNormal = shared.GetIntPtr(2)
	defNumGPULow    = shared.GetIntPtr(1)
)

type serverConfig struct {
	Address       string  `yaml:"connect_to"`
	Certificate   string  `yaml:"ssl_certificate"`
	CACertificate string  `yaml:"ssl_ca_certificate"`
	PrivateKey    string  `yaml:"ssl_private_key"`
	ServerName    *string `yaml:"server_name,omitempty"`
}

// Config describes the worker configuration variables
type Config struct {
	EngineDebug        bool         `yaml:"engine_debug"`
	AutoCPUAssignment  bool         `yaml:"auto_cpu_assignment"`
	ServerConn         serverConfig `yaml:"server"`
	SaveTaskFilePath   string       `yaml:"save_task_file_path"`
	SaveEngineFilePath string       `yaml:"save_engine_file_path"`
	Hashcat            struct {
		LogPath     string `yaml:"log_path"`
		PotfilePath string `yaml:"potfile_path"`
		SessionPath string `yaml:"session_path"`
		SharedPath  string `yaml:"shared_path"`
	} `yaml:"hashcat"`
	Intervals struct {
		// Beacon interval sets how frequently we beacon to the server with an overview of what we're doing
		Beacon *shared.HumanDuration `yaml:"beacon,omitempty"`
		// JobStatus interval sets how frequently we send job status messages to the server
		JobStatus *shared.HumanDuration `yaml:"job_status,omitempty"`
		// TerminationDelay is the time we wait for hashcat to exit when we are told to exit
		TerminationDelay *shared.HumanDuration `yaml:"termination_delay,omitempty"`
	} `yaml:"intervals,omitempty"`
	GPUPriorityAssignment struct {
		High   *int `yaml:"high,omitempty"`
		Normal *int `yaml:"normal,omitempty"`
		Low    *int `yaml:"low,omitempty"`
	} `yaml:"gpus_priority_limit"`
}

// Validate the worker config, set default values if none are present, and return any fatal config errors
func (s *Config) Validate() error {
	if s.ServerConn.Address == "" {
		return errors.New("server.connect_to must not be empty")
	}

	if s.SaveTaskFilePath == "" {
		return errors.New("save_task_path must not be empty")
	}

	if s.SaveEngineFilePath == "" {
		return errors.New("save_engine_file_path must not be empty")
	}

	if s.ServerConn.Certificate == "" || s.ServerConn.PrivateKey == "" {
		return errors.New("server.ssl_certficate and server.ssl_private_key must not be empty")
	}

	if s.Intervals.Beacon == nil {
		s.Intervals.Beacon = defBeaconInterval
	}

	if s.Intervals.JobStatus == nil {
		s.Intervals.JobStatus = defJobInterval
	}

	if s.Intervals.TerminationDelay == nil {
		s.Intervals.TerminationDelay = defTermDelayInterval
	}

	if s.GPUPriorityAssignment.High == nil {
		s.GPUPriorityAssignment.High = defNumGPUHigh
	}

	if s.GPUPriorityAssignment.Normal == nil {
		s.GPUPriorityAssignment.Normal = defNumGPUNormal
	}

	if s.GPUPriorityAssignment.Low == nil {
		s.GPUPriorityAssignment.Low = defNumGPULow
	}

	return nil
}
