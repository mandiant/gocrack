package hashcat

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/fireeye/gocrack/gocat"
	"github.com/fireeye/gocrack/gocat/hcargp"
	"github.com/fireeye/gocrack/gocat/restoreutil"
	"github.com/fireeye/gocrack/opencl"
	"github.com/fireeye/gocrack/server/rpc"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"
	"github.com/fireeye/gocrack/worker"

	"github.com/rs/zerolog/log"
)

type HashcatEngine struct {
	TaskID            string
	HashcatSharedPath string
	SessionPath       string
	PotfilePath       string

	// Task Options
	TaskFilePath   string
	DictionaryFile string
	MasksFile      string
	RulesFile      string
	Options        shared.HashcatUserOptions
	Upstream       rpc.GoCrackRPC
	CLDevices      storage.CLDevices

	engine *gocat.Hashcat
	// if isBruteForce is true, we'll allow for a checkpoint
	isBruteForce bool
}

// Initialize the hashcat engine
func (s *HashcatEngine) Initialize() error {
	hc, err := gocat.New(gocat.Options{
		SharedPath: s.HashcatSharedPath,
	}, s.callback())

	if err != nil {
		return err
	}
	s.engine = hc
	return nil
}

func (s *HashcatEngine) callback() gocat.EventCallback {
	return func(hc unsafe.Pointer, payload interface{}) {
		switch pl := payload.(type) {
		case gocat.LogPayload:
			fmt.Printf("LOG [%s] %s\n", pl.Level, pl.Message)
		case gocat.ActionPayload:
			fmt.Printf("ACTION [%d] %s\n", pl.HashcatEvent, pl.Message)
		case gocat.CrackedPayload:
			s.Upstream.SavedCrackedPassword(rpc.CrackedPasswordRequest{
				TaskID:    s.TaskID,
				Hash:      pl.Hash,
				Value:     pl.Value,
				CrackedAt: pl.CrackedAt,
			})
		case gocat.FinalStatusPayload:
			if pl.AllHashesCracked {
				if err := s.Upstream.ChangeTaskStatus(rpc.ChangeTaskStatusRequest{
					TaskID:    s.TaskID,
					NewStatus: storage.TaskStatusFinished,
				}); err != nil {
					log.Error().Err(err).Msg("All hashes cracked but failed to set final status")
				}
				return
			}

			// Safeguard us from panic'ing here when the task fails to start
			if pl.Status == nil {
				return
			}

			var tstatus storage.TaskStatus
			// https://github.com/hashcat/hashcat/blob/master/src/status.c#L22-L35
			switch pl.Status.Status {
			case "Cracked":
				tstatus = storage.TaskStatusFinished
			case "Exhausted":
				tstatus = storage.TaskStatusExhausted
			case "Unknown! Bug!":
				tstatus = storage.TaskStatusError
			case "Quit", "Aborted", "Paused", "Aborted (Checkpoint)", "Aborted (Runtime)":
				fallthrough
			default:
				tstatus = storage.TaskStatusStopped
			}

			if err := s.Upstream.ChangeTaskStatus(rpc.ChangeTaskStatusRequest{
				TaskID:    s.TaskID,
				NewStatus: tstatus,
			}); err != nil {
				log.Error().Err(err).Str("new_status", string(tstatus)).Msg("Failed to change task status")
			}

			if err := s.Upstream.SendTaskStatus(rpc.TaskStatusUpdate{
				TaskID:  s.TaskID,
				Payload: pl.Status,
				Engine:  storage.WorkerHashcatEngine,
				Final:   true,
			}); err != nil {
				log.Error().Err(err).Msg("Failed to send task status update to server")
			}
		}
	}
}

