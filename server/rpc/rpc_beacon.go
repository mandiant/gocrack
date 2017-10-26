package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"

	"github.com/gin-gonic/gin"
)

// BeaconPayloadType describes the type of payload being sent to the worker
type BeaconPayloadType uint8

const (
	// BeaconNewTask indicates the BeaconResponse.Payload is a "NewTask"
	BeaconNewTask BeaconPayloadType = 1 << iota
	// BeaconChangeTaskStatus indicates that the payload should be parsed as a task status change
	BeaconChangeTaskStatus
)

// NewTask describes the payload sent to the worker when a task should be executed
type NewTask struct {
	ID       string
	Engine   storage.WorkerCrackEngine
	Priority storage.WorkerPriority
	Devices  storage.CLDevices
}

// ChangeTaskStatus describes the payload sent to the worker
type ChangeTaskStatus struct {
	TaskID    string
	NewStatus storage.TaskStatus
}

// BeaconRequest describes the payload sent by the worker to the server
type BeaconRequest shared.Beacon

// PayloadItem is an item that contains data for a BeaconResponse where the data is unmarshalled based on it's Type
type PayloadItem struct {
	Type BeaconPayloadType
	Data json.RawMessage
}

// BeaconResponse describes the payload sent to the worker on a beacon
type BeaconResponse struct {
	Payloads   []PayloadItem
	ServerTime time.Time
}

// GetDevicesInUse returns a list of CLDevices inuse from a DeviceMap object
func GetDevicesInUse(d shared.DeviceMap) storage.CLDevices {
	var inuse storage.CLDevices

	for id, dev := range d {
		if dev.IsBusy {
			inuse = append(inuse, int(id))
		}
	}

	return inuse
}

func (s *RPCServer) workerBeacon(c *gin.Context) *RPCError {
	var req BeaconRequest

	if err := c.BindJSON(&req); err != nil {
		return &RPCError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
		}
	}

	s.wmgr.HostCheckingIn(shared.Beacon(req))
	host := s.wmgr.GetCurrentHostRecord(req.Hostname)
	if host == nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("worker manager had no record for %s?", req.Hostname),
		}
	}

	actions, err := s.stor.GetPendingTasks(storage.GetPendingTasksRequest{
		Hostname:        req.Hostname,
		DevicesInUse:    GetDevicesInUse(req.Devices),
		RunningTasks:    host.GetRunningTaskIDs(),
		CheckForNewTask: req.RequestNewTask,
	})

	if err != nil {
		return &RPCError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if actions == nil {
		c.JSON(http.StatusOK, BeaconResponse{
			ServerTime: time.Now().UTC(),
		})
		return nil
	}

	resp := BeaconResponse{
		ServerTime: time.Now().UTC(),
		Payloads:   make([]PayloadItem, len(actions)),
	}

	for i, action := range actions {
		switch action.Type {
		case storage.PendingTaskNewRequest:
			nextTask := action.Payload.(*storage.Task)
			ntreq := &NewTask{
				ID:       nextTask.TaskID,
				Engine:   nextTask.Engine,
				Priority: nextTask.Priority,
			}

			if nextTask.AssignedToDevices != nil && nextTask.AssignedToHost != "" {
				ntreq.Devices = *nextTask.AssignedToDevices
			}

			bytez, err := json.Marshal(ntreq)
			if err != nil {
				return &RPCError{
					StatusCode: http.StatusInternalServerError,
					Err:        err,
				}
			}

			resp.Payloads[i] = PayloadItem{
				Type: BeaconNewTask,
				Data: bytez,
			}
		case storage.PendingTaskStatusChange:
			newStatus := action.Payload.(storage.PendingTaskStatusChangeItem)
			bytez, err := json.Marshal(ChangeTaskStatus(newStatus))
			if err != nil {
				return &RPCError{
					StatusCode: http.StatusInternalServerError,
					Err:        err,
				}
			}

			resp.Payloads[i] = PayloadItem{
				Type: BeaconChangeTaskStatus,
				Data: bytez,
			}
		}
	}

	c.JSON(http.StatusOK, &resp)
	return nil
}
