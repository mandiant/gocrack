package hashcat

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/fireeye/gocat/v6/restoreutil"
	"github.com/fireeye/gocrack/opencl"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"
)

// ErrMalformedCLDevicesArg indicates that the "--opencl-devices" arg is not in the expected format
var ErrMalformedCLDevicesArg = errors.New("--opencl-devices argument is malformed")

func intSliceEq(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ModifyRestoreFileDevices will locate the arg "--opencl-devices" in the restore file and compare the devices against the input parameter of this function.
// If the devices are different, the argument in restoreutil.RestoreData will be modified.
func ModifyRestoreFileDevices(rd *restoreutil.RestoreData, devicesJobWillBeUsing storage.CLDevices, systemDevices shared.DeviceMap) (bool, error) {
	var tmp []int
	var clDevicesIndex = -1
	var clDeviceTypesIndex = -1

	for i, arg := range rd.Args {
		if strings.HasPrefix(arg, "--opencl-devices") {
			clDevicesIndex = i
		}

		if strings.HasPrefix(arg, "--opencl-device-types") {
			clDeviceTypesIndex = i
		}
	}

	if clDevicesIndex != -1 {
		arg := rd.Args[clDevicesIndex]
		devicesIdx := strings.Index(arg, "=")
		if devicesIdx == -1 {
			return false, ErrMalformedCLDevicesArg
		}

		// iterate over the devices that currently exist in the
		// restore file and convert the integers from strings
		for _, device := range strings.Split(arg[devicesIdx+1:], ",") {
			did, err := strconv.Atoi(device)
			if err != nil {
				return false, err
			}
			tmp = append(tmp, did)
		}

		// Sort the lists before doing an equality check
		sort.Ints(tmp)
		sort.Ints(devicesJobWillBeUsing)

		// if the slices are not equal, update the restore file
		if !intSliceEq(tmp, devicesJobWillBeUsing) {
			var devicesStr []string
			var devTypes []int
			var dGPUAdded, dCPUAdded bool

			for _, newid := range devicesJobWillBeUsing {
				devicesStr = append(devicesStr, fmt.Sprintf("%d", newid))
				// check and make sure it's actually a device on this system as well as get the type of device it is
				if device, deviceIsReal := systemDevices[newid]; deviceIsReal {
					switch device.Type {
					case opencl.DeviceTypeCPU:
						if !dCPUAdded {
							devTypes = append(devTypes, 1)
							dCPUAdded = true
						}
					case opencl.DeviceTypeGPU:
						if !dGPUAdded {
							devTypes = append(devTypes, 2)
							dGPUAdded = true
						}
					}
				}
			}

			rd.Args[clDevicesIndex] = fmt.Sprintf("--opencl-devices=%s", strings.Join(devicesStr, ","))
			if clDeviceTypesIndex != -1 {
				rd.Args[clDeviceTypesIndex] = fmt.Sprintf("--opencl-device-types=%s", shared.IntSliceToString(devTypes))
			}
			return true, nil
		}
	}
	return false, nil
}
