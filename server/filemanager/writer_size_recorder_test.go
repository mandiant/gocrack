package filemanager

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteSizeRecorder(t *testing.T) {
	testBufWithData := bytes.NewBufferString("this is a test string that we are going to copy!")
	dest := new(bytes.Buffer)
	szRec := &WriteSizeRecorder{}
	mw := io.MultiWriter(dest, szRec)

	if _, err := io.Copy(mw, testBufWithData); err != nil {
		assert.NoError(t, err)
		return
	}

	assert.Equal(t, int64(48), szRec.Size())
}

func TestWriteSizeLineRecorder(t *testing.T) {
	testBufWithData := bytes.NewBuffer([]byte("this is a test string that we are going to copy!\nbut with multiple lines!\nlike this!"))
	dest := new(bytes.Buffer)
	szRec := &WriteSizeLineRecorder{}
	mw := io.MultiWriter(dest, szRec)

	if _, err := io.Copy(mw, testBufWithData); err != nil {
		assert.NoError(t, err)
		return
	}

	assert.Equal(t, int64(84), szRec.Size())
	assert.Equal(t, int64(3), szRec.Lines())
}
