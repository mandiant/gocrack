package shared

import (
	"strconv"
	"strings"
	"time"

	"github.com/mandiant/gocrack/opencl"
)

// Device describes an OpenCL device on a worker
type Device struct {
	ID     int
	Name   string
	Type   opencl.DeviceType
	IsBusy bool
}

// DeviceMap stores OpenCL Device information by the device's unique ID.
// The methods attached to this are not safe for concurrent use
type DeviceMap map[int]*Device

// PickFreeDevices returns a list of device ID's that are not marked as Busy
func (s DeviceMap) PickFreeDevices(deviceType opencl.DeviceType, maxNumber int) []int {
	devicesToUse := []int{}

	for id, dev := range s {
		if dev.IsBusy {
			continue
		}

		if len(devicesToUse) == maxNumber {
			break
		}

		if deviceType == dev.Type {
			devicesToUse = append(devicesToUse, id)
		}
	}
	return devicesToUse
}

// MarkAsBusy marks devices as busy
func (s DeviceMap) MarkAsBusy(devices []int) {
	for _, deviceID := range devices {
		if device, ok := s[deviceID]; ok {
			device.IsBusy = true
		}
	}
}

// MarkAsFree marks devices as free
func (s DeviceMap) MarkAsFree(devices []int) {
	for _, deviceID := range devices {
		if device, ok := s[deviceID]; ok {
			device.IsBusy = false
		}
	}
}

// HasFreeDevices indicates if we have at least one device that can accept a task
func (s DeviceMap) HasFreeDevices() bool {
	for _, device := range s {
		if !device.IsBusy {
			return true
		}
	}
	return false
}

// TaskProcess is information returned by a worker about a given child process it's managing
type TaskProcess struct {
	Pid          int
	MemoryUsage  uint64
	RunningFor   time.Duration
	UsingDevices []int
}

// EngineVersion is a map of the engines contained within the worker along with their engine version
type EngineVersion map[string]string // map[engine_name]engine_version

// Beacon describes the payload sent by a worker
type Beacon struct {
	WorkerVersion  string
	Hostname       string
	RequestNewTask bool
	Devices        DeviceMap
	Processes      map[string]TaskProcess // map[taskid]TaskProcess
	Engines        EngineVersion
}

// GetIntPtr returns the address of i
func GetIntPtr(i int) *int {
	return &i
}

// GetStrPtr returns the address of s
func GetStrPtr(s string) *string {
	return &s
}

// GetBoolPtr returns the address of b
func GetBoolPtr(b bool) *bool {
	return &b
}

// IntSliceToString converts a list of integers to a comma separated list
func IntSliceToString(ints []int) string {
	tmp := make([]string, len(ints))
	for i, v := range ints {
		tmp[i] = strconv.Itoa(v)
	}

	return strings.Join(tmp, ",")
}
