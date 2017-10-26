// +build linux,cgo darwin,cgo

package gocat

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReentrantPatchWithPotfile(t *testing.T) {
	crackedHashes := map[string]*string{}

	hc, err := New(Options{
		SharedPath:        DefaultSharedPath,
		PatchEventContext: true,
	}, callbackForTests(crackedHashes))
	defer hc.Free()

	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, hc)
	assert.NoError(t, err)

	potfilePath, err := filepath.Abs("./testdata/two_md5.potfile")
	if err != nil {
		t.Fatal(err)
	}

	hashfilePath, err := filepath.Abs("./testdata/two_md5.hashes")
	if err != nil {
		t.Fatal(err)
	}

	err = hc.RunJob("-a", "0", "-m", "0", "--potfile-path", potfilePath, hashfilePath, "./testdata/test_dictionary.txt")
	assert.NoError(t, err)
	assert.Len(t, crackedHashes, 2)
	assert.Equal(t, "hello", *crackedHashes["5d41402abc4b2a76b9719d911017c592"])
	assert.Equal(t, "world", *crackedHashes["7d793037a0760186574b0282f2f435e7"])
}

func TestReentrantPatchWithPotfileMixed(t *testing.T) {
	crackedHashes := map[string]*string{}

	hc, err := New(Options{
		SharedPath:        DefaultSharedPath,
		PatchEventContext: true,
	}, callbackForTests(crackedHashes))
	defer hc.Free()

	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, hc)
	assert.NoError(t, err)

	potfilePath, err := filepath.Abs("./testdata/one_md5_in_potfile.potfile")
	if err != nil {
		t.Fatal(err)
	}

	// We need to make a backup of this because once this cracks, hashcat will update the potfile
	orignalPotfileContents, err := ioutil.ReadFile(potfilePath)
	if err != nil {
		t.Fatal(err)
	}

	hashfilePath, err := filepath.Abs("./testdata/one_md5_in_potfile.hashes")
	if err != nil {
		t.Fatal(err)
	}

	err = hc.RunJob("-a", "0", "-m", "0", "--potfile-path", potfilePath, hashfilePath, "./testdata/test_dictionary.txt")
	assert.NoError(t, err)
	assert.Len(t, crackedHashes, 1)
	// Unfortunately, I haven't found a way to show partial potfile results so the hash for "hello" won't be returned by this test
	assert.NotContains(t, "5d41402abc4b2a76b9719d911017c592", crackedHashes)
	// ...this one should crack successfully though
	assert.Equal(t, "chris", *crackedHashes["6b34fe24ac2ff8103f6fce1f0da2ef57"])

	fd, err := os.OpenFile(potfilePath, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		t.Logf("Test will be corrupted as we were unable to open the potfile to restore the original contents: %s", err)
		t.FailNow()
	}
	fd.WriteAt(orignalPotfileContents, 0)
}
