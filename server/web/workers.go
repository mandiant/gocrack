package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mandiant/gocrack/opencl"

	"github.com/gin-gonic/gin"
)

// WorkerDeviceType is the type of hardware device this is. GPU, CPU, FPGA, etc.
type WorkerDeviceType struct {
	opencl.DeviceType
}

// MarshalJSON returns a JSON string of the Device Type
func (s *WorkerDeviceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// WorkerDevice contains information about a device attached to a worker.
// Note: this should mirror shared.Device
type WorkerDevice struct {
	ID     int              `json:"id"`
	Name   string           `json:"name"`
	Type   WorkerDeviceType `json:"type"`
	IsBusy bool             `json:"-"`
}

// WorkerProcess describes a cracking process and basic metadata about it
type WorkerProcess struct {
	TaskID       string `json:"task_id"`
	PID          int    `json:"pid"`
	MemoryUsage  uint64 `json:"memory_usage"`
	RunningFor   string `json:"running_for"`
	UsingDevices []int  `json:"using_devices"`
	CreatedBy    string `json:"created_by"`
}

// WorkerItem describes a connected worker to the system
type WorkerItem struct {
	Hostname    string          `json:"hostname"`
	LastCheckin time.Time       `json:"last_seen"`
	Devices     []WorkerDevice  `json:"devices"`
	Processes   []WorkerProcess `json:"running_tasks"`
}

// WorkerResponse contains a list of workers along with information about them
type WorkerResponse []WorkerItem

func (s *Server) webGetActiveWorkers(c *gin.Context) *WebAPIError {
	resp := WorkerResponse{}

	workers := s.wmgr.GetCurrentWorkers()
	for hostname, worker := range workers {
		item := WorkerItem{
			Hostname:    hostname,
			LastCheckin: worker.LastCheckin,
			Devices:     make([]WorkerDevice, len(worker.LastBeacon.Devices)),
			Processes:   make([]WorkerProcess, 0),
		}

		for i, device := range worker.LastBeacon.Devices {
			item.Devices[i-1] = WorkerDevice{
				ID:     device.ID,
				Name:   device.Name,
				IsBusy: device.IsBusy,
				Type:   WorkerDeviceType{device.Type},
			}
		}

		var createdby string
		for taskid, proc := range worker.LastBeacon.Processes {
			// XXX(cschmitt): Do we want to cache this?
			task, err := s.stor.GetTaskByID(taskid)
			if err != nil {
				createdby = "Unknown"
			} else {
				createdby = task.CreatedBy
			}
			item.Processes = append(item.Processes, WorkerProcess{
				PID:          proc.Pid,
				MemoryUsage:  proc.MemoryUsage,
				TaskID:       taskid,
				RunningFor:   proc.RunningFor.String(),
				UsingDevices: proc.UsingDevices,
				CreatedBy:    createdby,
			})
		}
		resp = append(resp, item)
	}

	c.JSON(http.StatusOK, resp)

	return nil
}
