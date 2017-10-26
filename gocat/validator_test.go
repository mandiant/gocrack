package gocat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHashes(t *testing.T) {
	for i, test := range []struct {
		HashPathToTest       string
		HashType             uint32
		IsValid              bool
		ExpectedHashes       uint32
		ExpectedUniqueHashes uint32
		ExpectedSalts        uint32
		ExpectedNumErrors    int
	}{
		{
			HashPathToTest:       "./testdata/mix_of_invalid_and_valid.hashes",
			HashType:             0,
			IsValid:              false,
			ExpectedHashes:       1,
			ExpectedUniqueHashes: 1,
			ExpectedSalts:        1,
			ExpectedNumErrors:    2,
		},
		{
			HashPathToTest:       "./testdata/two_md5.hashes",
			HashType:             0,
			IsValid:              true,
			ExpectedHashes:       2,
			ExpectedUniqueHashes: 2,
			ExpectedSalts:        1,
		},
	} {
		vr, err := ValidateHashes(test.HashPathToTest, test.HashType)
		if err != nil {
			assert.FailNow(t, "failed to initialize the validator")
		}

		assert.Equalf(t, test.IsValid, vr.Valid, "failed equality check in test %d", i)
		assert.Equalf(t, test.ExpectedHashes, vr.NumHashes, "failed equality check in test %d", i)
		assert.Equalf(t, test.ExpectedUniqueHashes, vr.NumHashesUnique, "failed equality check in test %d", i)
		assert.Equalf(t, test.ExpectedSalts, vr.NumSalts, "failed equality check in test %d", i)
		assert.Equalf(t, test.ExpectedNumErrors, len(vr.Errors), "failed equality check in test %d", i)
	}
}
