package filemanager

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/mandiant/gocrack/shared"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	tdir, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "could not create temp directory")
	}

	for _, test := range []struct {
		ConfigStruct  Config
		ExpectedError error
	}{
		// Valid Config
		{
			ConfigStruct: Config{
				TaskUploadPath: tdir,
				EngineFilePath: tdir,
				TempPath:       tdir,
			},
			ExpectedError: nil,
		},
		// Fail File Manager
		{
			ConfigStruct: Config{
				TaskUploadPath: "not_a_valid_directory",
				EngineFilePath: "not_a_valid_directory",
				TempPath:       "not_a_valid_directory",
			},
			ExpectedError: fmt.Errorf(errNotValidFolder, "task_file_path"),
		},
		// Valid Import Directory
		{
			ConfigStruct: Config{
				TaskUploadPath: tdir,
				EngineFilePath: tdir,
				TempPath:       tdir,
				ImportPath:     shared.GetStrPtr(tdir),
			},
			ExpectedError: nil,
		},
		// Invalid import path
		{
			ConfigStruct: Config{
				TaskUploadPath: tdir,
				EngineFilePath: tdir,
				TempPath:       tdir,
				ImportPath:     shared.GetStrPtr("not valid"),
			},
			ExpectedError: fmt.Errorf(errNotValidFolder, "import_path"),
		},
	} {
		err := test.ConfigStruct.Validate()
		if test.ExpectedError != nil {
			assert.EqualError(t, err, test.ExpectedError.Error())
			continue
		}

		if test.ConfigStruct.TaskMaxSize == nil {
			assert.Equal(t, defaultMaxSize, test.ConfigStruct.TaskMaxSize)
		}

		assert.NoError(t, err)
	}
}
