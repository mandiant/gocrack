package worker

import (
	"github.com/fireeye/gocrack/opencl"
	"github.com/fireeye/gocrack/shared"
)

// GetAvailableDevices returns all available OpenCL devices with the same ID of hashcat
// Note: had to roll our own due to how OpenCL was integrated into hashcat. Could not use
// the hashcat OpenCL apis
func GetAvailableDevices() (shared.DeviceMap, error) {
	platforms, err := opencl.GetPlatforms()
	if err != nil {
		return nil, err
	}

	gDeviceID := 0
	devs := make(map[int]*shared.Device, 0)

	for _, platform := range platforms {
		devices, err := platform.GetDevices(opencl.DeviceTypeAll)
		if err != nil {
			if err == opencl.ErrDeviceNotFound {
				continue
			}
			return nil, err
		}

		for _, device := range devices {
			// this is to ensure our OpenCL device code & hashcat's has the same device ID's.
			gDeviceID++
			dev := &shared.Device{ID: gDeviceID}

			if dev.Name, err = device.Name(); err != nil {
				return nil, err
			}

			if dev.Type, err = device.Type(); err != nil {
				return nil, err
			}
			devs[gDeviceID] = dev
		}
	}

	if len(devs) == 0 {
		return nil, opencl.ErrDeviceNotFound
	}

	return devs, nil
}
