package client

import (
	"bytes"
	"fmt"
	"io"

	"github.com/mandiant/gocrack/server/rpc"
)

// ChangeTaskStatus instructs the server to change the status of a task
func (s *RPCClient) ChangeTaskStatus(request rpc.ChangeTaskStatusRequest) error {
	return s.performJSONCall("POST", "/rpc/v1/task/status_change", request, nil)
}

// GetTask retrieves the task payload
func (s *RPCClient) GetTask(request rpc.RequestTaskPayload) (*rpc.NewTaskPayloadResponse, error) {
	var resp rpc.NewTaskPayloadResponse
	if err := s.performJSONCall("POST", "/rpc/v1/task/payload", request, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetFile retrieves a shared or task file for a task
func (s *RPCClient) GetFile(request rpc.TaskFileGetRequest) (io.ReadCloser, string, error) {
	resp, err := s.performFileCall("POST", "/rpc/v1/file", request)
	if err != nil {
		return nil, "", err
	}

	return resp.File, resp.Hash, err
}

// SavedCrackedPassword instructs the server of a newly cracked password
func (s *RPCClient) SavedCrackedPassword(request rpc.CrackedPasswordRequest) error {
	return s.performJSONCall("POST", "/rpc/v1/task/cracked", request, nil)
}

// SendTaskStatus sends a real time (engine) status update to the server
func (s *RPCClient) SendTaskStatus(request rpc.TaskStatusUpdate) error {
	return s.performJSONCall("POST", "/rpc/v1/task/status", request, nil)
}

// SendCheckpointFile saves the restore point file on the server
func (s *RPCClient) SendCheckpointFile(request rpc.TaskCheckpointSaveRequest) error {
	return s.performJSONCall("POST", "/rpc/v1/task/checkpoint", request, nil)
}

// GetCheckpointFile retrieves the checkpoint file from the server
func (s *RPCClient) GetCheckpointFile(taskID string) ([]byte, error) {
	var buf bytes.Buffer
	path := fmt.Sprintf("/rpc/v1/task/checkpoint/%s", taskID)
	if err := s.performJSONCall("GET", path, nil, &buf); err != nil {
		return nil, err
	}

	if buf.Len() == 0 {
		return nil, rpc.ErrNoCheckpoint
	}
	return buf.Bytes(), nil
}