// Start the hashcat engine
func (s *HashcatEngine) Start() error {
	var opts hcargp.HashcatSessionOptions

	restoreFilePath := filepath.Join(s.SessionPath, fmt.Sprintf("%s.restore", s.TaskID))
	outFilePath := filepath.Join(s.SessionPath, fmt.Sprintf("%s.outfile", s.TaskID))

	// Get a list of active devices on this machine
	machDevices, err := worker.GetAvailableDevices()
	if err != nil {
		return err
	}

	// Grab the checkpoint from the server if one exists
	if err := getCheckpoint(s.Upstream, s.TaskID, restoreFilePath); err != nil && err != rpc.ErrNoCheckpoint {
		return err
	}

	// If the restore file exists, set the restore flag
	if _, err := os.Stat(restoreFilePath); !os.IsNotExist(err) {
		// When restoring, we can really only set restore related options
		opts = hcargp.HashcatSessionOptions{
			RestoreSession:  hcargp.GetBoolPtr(true),
			SessionName:     hcargp.GetStringPtr(s.TaskID),
			RestoreFilePath: hcargp.GetStringPtr(restoreFilePath),
		}

		checkpoint, err := restoreutil.ReadRestoreFile(restoreFilePath)
		if err != nil {
			return err
		}

		log.Debug().
			Uint32("version", checkpoint.Version).
			Uint32("argc", checkpoint.ArgCount).
			Strs("args", checkpoint.Args).
			Msg("Parsed Restore File")

		changed, err := ModifyRestoreFileDevices(&checkpoint, s.CLDevices, machDevices)
		if err != nil {
			return err
		}

		if changed {
			if err := changeLocalCheckpoint(checkpoint, restoreFilePath); err != nil {
				return err
			}
		}
	} else {
		// Not a restore
		opts = hcargp.HashcatSessionOptions{
			AttackMode:      hcargp.GetIntPtr(int(s.Options.AttackMode)),
			HashType:        hcargp.GetIntPtr(s.Options.HashType),
			PotfileDisable:  hcargp.GetBoolPtr(false),
			PotfilePath:     hcargp.GetStringPtr(s.PotfilePath),
			InputFile:       s.TaskFilePath,
			SessionName:     hcargp.GetStringPtr(s.TaskID),
			RestoreFilePath: hcargp.GetStringPtr(restoreFilePath),
			OutfilePath:     hcargp.GetStringPtr(outFilePath),
		}

		if s.MasksFile != "" && s.Options.AttackMode == shared.AttackModeBruteForce {
			opts.DictionaryMaskDirectoryInput = hcargp.GetStringPtr(s.MasksFile)
			s.isBruteForce = true
		}

		if s.RulesFile != "" && s.Options.AttackMode == shared.AttackModeStraight {
			opts.RulesFile = hcargp.GetStringPtr(s.RulesFile)
		}

		if s.DictionaryFile != "" && s.Options.AttackMode == shared.AttackModeStraight {
			opts.DictionaryMaskDirectoryInput = hcargp.GetStringPtr(s.DictionaryFile)
		}

		if len(s.CLDevices) > 0 {
			var devTypes []int
			var dGPUAdded, dCPUAdded bool

			for _, requestedDevice := range s.CLDevices {
				dev, ok := machDevices[requestedDevice]
				if !ok {
					return fmt.Errorf("`%d` is not an available device on this platform", requestedDevice)
				}

				switch dev.Type {
				case opencl.DeviceTypeCPU:
					if !dCPUAdded {
						devTypes = append(devTypes, 1)
						dCPUAdded = true
					}
				case opencl.DeviceTypeGPU:
					if !dGPUAdded {
						devTypes = append(devTypes, 2)
						dGPUAdded = true
					}
				}
			}

			opts.OpenCLDevices = hcargp.GetStringPtr(s.CLDevices.String())
			opts.OpenCLDeviceTypes = hcargp.GetStringPtr(shared.IntSliceToString(devTypes))
		}
	}

	defer saveCheckpoint(s.Upstream, s.TaskID, restoreFilePath)
	return s.engine.RunJobWithOptions(opts)
}

// Stop the hashcat engine. If the engine is brute forcing, we attempt to stop at a checkpoint otherwise we abort
func (s *HashcatEngine) Stop() error {
	if s.engine == nil {
		return nil
	}

	if s.isBruteForce {
		// Attempt to checkpoint... if we cant, immediately abort
		if err := s.engine.StopAtCheckpoint(); err != nil {
			s.engine.AbortRunningTask()
			return err
		}
	}
	s.engine.AbortRunningTask()
	return nil
}

// GetStatus returns the status of the engine
func (s *HashcatEngine) GetStatus() interface{} {
	return s.engine.GetStatus()
}

// Cleanup releases the engine and cleans up any allocated resources
func (s *HashcatEngine) Cleanup() {
	s.engine.Free()
}
