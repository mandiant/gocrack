package opencl

// #include "include.h"
import "C"
import "unsafe"

const maxPlatforms = 32

type Platform struct {
	id C.cl_platform_id
}

// GetPlatforms returns a list of available OpenCL platforms
func GetPlatforms() ([]*Platform, error) {
	var platformIds [maxPlatforms]C.cl_platform_id
	var nPlatforms C.cl_uint

	if err := C.clGetPlatformIDs(C.cl_uint(maxPlatforms), &platformIds[0], &nPlatforms); err != C.CL_SUCCESS {
		return nil, toError(err)
	}

	platforms := make([]*Platform, nPlatforms)
	for i := 0; i < int(nPlatforms); i++ {
		platforms[i] = &Platform{id: platformIds[i]}
	}
	return platforms, nil
}

// GetDevices returns a list of devices for this particular platform
func (p *Platform) GetDevices(deviceType DeviceType) ([]*Device, error) {
	return GetDevices(p, deviceType)
}

func (p *Platform) getInfoString(param C.cl_platform_info) (string, error) {
	var strN C.size_t
	// Get the size of the buffer to allocate
	if retval := C.clGetPlatformInfo(p.id, param, 0, nil, &strN); retval != C.CL_SUCCESS {
		return "", toError(retval)
	}

	b := make([]byte, strN)
	if retval := C.clGetPlatformInfo(p.id, param, strN, unsafe.Pointer(&b[0]), &strN); retval != C.CL_SUCCESS {
		return "", toError(retval)
	}
	return string(b[:len(b)-1]), nil
}

// Name returns the name of the OpenCL platform
func (p *Platform) Name() (string, error) {
	return p.getInfoString(C.CL_PLATFORM_NAME)
}

// Vendor returns the name of the vendor for this OpenCL platform
func (p *Platform) Vendor() (string, error) {
	return p.getInfoString(C.CL_PLATFORM_VENDOR)
}

// Profile returns the profile for this OpenCL platform
func (p *Platform) Profile() (string, error) {
	return p.getInfoString(C.CL_PLATFORM_PROFILE)
}

// Version returns the max OpenCL version for this platform
func (p *Platform) Version() (string, error) {
	return p.getInfoString(C.CL_PLATFORM_VERSION)
}
