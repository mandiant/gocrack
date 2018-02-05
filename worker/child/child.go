package child

import (
	"fmt"
	"os"
	"github.com/fireeye/gocrack/server/rpc"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/worker"
	"time"
	"github.com/rs/zerolog/log"
)

// Worker is a version of the worker spawned by the parent and runs the actual engine task
type Worker struct {
	// Unexported fields below
	taskid  string
	devices []int
	cfg     *worker.Config
	rc      rpc.GoCrackRPC
	t       *Task
}

// New instantiates the child worker process
func New(cfg *worker.Config, taskid string, devices []int) *Worker {
	return &Worker{
		taskid:  taskid,
		devices: devices,
		cfg:     cfg,
	}
}

// Start the processing of the task
func (s *Worker) Start() error {
	client, err := worker.InitRPCChannel(*s.cfg)
	if err != nil {
		return err
	}
	s.rc = client

	defer func() {
		if r := recover(); r != nil {
			if client != nil {
				// we dont really care about the error here...
				hostname, _ := os.Hostname()
				errStr := fmt.Sprintf("A panic occurred. Check logs on %s for more details", hostname)
				client.ChangeTaskStatus(rpc.ChangeTaskStatusRequest{
					TaskID:    s.taskid,
					NewStatus: storage.TaskStatusError,
					Error:     &errStr,
				})
			}
			log.Error().Str("task_id", s.taskid).Msg("A critical error occurred while running task (panic)")
		}
	}()

	s.t = NewTask(s.taskid, s.devices, s.cfg, s.rc) //Get the task in order to collect the task duration
	resp, err:= s.t.c.GetTask(rpc.RequestTaskPayload{
                TaskID: s.t.taskid,
        })
	if resp.TaskDuration !=0{ //If the task duration is 0 (not set), we don't run the timer
		timer := time.NewTimer(time.Second * time.Duration(resp.TaskDuration))
		go func() {
			<-timer.C
			log.Warn().Msg("Timer expired, stopping task")
			s.t.Stop()
			}()
	}

	if err := s.t.Start(); err != nil {
		log.Error().Err(err).Str("task_id", s.taskid).Msg("An error occurred while processing a task")
		errptr := err.Error()
		if rpcerr := client.ChangeTaskStatus(rpc.ChangeTaskStatusRequest{
			TaskID:    s.taskid,
			NewStatus: storage.TaskStatusError,
			Error:     &errptr,
		}); rpcerr != nil {
			log.Error().Err(rpcerr).Msg("Failed to change tasks status to error")
		}
	}
	return nil
}

// Stop the task
func (s *Worker) Stop() error {
	if s.t != nil {
		s.t.Stop()
	}
	return nil
}
