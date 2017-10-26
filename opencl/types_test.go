package opencl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceType(t *testing.T) {
	dt := DeviceTypeCPU | DeviceTypeAccelerator
	out := dt.String()

	assert.Equal(t, "CPU|Accelerator", out)

	dt = DeviceTypeAll
	assert.Equal(t, "CPU|GPU|Accelerator|Default", dt.String())
}
