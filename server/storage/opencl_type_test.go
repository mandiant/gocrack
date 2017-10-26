package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLDevicesString(t *testing.T) {
	test := CLDevices{2, 3, 4}

	out := test.String()
	assert.Equal(t, "2,3,4", out)
}

func TestCLDevicesValue(t *testing.T) {
	test := CLDevices{2, 3, 4}

	out, err := test.Value()
	assert.Nil(t, err)
	assert.Equal(t, "[2,3,4]", out)

	test = CLDevices{}
	out, err = test.Value()
	assert.Nil(t, err)
	assert.Nil(t, out)
}

func TestCLDevicesScan(t *testing.T) {
	var devs CLDevices

	err := devs.Scan(nil)
	assert.Nil(t, err)
	assert.Nil(t, devs)

	err = devs.Scan([]byte(`[2,3,4]`))
	assert.Nil(t, err)
	assert.Equal(t, CLDevices{2, 3, 4}, devs)

	err = devs.Scan("hello")
	assert.EqualError(t, err, "bad type assertion")
}
