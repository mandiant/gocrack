package parent

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/fireeye/gocrack/opencl"
	"github.com/fireeye/gocrack/server/rpc"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/worker"

	"github.com/rs/zerolog/log"
)

func (s *Worker) createTask(newTask rpc.NewTask) {
	maxNumGPUs := *s.cfg.GPUPriorityAssignment.Normal

	switch newTask.Priority {
	case storage.WorkerPriorityHigh:
		maxNumGPUs = *s.cfg.GPUPriorityAssignment.High
	case storage.WorkerPriorityNormal:
		maxNumGPUs = *s.cfg.GPUPriorityAssignment.Normal
	case storage.WorkerPriorityLow:
		maxNumGPUs = *s.cfg.GPUPriorityAssignment.Low
	}

	// Pick some GPUs to run the task on...
	if newTask.Devices == nil {
		freeGPUs := s.devices.PickFreeDevices(opencl.DeviceTypeGPU, maxNumGPUs)
		if len(freeGPUs) > 0 {
			log.Info().
				Interface("devices", freeGPUs).
				Str("task_id", newTask.ID).
				Msg("Automatically assigned GPUs to task")
			newTask.Devices = freeGPUs
		} else { // No GPUs are available :(
			if !s.cfg.AutoCPUAssignment {
				return
			}

			freeCPUs := s.devices.PickFreeDevices(opencl.DeviceTypeCPU, maxNumGPUs)
			if len(freeCPUs) < 0 {
				return
			}

			log.Info().
				Interface("devices", newTask.Devices).
				Str("task_id", newTask.ID).
				Msg("Automatically assigned CPUs to task")
			newTask.Devices = freeCPUs
		}
	}

	if err := s.rc.ChangeTaskStatus(rpc.ChangeTaskStatusRequest{
		TaskID:    newTask.ID,
		NewStatus: storage.TaskStatusDequeued,
	}); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to send task status of dequeued to server but still continuing execution")
	}

	// Mark the selected devices as busy
	s.devices.MarkAsBusy([]int(newTask.Devices))

	m := NewRemoteProcess(s.edebugmsgs)
	s.procs.RegisterProcess(m, newTask.ID, []int(newTask.Devices))

	args := append(os.Args[1:], []string{"--worker", "--taskid", newTask.ID, "--devices", newTask.Devices.String()}...)
	log.Info().
		Str("task_id", newTask.ID).
		Str("devices", newTask.Devices.String()).
		Str("process", os.Args[0]).
		Str("args", strings.Join(args, " ")).
		Msg("Attempting start of child process for task")

	m.Start(os.Args[0], args...)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		// Wait for the remote process to finish
		m.Wait()
		s.procs.UnregisterProcess(m, newTask.ID)
		s.devices.MarkAsFree([]int(newTask.Devices))
		log.Info().
			Str("task_id", newTask.ID).
			Int("was_on_pid", m.GetRunningPID()).
			Int("return_code", m.GetReturnCode()).
			Msg("Child process for a task has exited")
	}()
}

// beacon is the main routine for the process that checks into the server and
// parses any requests that the server might have for us
func (s *Worker) beacon(hostname string) {
	defer s.wg.Done()

	// Note: Setting a ticker to a very low value is not advised.
	log.Info().Str("every", s.cfg.Intervals.Beacon.String()).Msg("Beacon OK")
	tickEvery := time.NewTicker(s.cfg.Intervals.Beacon.Duration)

Loop:
	for {
		select {
		case <-s.stop:
			break Loop
		case <-tickEvery.C:
		}

		resp, err := s.rc.Beacon(rpc.BeaconRequest{
			WorkerVersion:  worker.CompileRev,
			Hostname:       hostname,
			Devices:        s.devices,
			RequestNewTask: s.devices.HasFreeDevices(),
			Processes:      s.procs.GetBeaconInfo(),
		})

		if err != nil {
			log.Error().Err(err).Msg("An error occurred while beaconing to the server")
			continue
		}

		// XXX: check for skew between the server & client
		if resp == nil {
			continue
		}

		for _, item := range resp.Payloads {
			switch item.Type {
			case rpc.BeaconNewTask:
				var pl rpc.NewTask
				if err := json.Unmarshal(item.Data, &pl); err != nil {
					log.Error().Err(err).Msg("Unexpected error decoding NewTask payload")
					continue
				}
				log.Debug().Str("TaskID", pl.ID).Msg("Beacon contains request to process a new task")
				s.createTask(pl)
			case rpc.BeaconChangeTaskStatus:
				var pl rpc.ChangeTaskStatus
				if err := json.Unmarshal(item.Data, &pl); err != nil {
					log.Error().Err(err).Msg("Unexpected error decoding NewTask payload")
					continue
				}
				// Tell the ProcessManager to send a SIGINT to the running engine
				s.procs.StopTaskByID(pl.TaskID)
			}
		}
	}
	log.Warn().Msg("Beaconing has stopped")
}
