package filemanager

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	fakestor "github.com/mandiant/gocrack/server/filemanager/test"
	"github.com/mandiant/gocrack/server/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSplitFilePath(t *testing.T) {
	for i, test := range []struct {
		RootPath      string
		InitialString string
		OutString     string
	}{
		{
			RootPath:      "/tmp/hello",
			InitialString: "c9c435d08cda8bdcedf0e61c229aebc89e44cf21",
			OutString:     "/tmp/hello/c9/c4/35/d0/8c/da/8b/c9c435d08cda8bdcedf0e61c229aebc89e44cf21",
		},
		{
			RootPath:      "/",
			InitialString: "5d41402abc4b2a76b9719d911017c592",
			OutString:     "/5d/41/40/2a/bc/4b/2a/5d41402abc4b2a76b9719d911017c592",
		},
		{
			RootPath:      "/tmp/",
			InitialString: "d78ddfec-5e12-4238-a46b-8e218d4cfcbe",
			OutString:     "/tmp/d7/8d/df/ec/5e/12/42/d78ddfec5e124238a46b8e218d4cfcbe",
		},
	} {
		out := splitFilePath(test.RootPath, test.InitialString)
		if out != test.OutString {
			assert.Equalf(t, test.OutString, out, "output from splitFilePath does not match in test %d", i)
		}
	}
}
func TestSaveEngineFile(t *testing.T) {
	simpl := &fakestor.TestStorageImpl{}
	path, err := ioutil.TempDir("", "")
	if err != nil {
		assert.FailNow(t, "failed to create temp directory", err.Error())
	}
	defer os.RemoveAll(path)

	fm := New(simpl, Config{
		TaskUploadPath: path,
		EngineFilePath: path,
		TempPath:       path,
	})

	testfile, err := ioutil.TempFile("", "")
	if err != nil {
		assert.FailNow(t, "failed to create test file", err.Error())
	}
	defer os.Remove(testfile.Name())

	if _, err := testfile.WriteString("helloworld\nthisisatest"); err != nil {
		assert.FailNow(t, "failed to create test file", err.Error())
	}

	if _, err := testfile.Seek(io.SeekStart, 0); err != nil {
		assert.FailNow(t, "failed to create test file", err.Error())
	}

	uuidToCreate := uuid.NewString()
	fresp, err := fm.SaveFile(testfile, "testing", uuidToCreate, storage.EngineFileDictionary)
	if err != nil {
		assert.FailNow(t, "failed to save test file", err.Error())
	}

	assert.Equal(t, "fc4c80ef680e79f2b102802cdca469b0e23eadc8", fresp.SHA1)
	assert.Equal(t, int64(2), fresp.NumberOfLines)
	assert.NotEmpty(t, fresp.SavedTo)
}
