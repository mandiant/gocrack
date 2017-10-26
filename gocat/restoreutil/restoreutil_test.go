package restoreutil

import (
	"bytes"
	"crypto/md5"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRestoreFileContents(rd RestoreData, t *testing.T) {
	assert.Equal(t, uint32(0x15e), rd.Version)
	assert.Equal(t, "/Users/cschmitt/Desktop", rd.WorkingDirectory)
	assert.Equal(t, uint32(0x0), rd.DictionaryPosition)
	assert.Equal(t, uint32(0x4), rd.MasksPosition)
	assert.Equal(t, uint64(0xaf0000), rd.WordsPosition)
	assert.Equal(t, uint32(8), rd.ArgCount)
	assert.Equal(t, 8, len(rd.Args))
}

func TestReadRestoreFile(t *testing.T) {
	rd, err := ReadRestoreFile("./testdata/unittest_example.restore")
	assert.Nil(t, err)
	testRestoreFileContents(rd, t)
}

func TestRestoreBytes(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/unittest_example.restore")
	assert.Nil(t, err)

	rd, err := ReadRestoreBytes(b)
	assert.Nil(t, err)
	testRestoreFileContents(rd, t)
}

func TestWrite(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/unittest_example.restore")
	assert.Nil(t, err)

	beforeXsum := md5.Sum(b)

	rd, err := ReadRestoreBytes(b)
	assert.Nil(t, err)

	buf := new(bytes.Buffer)
	// Write the restore file to a byte Buffer
	err = rd.Write(buf)
	assert.Nil(t, err)

	afterXsum := md5.Sum(buf.Bytes())
	assert.Equalf(t, beforeXsum, afterXsum, "checksum mismatch")
}
