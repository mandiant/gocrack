package opencl

// #include "include.h"
import "C"
import "unsafe"

const maxDeviceCount = 64

type Device struct {
	id C.cl_device_id
}

// GetDevices returns a list of available devices on a platform. 'platform' refers
// to the platform returned by GetPlatforms or can be nil. If platform
// is nil, the behavior is implementation-defined.
func GetDevices(platform *Platform, deviceType DeviceType) ([]*Device, error) {
	var deviceIds [maxDeviceCount]C.cl_device_id
	var numDevices C.cl_uint
	var platformID C.cl_platform_id

	if platform != nil {
		platformID = platform.id
	}

	if err := C.clGetDeviceIDs(platformID, C.cl_device_type(deviceType), C.cl_uint(maxDeviceCount), &deviceIds[0], &numDevices); err != C.CL_SUCCESS {
		return nil, toError(err)
	}

	if numDevices > maxDeviceCount {
		numDevices = maxDeviceCount
	}

	devices := make([]*Device, numDevices)
	for i := 0; i < int(numDevices); i++ {
		devices[i] = &Device{id: deviceIds[i]}
	}
	return devices, nil
}

func (s *Device) getInfoString(param C.cl_device_info) (string, error) {
	var strN C.size_t
	// Get the size of the buffer to allocate
	if retval := C.clGetDeviceInfo(s.id, param, 0, nil, &strN); retval != C.CL_SUCCESS {
		return "", toError(retval)
	}

	b := make([]byte, strN)
	if retval := C.clGetDeviceInfo(s.id, param, strN, unsafe.Pointer(&b[0]), &strN); retval != C.CL_SUCCESS {
		return "", toError(retval)
	}
	return string(b[:len(b)-1]), nil
}

// Name returns the name of the OpenCL device
func (s *Device) Name() (string, error) {
	return s.getInfoString(C.CL_DEVICE_NAME)
}

// DriverVersion returns the OpenCL driver revision
func (s *Device) DriverVersion() (string, error) {
	return s.getInfoString(C.CL_DRIVER_VERSION)
}

// OpenCLVersion returns the version of OpenCL supported by this device
func (s *Device) OpenCLVersion() (string, error) {
	return s.getInfoString(C.CL_DEVICE_OPENCL_C_VERSION)
}

// Type returns the type of OpenCL device this is (CPU, GPU, etc.)
func (s *Device) Type() (dt DeviceType, err error) {
	var deviceType C.cl_device_type

	if retval := C.clGetDeviceInfo(s.id, C.CL_DEVICE_TYPE, C.size_t(unsafe.Sizeof(deviceType)), unsafe.Pointer(&deviceType), nil); retval != C.CL_SUCCESS {
		err = toError(retval)
		return
	}
	return DeviceType(deviceType), nil
}
