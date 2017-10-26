package filemanager

import "bytes"

var bNewLine = []byte("\n")

// WriteSizeRecorder records the total size of a write (helpful in a multiwriter)
type WriteSizeRecorder struct {
	totalSz int64
}

func (w *WriteSizeRecorder) Write(p []byte) (n int, err error) {
	w.totalSz = w.totalSz + int64(len(p))
	return len(p), nil
}

// Size indicates the total size written to this recorder
func (w WriteSizeRecorder) Size() int64 {
	return w.totalSz
}

// WriteSizeLineRecorder records the total size of a write and includes the # of lines
type WriteSizeLineRecorder struct {
	sz       int64
	totlines int64
}

func (w *WriteSizeLineRecorder) Write(p []byte) (n int, err error) {
	w.sz = w.sz + int64(len(p))
	w.totlines = w.totlines + int64(bytes.Count(p, bNewLine))

	return len(p), nil
}

// Size indicates the total size written to this recorder
func (w WriteSizeLineRecorder) Size() int64 {
	return w.sz
}

// Lines returns the total number of new lines detected
func (w WriteSizeLineRecorder) Lines() int64 {
	return w.totlines + 1 // totlines will always be -1 off due to starting at 0
}
