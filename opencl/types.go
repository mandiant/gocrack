package opencl

// #include "include.h"
import "C"
import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnknown is a result from an OpenCL function (e.g. CL_SUCCESS but null pointer)
var ErrUnknown = errors.New("cl: unknown error")

// DeviceType is the type of OpenCL Device (e.g. FPGA, CPU, GPU)
type DeviceType uint32

const (
	DeviceTypeCPU         DeviceType = C.CL_DEVICE_TYPE_CPU
	DeviceTypeGPU         DeviceType = C.CL_DEVICE_TYPE_GPU
	DeviceTypeAccelerator DeviceType = C.CL_DEVICE_TYPE_ACCELERATOR
	DeviceTypeDefault     DeviceType = C.CL_DEVICE_TYPE_DEFAULT
	DeviceTypeAll         DeviceType = C.CL_DEVICE_TYPE_ALL
)

func (dt DeviceType) String() string {
	var parts []string
	if dt&DeviceTypeCPU != 0 {
		parts = append(parts, "CPU")
	}
	if dt&DeviceTypeGPU != 0 {
		parts = append(parts, "GPU")
	}
	if dt&DeviceTypeAccelerator != 0 {
		parts = append(parts, "Accelerator")
	}
	if dt&DeviceTypeDefault != 0 {
		parts = append(parts, "Default")
	}
	if parts == nil {
		parts = append(parts, "None")
	}
	return strings.Join(parts, "|")
}

type ErrOther int

func (e ErrOther) Error() string {
	return fmt.Sprintf("cl: error %d", e)
}

var (
	ErrDeviceNotFound                     = errors.New("cl: Device Not Found")
	ErrDeviceNotAvailable                 = errors.New("cl: Device Not Available")
	ErrCompilerNotAvailable               = errors.New("cl: Compiler Not Available")
	ErrMemObjectAllocationFailure         = errors.New("cl: Mem Object Allocation Failure")
	ErrOutOfResources                     = errors.New("cl: Out Of Resources")
	ErrOutOfHostMemory                    = errors.New("cl: Out Of Host Memory")
	ErrProfilingInfoNotAvailable          = errors.New("cl: Profiling Info Not Available")
	ErrMemCopyOverlap                     = errors.New("cl: Mem Copy Overlap")
	ErrImageFormatMismatch                = errors.New("cl: Image Format Mismatch")
	ErrImageFormatNotSupported            = errors.New("cl: Image Format Not Supported")
	ErrBuildProgramFailure                = errors.New("cl: Build Program Failure")
	ErrMapFailure                         = errors.New("cl: Map Failure")
	ErrMisalignedSubBufferOffset          = errors.New("cl: Misaligned Sub Buffer Offset")
	ErrExecStatusErrorForEventsInWaitList = errors.New("cl: Exec Status Error For Events In Wait List")
	ErrCompileProgramFailure              = errors.New("cl: Compile Program Failure")
	ErrLinkerNotAvailable                 = errors.New("cl: Linker Not Available")
	ErrLinkProgramFailure                 = errors.New("cl: Link Program Failure")
	ErrDevicePartitionFailed              = errors.New("cl: Device Partition Failed")
	ErrKernelArgInfoNotAvailable          = errors.New("cl: Kernel Arg Info Not Available")
	ErrInvalidValue                       = errors.New("cl: Invalid Value")
	ErrInvalidDeviceType                  = errors.New("cl: Invalid Device Type")
	ErrInvalidPlatform                    = errors.New("cl: Invalid Platform")
	ErrInvalidDevice                      = errors.New("cl: Invalid Device")
	ErrInvalidContext                     = errors.New("cl: Invalid Context")
	ErrInvalidQueueProperties             = errors.New("cl: Invalid Queue Properties")
	ErrInvalidCommandQueue                = errors.New("cl: Invalid Command Queue")
	ErrInvalidHostPtr                     = errors.New("cl: Invalid Host Ptr")
	ErrInvalidMemObject                   = errors.New("cl: Invalid Mem Object")
	ErrInvalidImageFormatDescriptor       = errors.New("cl: Invalid Image Format Descriptor")
	ErrInvalidImageSize                   = errors.New("cl: Invalid Image Size")
	ErrInvalidSampler                     = errors.New("cl: Invalid Sampler")
	ErrInvalidBinary                      = errors.New("cl: Invalid Binary")
	ErrInvalidBuildOptions                = errors.New("cl: Invalid Build Options")
	ErrInvalidProgram                     = errors.New("cl: Invalid Program")
	ErrInvalidProgramExecutable           = errors.New("cl: Invalid Program Executable")
	ErrInvalidKernelName                  = errors.New("cl: Invalid Kernel Name")
	ErrInvalidKernelDefinition            = errors.New("cl: Invalid Kernel Definition")
	ErrInvalidKernel                      = errors.New("cl: Invalid Kernel")
	ErrInvalidArgIndex                    = errors.New("cl: Invalid Arg Index")
	ErrInvalidArgValue                    = errors.New("cl: Invalid Arg Value")
	ErrInvalidArgSize                     = errors.New("cl: Invalid Arg Size")
	ErrInvalidKernelArgs                  = errors.New("cl: Invalid Kernel Args")
	ErrInvalidWorkDimension               = errors.New("cl: Invalid Work Dimension")
	ErrInvalidWorkGroupSize               = errors.New("cl: Invalid Work Group Size")
	ErrInvalidWorkItemSize                = errors.New("cl: Invalid Work Item Size")
	ErrInvalidGlobalOffset                = errors.New("cl: Invalid Global Offset")
	ErrInvalidEventWaitList               = errors.New("cl: Invalid Event Wait List")
	ErrInvalidEvent                       = errors.New("cl: Invalid Event")
	ErrInvalidOperation                   = errors.New("cl: Invalid Operation")
	ErrInvalidGlObject                    = errors.New("cl: Invalid Gl Object")
	ErrInvalidBufferSize                  = errors.New("cl: Invalid Buffer Size")
	ErrInvalidMipLevel                    = errors.New("cl: Invalid Mip Level")
	ErrInvalidGlobalWorkSize              = errors.New("cl: Invalid Global Work Size")
	ErrInvalidProperty                    = errors.New("cl: Invalid Property")
	ErrInvalidImageDescriptor             = errors.New("cl: Invalid Image Descriptor")
	ErrInvalidCompilerOptions             = errors.New("cl: Invalid Compiler Options")
	ErrInvalidLinkerOptions               = errors.New("cl: Invalid Linker Options")
	ErrInvalidDevicePartitionCount        = errors.New("cl: Invalid Device Partition Count")
	ErrNoValidICDs                        = errors.New("cl: No valid ICDs found")
)

