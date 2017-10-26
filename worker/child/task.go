package child

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fireeye/gocrack/server/rpc"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"
	"github.com/fireeye/gocrack/worker"
	"github.com/fireeye/gocrack/worker/engines"
	"github.com/fireeye/gocrack/worker/engines/hashcat"

	"github.com/rs/zerolog/log"
)

type Task struct {
	// unexported fields below
	taskid  string
	devices storage.CLDevices
	done    chan bool
	wg      *sync.WaitGroup
	cfg     *worker.Config
	c       rpc.GoCrackRPC
	impl    engines.EngineImpl
}

func NewTask(taskid string, devices []int, cfg *worker.Config, c rpc.GoCrackRPC) *Task {
	return &Task{
		taskid:  taskid,
		devices: storage.CLDevices(devices),
		done:    make(chan bool, 1),
		c:       c,
		cfg:     cfg,
		wg:      &sync.WaitGroup{},
	}
}

// DownloadFile grabs a file from the server and stores it in the appropriate folder
func (t *Task) DownloadFile(fileid string, filetype rpc.FileType) (string, error) {
	var fp string

	switch filetype {
	case rpc.FileTypeEngine:
		fp = filepath.Join(t.cfg.SaveEngineFilePath, fileid)
	case rpc.FileTypeTask:
		fp = filepath.Join(t.cfg.SaveTaskFilePath, fileid)
	default:
		return "", errors.New("unknown file type")
	}

	fpLock, err := acquireLock(fp)
	if err != nil {
		return "", err
	}
	defer func() {
		fpLock.Unlock()
		log.Debug().
			Str("file_id", fileid).
			Msg("Released lockfile")
	}()

	filebody, serverHash, err := t.c.GetFile(rpc.TaskFileGetRequest{
		FileID: fileid,
		Type:   filetype,
	})
	if err != nil {
		return "", err
	}
	defer filebody.Close()

	// Let's check and see if our files are the same...
	if _, err := os.Stat(fp); err == nil {
		ourHash, err := checkFileHash(fp)
		if err != nil {
			return "", err
		}

		if ourHash == serverHash {
			log.Debug().
				Str("hash", serverHash).
				Str("file_id", fileid).
				Uint8("type", uint8(filetype)).
				Msg("Skipping file download as content is the same")
			return fp, nil
		}
	}

	log.Debug().
		Str("hash", serverHash).
		Str("file_id", fileid).
		Uint8("type", uint8(filetype)).
		Msg("Downloading file via RPC")

	fd, err := os.Create(fp)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	if _, err := io.Copy(fd, filebody); err != nil {
		return "", err
	}

	fd.Sync()

	log.Debug().Str("hash", serverHash).
		Str("file_id", fileid).
		Uint8("type", uint8(filetype)).
		Msg("Downloaded file via RPC")
	return fp, nil
}

func (t *Task) sendPeriodicStatus(engine storage.WorkerCrackEngine) {
	tickEvery := time.NewTicker(t.cfg.Intervals.JobStatus.Duration)
	defer func() {
		tickEvery.Stop()
		t.wg.Done()
	}()

	for {
		select {
		case <-t.done:
			return
		case <-tickEvery.C:
			status := t.impl.GetStatus()
			if status == nil {
				continue
			}

			if err := t.c.SendTaskStatus(rpc.TaskStatusUpdate{
				TaskID:  t.taskid,
				Payload: status,
				Engine:  engine,
			}); err != nil {
				log.Error().Err(err).Msg("Failed to send task status update to server")
			}
		}
	}
}

// Start the task and block until the engine is done
func (t *Task) Start() error {
	resp, err := t.c.GetTask(rpc.RequestTaskPayload{
		TaskID: t.taskid,
	})
	if err != nil {
		return err
	}

	if err = t.c.ChangeTaskStatus(rpc.ChangeTaskStatusRequest{
		TaskID:    t.taskid,
		NewStatus: storage.TaskStatusRunning,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to change tasks status to Running...continuing with job")
	}

	taskFilePath, err := t.DownloadFile(resp.FileID, rpc.FileTypeTask)
	if err != nil {
		return err
	}
	defer os.Remove(taskFilePath)

	log.Debug().Str("task_file", taskFilePath).Msg("Downloaded Task File to temporary directory")

	switch resp.Engine {
	case storage.WorkerHashcatEngine:
		var hashcatOpts shared.HashcatUserOptions
		if err := json.Unmarshal(resp.EnginePayload, &hashcatOpts); err != nil {
			return err
		}
		hc := &hashcat.HashcatEngine{
			TaskID:            t.taskid,
			SessionPath:       t.cfg.Hashcat.SessionPath,
			HashcatSharedPath: t.cfg.Hashcat.SharedPath,
			TaskFilePath:      taskFilePath,
			Options:           hashcatOpts,
			CLDevices:         t.devices,
			Upstream:          t.c,
		}

		if hashcatOpts.DictionaryFile != nil {
			dictFilePath, err := t.DownloadFile(*hashcatOpts.DictionaryFile, rpc.FileTypeEngine)
			if err != nil {
				return err
			}
			hc.DictionaryFile = dictFilePath
		}

		if hashcatOpts.Masks != nil {
			masksFilePath, err := t.DownloadFile(*hashcatOpts.Masks, rpc.FileTypeEngine)
			if err != nil {
				return err
			}
			hc.MasksFile = masksFilePath
		}

		if hashcatOpts.ManglingRuleFile != nil {
			manglingRuleFile, err := t.DownloadFile(*hashcatOpts.ManglingRuleFile, rpc.FileTypeEngine)
			if err != nil {
				return err
			}
			hc.RulesFile = manglingRuleFile
		}
		t.impl = hc
	default:
		return fmt.Errorf("unknown engine %d", resp.Engine)
	}

	if err = t.impl.Initialize(); err != nil {
		return err
	}

	t.wg.Add(2) // periodic status goroutine + 1 indicating hashcat is running
	go t.sendPeriodicStatus(resp.Engine)
	defer func() {
		t.impl.Cleanup()
		close(t.done) // signal the stop to the goroutines
		t.wg.Done()   // decremnt the WG by 1 as we added 1 for our running worker
	}()

	return t.impl.Start()
}

// Stop the task
func (t *Task) Stop() {
	log.Debug().Msg("Stop requested")

	// Stop the engine
	if t.impl != nil {
		if err := t.impl.Stop(); err != nil {
			log.Error().Err(err).Msg("Error stopping engine")
		}
	}

StopSelect:
	select {
	case <-t.done:
		break StopSelect
	case <-time.After(t.cfg.Intervals.TerminationDelay.Duration):
		t.c.ChangeTaskStatus(rpc.ChangeTaskStatusRequest{
			TaskID:    t.taskid,
			NewStatus: storage.TaskStatusStopped,
		})
		log.Fatal().Msg("Task took too long to exit, child is exiting abruptly")
		return
	}

	t.wg.Wait()
}
