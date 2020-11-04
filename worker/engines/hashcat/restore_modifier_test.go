package hashcat

import (
	"testing"

	"github.com/fireeye/gocat/v6/restoreutil"
	"github.com/fireeye/gocrack/opencl"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"

	"github.com/stretchr/testify/assert"
)

func TestIntSliceEq(t *testing.T) {
	for _, test := range []struct {
		SliceA  []int
		SliceB  []int
		IsEqual bool
	}{
		{
			SliceA:  []int{1, 2, 3},
			SliceB:  []int{1, 2, 3},
			IsEqual: true,
		},
		{
			SliceA:  []int{1, 3, 5},
			SliceB:  []int{1, 2, 3},
			IsEqual: false,
		},
		{
			SliceA:  []int{1},
			SliceB:  []int{2, 3},
			IsEqual: false,
		},
	} {
		equals := intSliceEq(test.SliceA, test.SliceB)
		assert.Equalf(t, test.IsEqual, equals, "expected SliceA == SliceB")
	}
}

func TestModifyRestoreFile(t *testing.T) {
	rd := &restoreutil.RestoreData{
		ArgCount: 2,
		Args: []string{
			"--opencl-devices=1",
			"--opencl-device-types=1",
		},
	}

	devicesOnMachine := shared.DeviceMap{
		1: &shared.Device{
			ID:     1,
			Name:   "My Awesome CPU",
			Type:   opencl.DeviceTypeCPU,
			IsBusy: false,
		},
		4: &shared.Device{
			ID:     4,
			Name:   "My Awesome GPU",
			Type:   opencl.DeviceTypeGPU,
			IsBusy: false,
		},
	}

	modified, err := ModifyRestoreFileDevices(rd, storage.CLDevices{4}, devicesOnMachine)
	assert.NoError(t, err)

	assert.True(t, modified)
	assert.Len(t, rd.Args, int(rd.ArgCount))
	assert.Equal(t, "--opencl-devices=4", rd.Args[0])
	assert.Equal(t, "--opencl-device-types=2", rd.Args[1])
}
