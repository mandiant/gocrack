package client

import (
	"bytes"
	"compress/gzip"
	"sync"
	"testing"

	"github.com/fireeye/gocrack/shared/pool"

	"github.com/stretchr/testify/assert"
)

// testBytes is "hello world"
var testBytes = []byte{31, 139, 8, 0, 121, 22, 184, 89, 2, 255, 203, 72, 205, 201, 201, 47, 207, 47, 202, 73, 1, 0, 173, 32, 235, 249, 10, 0, 0, 0}

// TestByteReadCloser implements io.ReadCloser
type TestByteReadCloser struct {
	*bytes.Reader
}

func (s TestByteReadCloser) Close() error {
	return nil
}

func TestInternal_gzReader(t *testing.T) {
	b := bytes.NewReader(testBytes)
	rdr := TestByteReadCloser{b}
	pool := pool.MetricsPool{
		P: sync.Pool{
			New: func() interface{} {
				return new(gzip.Reader)
			},
		},
	}

	gz := pool.Get().(*gzip.Reader)
	gz.Reset(b)
	assert.Equal(t, pool.Releases(), int32(1))

	c := gzReader{sock: rdr, gz: gz, p: &pool}
	c.Close()

	assert.Equal(t, pool.Releases(), int32(0))
}
