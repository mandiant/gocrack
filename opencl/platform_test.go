package opencl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type getterTest struct {
	Out   string
	Error error
}

func TestGetPlatforms(t *testing.T) {
	platforms, err := GetPlatforms()
	if err == ErrInvalidValue || err == ErrNoValidICDs {
		t.Skip("Skipping test as no platforms are available")
	}
	assert.Nil(t, err)
	assert.True(t, len(platforms) > 0)

	getterTests := make(map[string]getterTest)
	for _, platform := range platforms {
		val, err := platform.Name()
		getterTests["Name"] = getterTest{val, err}
		val, err = platform.Profile()
		getterTests["Profile"] = getterTest{val, err}
		val, err = platform.Vendor()
		getterTests["Vendor"] = getterTest{val, err}
		val, err = platform.Version()
		getterTests["Version"] = getterTest{val, err}

		for api, test := range getterTests {
			assert.Nilf(t, test.Error, "platform.%s() should have a nil error", api)
			assert.NotEmpty(t, test.Out, "platform.%s() should have returned an empty string", api)
		}

		devices, err := platform.GetDevices(DeviceTypeAll)
		assert.Nil(t, err, "This can happen if you have a platform installed that legitimately has no devices")
		assert.True(t, len(devices) > 0)
	}
}
