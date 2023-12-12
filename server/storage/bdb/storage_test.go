package bdb

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mandiant/gocrack/server/storage"

	"github.com/stretchr/testify/assert"
)

type storageTester struct {
	dirpath string
	*BoltBackend
}

func initTest(t assert.TestingT) *storageTester {
	path, err := ioutil.TempDir("", "storage_tests")
	if err != nil {
		assert.FailNow(t, "Failed to create temp directory for database", err.Error())
	}

	testDBPath := filepath.Join(path, "test.db")
	db, err := Init(storage.Config{
		ConnectionString: testDBPath,
	})
	if err != nil {
		assert.FailNow(t, "Failed to initialize database", err.Error())
	}

	return &storageTester{
		dirpath:     path,
		BoltBackend: db.(*BoltBackend),
	}
}

func (st *storageTester) DestroyTest() {
	if err := st.Close(); err != nil {
		log.Printf("Failed to close database: %s\n", err)
	}
	if err := os.RemoveAll(st.dirpath); err != nil {
		log.Printf("Failed to destroy directory %s: %s\n", st.dirpath, err)
	}

}
