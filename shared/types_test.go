package shared

import (
	"testing"

	"github.com/mandiant/gocrack/opencl"
	"github.com/stretchr/testify/assert"
)

func TestDeviceMap(t *testing.T) {
	dmTest := DeviceMap{
		1: &Device{
			ID:   1,
			Name: "My CPU1",
			Type: opencl.DeviceTypeCPU,
		},
		2: &Device{
			ID:   2,
			Name: "My GPU",
			Type: opencl.DeviceTypeGPU,
		},
		3: &Device{
			ID:   3,
			Name: "My GPU2",
			Type: opencl.DeviceTypeGPU,
		},
	}

	devices := dmTest.PickFreeDevices(opencl.DeviceTypeGPU, 1)
	assert.NotContainsf(t, devices, 1, "devices list contains a CPU")

	dmTest.MarkAsBusy(devices)
	for _, device := range devices {
		assert.True(t, dmTest[device].IsBusy)
	}

	dmTest.MarkAsFree(devices)
	for _, device := range devices {
		assert.False(t, dmTest[device].IsBusy)
	}
}

func TestIntSliceToString(t *testing.T) {
	ints := []int{3, 10, 200}
	out := IntSliceToString(ints)

	assert.Equal(t, "3,10,200", out)
}

func TestGetPtrs(t *testing.T) {
	var testString = "hello"
	assert.Equal(t, testString, *GetStrPtr(testString))

	var testInt = 1337
	assert.Equal(t, testInt, *GetIntPtr(testInt))
}
