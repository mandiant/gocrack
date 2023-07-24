package rpc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mandiant/gocrack/server/storage"

	"github.com/gin-gonic/gin"
)

type FileType uint8

const (
	FileTypeTask FileType = 1 << iota
	FileTypeEngine
)

type ChangeTaskStatusRequest struct {
	TaskID    string
	NewStatus storage.TaskStatus
	Error     *string
}

type RequestTaskPayload struct {
	TaskID string
}

type NewTaskPayloadResponse struct {
	TaskID        string
	FileID        string
	Engine        storage.WorkerCrackEngine
	Priority      storage.WorkerPriority
	EnginePayload json.RawMessage
	TaskDuration  int
}

type TaskFileGetRequest struct {
	FileID string
	Type   FileType
}

type CrackedPasswordRequest struct {
	TaskID    string
	Hash      string
	Value     string
	CrackedAt time.Time
}

type TaskStatusUpdate struct {
	Engine  storage.WorkerCrackEngine
	TaskID  string
	Final   bool
	Payload interface{}
}

type TaskCheckpointSaveRequest struct {
	TaskID string
	Data   []byte
}

func (s *RPCServer) changeTaskStatus(c *gin.Context) *RPCError {
	var req ChangeTaskStatusRequest

	if err := c.BindJSON(&req); err != nil {
		return &RPCError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}

	if err := s.stor.ChangeTaskStatus(req.TaskID, req.NewStatus, req.Error); err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if err := s.wmgr.BroadcastTaskStatusChange(req.TaskID, req.NewStatus); err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}

func (s *RPCServer) getTaskPayload(c *gin.Context) *RPCError {
	var req RequestTaskPayload
	if err := c.BindJSON(&req); err != nil {
		return &RPCError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}

	task, err := s.stor.GetTaskByID(req.TaskID)
	if err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	engineBytes, err := json.Marshal(task.EnginePayload)
	if err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	c.JSON(http.StatusOK, &NewTaskPayloadResponse{
		TaskID:        task.TaskID,
		FileID:        task.FileID,
		Engine:        task.Engine,
		EnginePayload: engineBytes,
		Priority:      task.Priority,
		TaskDuration:  task.TaskDuration,
	})

	return nil
}

func (s *RPCServer) getTaskFile(c *gin.Context) *RPCError {
	var (
		req            TaskFileGetRequest
		locationOnDisk string
	)

	if err := c.BindJSON(&req); err != nil {
		return &RPCError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}

	switch req.Type {
	case FileTypeTask:
		tf, err := s.stor.GetTaskFileByID(req.FileID)
		if err != nil {
			return &RPCError{
				StatusCode: http.StatusBadRequest,
				Err:        err,
			}
		}
		c.Header("X-FileHash-SHA1", tf.SHA1Hash)
		locationOnDisk = tf.SavedAt
	case FileTypeEngine:
		sf, err := s.stor.GetEngineFileByID(req.FileID)
		if err != nil {
			return &RPCError{
				StatusCode: http.StatusBadRequest,
				Err:        err,
			}
		}
		c.Header("X-FileHash-SHA1", sf.SHA1Hash)
		locationOnDisk = sf.SavedAt
	default:
		return &RPCError{
			StatusCode: http.StatusBadRequest,
			Err:        fmt.Errorf("unexpected type of %d for GetFile", req.Type),
		}
	}

	fd, err := os.Open(locationOnDisk)
	if err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}
	defer fd.Close()

	if _, err := io.Copy(c.Writer, fd); err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	return nil
}

func (s *RPCServer) saveCrackedPassword(c *gin.Context) *RPCError {
	var req CrackedPasswordRequest

	if err := c.BindJSON(&req); err != nil {
		return &RPCError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}

	crackedCounter.Inc()
	if err := s.stor.SaveCrackedHash(req.TaskID, req.Hash, req.Value, req.CrackedAt); err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if err := s.wmgr.BroadcastCrackedPassword(req.TaskID, req.Hash, req.Value, req.CrackedAt); err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}

func (s *RPCServer) taskStatusUpdate(c *gin.Context) *RPCError {
	var req TaskStatusUpdate
	if err := c.BindJSON(&req); err != nil {
		return &RPCError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}

	switch req.Final {
	case true:
		if err := s.wmgr.BroadcastFinalStatus(req.TaskID, req.Payload); err != nil {
			return &RPCError{
				StatusCode: http.StatusInternalServerError,
				Err:        err,
			}
		}
	default:
		if err := s.wmgr.BroadcastEngineStatusUpdate(req.TaskID, req.Payload); err != nil {
			return &RPCError{
				StatusCode: http.StatusInternalServerError,
				Err:        err,
			}
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}

func (s *RPCServer) taskRestoreFile(c *gin.Context) *RPCError {
	var req TaskCheckpointSaveRequest
	if err := c.BindJSON(&req); err != nil {
		return &RPCError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}

	if err := s.stor.SaveTaskCheckpoint(storage.CheckpointFile(req)); err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}

func (s *RPCServer) getCheckpointFile(c *gin.Context) *RPCError {
	var taskID = c.Param("taskid")
	var b []byte
	var err error

	if b, err = s.stor.GetTaskCheckpoint(taskID); err != nil {
		// No checkpoint exists for this task...
		if err == storage.ErrNotFound {
			c.Status(http.StatusNoContent)
			return nil
		}
		goto Error
	}

	if _, err := c.Writer.Write(b); err != nil {
		goto Error
	}

	return nil
Error:
	return &RPCError{
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}
