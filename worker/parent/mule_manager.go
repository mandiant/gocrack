package parent

// mule_manager.go contains code that the "parent" process uses to communicate & manage the "child" processes (hashcat, JtR, etc.)

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/fireeye/gocrack/shared"

	"github.com/shirou/gopsutil/process"
)

/*
Notes:

There are several ways the child process can fail...

1. The child process is terminated either by the user or the system
	* the monitor goroutine should be monitoring for the process to exit and call the cleanup function which signals all the goroutines to exit
2. The parent gets SIGINT
	* all the subprocesses should get the same signal and attept to gracefully exit (and abruptly stop their task)
*/

// PIDUNKN stands for PID Unknown and indicates the process has not started yet (or has exited)
const PIDUNKN = -1

type rpmDevices struct {
	devices []int
	*RemoteProcessManager
}

// ProcessesByTask is a mapping of current running processes stored by their taskid
type ProcessesByTask struct {
	// unexported fields below
	l    *sync.RWMutex
	data map[string]rpmDevices
}

// NewProcessesByTask creates a tracker for task processes
func NewProcessesByTask() ProcessesByTask {
	return ProcessesByTask{
		l:    &sync.RWMutex{},
		data: make(map[string]rpmDevices),
	}
}

// GetBeaconInfo returns information about the current running processes on this worker
func (s *ProcessesByTask) GetBeaconInfo() map[string]shared.TaskProcess {
	out := make(map[string]shared.TaskProcess)
	now := time.Now().UTC()

	s.l.RLock()
	defer s.l.RUnlock()

	for taskid, proc := range s.data {
		out[taskid] = shared.TaskProcess{
			Pid:          proc.GetRunningPID(),
			MemoryUsage:  proc.GetMemoryRSS(),
			RunningFor:   now.Sub(proc.started),
			UsingDevices: proc.devices,
		}
	}
	return out
}

// RegisterProcess registers the child process/worker which is performing the cracking session. This should be called before the process starts
func (s *ProcessesByTask) RegisterProcess(rpm *RemoteProcessManager, taskid string, devices []int) {
	s.l.Lock()
	defer s.l.Unlock()

	s.data[taskid] = rpmDevices{devices, rpm}
}

// UnregisterProcess removes the child process/worker from the manager. This should be called when the process exits
func (s *ProcessesByTask) UnregisterProcess(rpm *RemoteProcessManager, taskid string) {
	s.l.Lock()
	defer s.l.Unlock()

	if _, ok := s.data[taskid]; ok {
		delete(s.data, taskid)
	}
}

// StopTaskByID sends a SIGTERM to a running child process being monitored by the parent. If the task does not exist, nothing happens.
func (s *ProcessesByTask) StopTaskByID(taskid string) {
	s.l.Lock()
	defer s.l.Unlock()

	if _, ok := s.data[taskid]; ok {
		s.data[taskid].StopTerm()
	}
}

// RemoteOutput is a channel on which messages from child processes pipes are sent to
type RemoteOutput chan PipeResponse

// RemoteProcessManager is a child process of the worker which does all the heavy lifting
type RemoteProcessManager struct {
	// unexported fields below
	pid           int
	stopRequested bool
	p             *exec.Cmd
	wg            *sync.WaitGroup
	psu           *process.Process
	ch            RemoteOutput // this is killed within worker/parent
	rc            int
	started       time.Time
}

// PipeResponse is the data returned from the pipe
type PipeResponse struct {
	Data     string
	Pid      int
	IsStdErr bool
}

// NewRemoteProcess creates a process wrapper that allows stderr/out to be recorded as well as administratively manage the process
func NewRemoteProcess(ch RemoteOutput) *RemoteProcessManager {
	return &RemoteProcessManager{
		wg:  &sync.WaitGroup{},
		pid: PIDUNKN,
		ch:  ch,
	}
}

// piperdr is a goroutine that reads either Stdout or Sterr from the child process
func (s *RemoteProcessManager) piperdr(pipe io.ReadCloser, isErrPipe bool) {
	defer s.wg.Done()
	defer pipe.Close()

PipeLoop:
	for {
		scanner := bufio.NewScanner(pipe)
		// This will stop when we EOF
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case s.ch <- PipeResponse{
				Data:     line,
				IsStdErr: isErrPipe,
				Pid:      s.pid,
			}:
				// Sent message
			default:
				// Dropped
			}
		}
		break PipeLoop
	}
}

// GetMemoryRSS returns the resident size of the remote process. If we are unable to retrieve it, we return 0
func (s *RemoteProcessManager) GetMemoryRSS() uint64 {
	if s.psu == nil {
		return 0
	}

	stat, err := s.psu.MemoryInfo()
	if err != nil {
		return 0
	}

	return stat.RSS
}

// manager watches the process and will call necessary cleanup functions after the process has exited
func (s *RemoteProcessManager) manager() {
	defer s.wg.Done()

	s.pid = s.p.Process.Pid
	done := make(chan error)
	defer close(done)

	// if we error here, we simply cant get memory usage
	psu, _ := process.NewProcess(int32(s.pid))
	if psu != nil {
		s.psu = psu
	}

	// unmanaged goroutine.... this simply will notify the done channel when the process exits
	go func() { done <- s.p.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			// Attempt and get the return code...
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					s.rc = status.ExitStatus()
				}
			}
		}
		return
	}
}

// Start the mule process up and wait for it to exit
func (s *RemoteProcessManager) Start(process string, args ...string) (err error) {
	cmd := exec.Command(process, args...)
	s.p = cmd

	if s.ch != nil {
		stpipe, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		sterrPipe, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		s.wg.Add(2)
		go s.piperdr(stpipe, false)
		go s.piperdr(sterrPipe, true)
	}

	// Start the process asynchronously
	if err := cmd.Start(); err != nil {
		return err
	}
	s.started = time.Now().UTC()

	s.wg.Add(1)
	go s.manager()

	return nil
}

// GetRunningPID returns the current PID of the process. If the process has not started yet, the value is -1.
func (s *RemoteProcessManager) GetRunningPID() int {
	return s.pid
}

// GetReturnCode returns the return code of the process after it exits
func (s *RemoteProcessManager) GetReturnCode() int {
	return s.rc
}

// Wait blocks until the process has completely exited
func (s *RemoteProcessManager) Wait() {
	s.wg.Wait()
}

// StopTerm stops the mule process by sending it a SIGERM
func (s *RemoteProcessManager) StopTerm() {
	// Only send the signal if the process isnt nill and the stop hasn't already been processed
	if s.p != nil && !s.stopRequested {
		s.stopRequested = true
		s.p.Process.Signal(syscall.SIGTERM)
	}
}

// StopKill stops the mule process by sending it a SIGKILL
func (s *RemoteProcessManager) StopKill() {
	if s.p != nil {
		s.p.Process.Kill()
	}
}