var errorMap = map[C.cl_int]error{
	C.CL_SUCCESS:                                   nil,
	C.CL_DEVICE_NOT_FOUND:                          ErrDeviceNotFound,
	C.CL_DEVICE_NOT_AVAILABLE:                      ErrDeviceNotAvailable,
	C.CL_COMPILER_NOT_AVAILABLE:                    ErrCompilerNotAvailable,
	C.CL_MEM_OBJECT_ALLOCATION_FAILURE:             ErrMemObjectAllocationFailure,
	C.CL_OUT_OF_RESOURCES:                          ErrOutOfResources,
	C.CL_OUT_OF_HOST_MEMORY:                        ErrOutOfHostMemory,
	C.CL_PROFILING_INFO_NOT_AVAILABLE:              ErrProfilingInfoNotAvailable,
	C.CL_MEM_COPY_OVERLAP:                          ErrMemCopyOverlap,
	C.CL_IMAGE_FORMAT_MISMATCH:                     ErrImageFormatMismatch,
	C.CL_IMAGE_FORMAT_NOT_SUPPORTED:                ErrImageFormatNotSupported,
	C.CL_BUILD_PROGRAM_FAILURE:                     ErrBuildProgramFailure,
	C.CL_MAP_FAILURE:                               ErrMapFailure,
	C.CL_MISALIGNED_SUB_BUFFER_OFFSET:              ErrMisalignedSubBufferOffset,
	C.CL_EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST: ErrExecStatusErrorForEventsInWaitList,
	C.CL_INVALID_VALUE:                             ErrInvalidValue,
	C.CL_INVALID_DEVICE_TYPE:                       ErrInvalidDeviceType,
	C.CL_INVALID_PLATFORM:                          ErrInvalidPlatform,
	C.CL_INVALID_DEVICE:                            ErrInvalidDevice,
	C.CL_INVALID_CONTEXT:                           ErrInvalidContext,
	C.CL_INVALID_QUEUE_PROPERTIES:                  ErrInvalidQueueProperties,
	C.CL_INVALID_COMMAND_QUEUE:                     ErrInvalidCommandQueue,
	C.CL_INVALID_HOST_PTR:                          ErrInvalidHostPtr,
	C.CL_INVALID_MEM_OBJECT:                        ErrInvalidMemObject,
	C.CL_INVALID_IMAGE_FORMAT_DESCRIPTOR:           ErrInvalidImageFormatDescriptor,
	C.CL_INVALID_IMAGE_SIZE:                        ErrInvalidImageSize,
	C.CL_INVALID_SAMPLER:                           ErrInvalidSampler,
	C.CL_INVALID_BINARY:                            ErrInvalidBinary,
	C.CL_INVALID_BUILD_OPTIONS:                     ErrInvalidBuildOptions,
	C.CL_INVALID_PROGRAM:                           ErrInvalidProgram,
	C.CL_INVALID_PROGRAM_EXECUTABLE:                ErrInvalidProgramExecutable,
	C.CL_INVALID_KERNEL_NAME:                       ErrInvalidKernelName,
	C.CL_INVALID_KERNEL_DEFINITION:                 ErrInvalidKernelDefinition,
	C.CL_INVALID_KERNEL:                            ErrInvalidKernel,
	C.CL_INVALID_ARG_INDEX:                         ErrInvalidArgIndex,
	C.CL_INVALID_ARG_VALUE:                         ErrInvalidArgValue,
	C.CL_INVALID_ARG_SIZE:                          ErrInvalidArgSize,
	C.CL_INVALID_KERNEL_ARGS:                       ErrInvalidKernelArgs,
	C.CL_INVALID_WORK_DIMENSION:                    ErrInvalidWorkDimension,
	C.CL_INVALID_WORK_GROUP_SIZE:                   ErrInvalidWorkGroupSize,
	C.CL_INVALID_WORK_ITEM_SIZE:                    ErrInvalidWorkItemSize,
	C.CL_INVALID_GLOBAL_OFFSET:                     ErrInvalidGlobalOffset,
	C.CL_INVALID_EVENT_WAIT_LIST:                   ErrInvalidEventWaitList,
	C.CL_INVALID_EVENT:                             ErrInvalidEvent,
	C.CL_INVALID_OPERATION:                         ErrInvalidOperation,
	C.CL_INVALID_GL_OBJECT:                         ErrInvalidGlObject,
	C.CL_INVALID_BUFFER_SIZE:                       ErrInvalidBufferSize,
	C.CL_INVALID_MIP_LEVEL:                         ErrInvalidMipLevel,
	C.CL_INVALID_GLOBAL_WORK_SIZE:                  ErrInvalidGlobalWorkSize,
	C.CL_INVALID_PROPERTY:                          ErrInvalidProperty,
	// OpenCL 1.2
	C.CL_COMPILE_PROGRAM_FAILURE:        ErrCompileProgramFailure,
	C.CL_DEVICE_PARTITION_FAILED:        ErrDevicePartitionFailed,
	C.CL_INVALID_COMPILER_OPTIONS:       ErrInvalidCompilerOptions,
	C.CL_INVALID_DEVICE_PARTITION_COUNT: ErrInvalidDevicePartitionCount,
	C.CL_INVALID_IMAGE_DESCRIPTOR:       ErrInvalidImageDescriptor,
	C.CL_INVALID_LINKER_OPTIONS:         ErrInvalidLinkerOptions,
	C.CL_KERNEL_ARG_INFO_NOT_AVAILABLE:  ErrKernelArgInfoNotAvailable,
	C.CL_LINK_PROGRAM_FAILURE:           ErrLinkProgramFailure,
	C.CL_LINKER_NOT_AVAILABLE:           ErrLinkerNotAvailable,
	// Extensions
	-1001: ErrNoValidICDs, // CL_PLATFORM_NOT_FOUND_KHR
}

func toError(code C.cl_int) error {
	if err, ok := errorMap[code]; ok {
		return err
	}
	return ErrOther(code)
}
