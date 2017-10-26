package gocat

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"unsafe"

	"github.com/fireeye/gocrack/gocat/hcargp"

	"github.com/stretchr/testify/assert"
)

const (
	// Set this to true if you want the gocat callbacks used in the tests to print out
	DebugTest         bool   = true
	DefaultSharedPath string = "/usr/local/share/hashcat"
)

type testStruct struct {
	opts          hcargp.HashcatSessionOptions
	expectedError error
}

func emptyCallback(hc unsafe.Pointer, payload interface{}) {}

func callbackForTests(resultsmap map[string]*string) EventCallback {
	return func(hc unsafe.Pointer, payload interface{}) {
		switch pl := payload.(type) {
		case LogPayload:
			if DebugTest {
				fmt.Printf("LOG [%s] %s\n", pl.Level, pl.Message)
			}
		case ActionPayload:
			if DebugTest {
				fmt.Printf("ACTION [%d] %s\n", pl.HashcatEvent, pl.Message)
			}
		case CrackedPayload:
			if DebugTest {
				fmt.Printf("CRACKED %s -> %s\n", pl.Hash, pl.Value)
			}
			if resultsmap != nil {
				resultsmap[pl.Hash] = hcargp.GetStringPtr(pl.Value)
			}
		case FinalStatusPayload:
			if DebugTest {
				fmt.Printf("FINAL STATUS -> %v\n", pl.Status)
			}
		case TaskInformationPayload:
			if DebugTest {
				fmt.Printf("TASK INFO -> %v\n", pl)
			}
		}
	}
}

func TestOptionsExecPath(t *testing.T) {
	// Valid
	opts := Options{
		ExecutablePath: "",
		SharedPath:     "/tmp",
	}

	err := opts.validate()
	assert.Nil(t, err)
	assert.True(t, strings.HasSuffix(opts.ExecutablePath, "_test"))

	// Not valid because executable path was incorrectly set by the user
	opts.ExecutablePath = "/nope"
	err = opts.validate()
	assert.Error(t, err)
}

func TestGoCatOptionsValidatorErrors(t *testing.T) {
	for _, test := range []struct {
		opts          Options
		expectedError error
		expectedOpts  map[string]interface{}
	}{
		{
			opts: Options{
				SharedPath: "",
			},
			expectedError: ErrNoSharedPath,
		},
		{
			opts: Options{
				SharedPath:     "/deadbeef",
				ExecutablePath: "",
			},
		},
	} {
		err := test.opts.validate()
		assert.Equal(t, test.expectedError, err)
	}
}

func TestGoCatCrackingMD5(t *testing.T) {
	crackedHashes := map[string]*string{}

	hc, err := New(Options{
		SharedPath: DefaultSharedPath,
	}, callbackForTests(crackedHashes))
	defer hc.Free()

	assert.NotNil(t, hc)
	assert.NoError(t, err)

	err = hc.RunJob("-a", "0", "-m", "0", "--potfile-disable", "5d41402abc4b2a76b9719d911017c592", "./testdata/test_dictionary.txt")
	assert.NoError(t, err)
	assert.Len(t, crackedHashes, 1)
	assert.Equal(t, "hello", *crackedHashes["5d41402abc4b2a76b9719d911017c592"])
}

func TestGoCatReusingContext(t *testing.T) {
	crackedHashes := map[string]*string{}

	hc, err := New(Options{
		SharedPath: DefaultSharedPath,
	}, callbackForTests(crackedHashes))
	defer hc.Free()

	assert.NotNil(t, hc)
	assert.NoError(t, err)

	err = hc.RunJob("-a", "0", "-m", "0", "--potfile-disable", "5d41402abc4b2a76b9719d911017c592", "./testdata/test_dictionary.txt")
	assert.NoError(t, err)
	assert.Len(t, crackedHashes, 1)
	assert.Equal(t, "hello", *crackedHashes["5d41402abc4b2a76b9719d911017c592"])

	err = hc.RunJob("-a", "0", "-m", "0", "--potfile-disable", "9f9d51bc70ef21ca5c14f307980a29d8", "./testdata/test_dictionary.txt")
	assert.NoError(t, err)
	assert.Len(t, crackedHashes, 2) // the previous run will still exist in this map
	assert.Equal(t, "bob", *crackedHashes["9f9d51bc70ef21ca5c14f307980a29d8"])
}

func TestGoCatRunJobWithOptions(t *testing.T) {
	crackedHashes := map[string]*string{}

	hc, err := New(Options{
		SharedPath: DefaultSharedPath,
	}, callbackForTests(crackedHashes))
	defer hc.Free()

	assert.NotNil(t, hc)
	assert.NoError(t, err)

	err = hc.RunJobWithOptions(hcargp.HashcatSessionOptions{
		AttackMode:                   hcargp.GetIntPtr(0),
		HashType:                     hcargp.GetIntPtr(0),
		PotfileDisable:               hcargp.GetBoolPtr(true),
		InputFile:                    "9f9d51bc70ef21ca5c14f307980a29d8",
		DictionaryMaskDirectoryInput: hcargp.GetStringPtr("./testdata/test_dictionary.txt"),
	})

	assert.NoError(t, err)
	assert.Len(t, crackedHashes, 1) // the previous run will still exist in this map
	assert.Equal(t, "bob", *crackedHashes["9f9d51bc70ef21ca5c14f307980a29d8"])
}

func TestGocatRussianHashes(t *testing.T) {
	crackedHashes := map[string]*string{}

	hc, err := New(Options{
		SharedPath: DefaultSharedPath,
	}, callbackForTests(crackedHashes))
	defer hc.Free()

	assert.NotNil(t, hc)
	assert.NoError(t, err)

	err = hc.RunJobWithOptions(hcargp.HashcatSessionOptions{
		AttackMode:                   hcargp.GetIntPtr(0),
		HashType:                     hcargp.GetIntPtr(0),
		PotfileDisable:               hcargp.GetBoolPtr(true),
		InputFile:                    "./testdata/russian_test.hashes",
		DictionaryMaskDirectoryInput: hcargp.GetStringPtr("./testdata/russian_test.dictionary"),
	})

	assert.NoError(t, err)
	assert.Len(t, crackedHashes, 4) // the previous run will still exist in this map
	fmt.Println("HI", crackedHashes)
	fmt.Println(crackedHashes)
}

func TestGoCatStopAtCheckpointWithNoRunningSession(t *testing.T) {
	hc, err := New(Options{
		SharedPath: DefaultSharedPath,
	}, emptyCallback)
	defer hc.Free()

	assert.NotNil(t, hc)
	assert.NoError(t, err)

	err = hc.StopAtCheckpoint()
	assert.Equal(t, ErrUnableToStopAtCheckpoint, err)
}

func ExampleHashcat_RunJobWithOptions() {

	eventCallback := func(hc unsafe.Pointer, payload interface{}) {
		switch pl := payload.(type) {
		case LogPayload:
			if DebugTest {
				fmt.Printf("LOG [%s] %s\n", pl.Level, pl.Message)
			}
		case ActionPayload:
			if DebugTest {
				fmt.Printf("ACTION [%d] %s\n", pl.HashcatEvent, pl.Message)
			}
		case CrackedPayload:
			if DebugTest {
				fmt.Printf("CRACKED %s -> %s\n", pl.Hash, pl.Value)
			}
		case FinalStatusPayload:
			if DebugTest {
				fmt.Printf("FINAL STATUS -> %v\n", pl.Status)
			}
		case TaskInformationPayload:
			if DebugTest {
				fmt.Printf("TASK INFO -> %v\n", pl)
			}
		}
	}

	hc, err := New(Options{
		SharedPath: "/usr/local/share/hashcat",
	}, eventCallback)
	defer hc.Free()

	if err != nil {
		log.Fatal(err)
	}

	err = hc.RunJobWithOptions(hcargp.HashcatSessionOptions{
		AttackMode:                   hcargp.GetIntPtr(0),
		HashType:                     hcargp.GetIntPtr(0),
		PotfileDisable:               hcargp.GetBoolPtr(true),
		InputFile:                    "9f9d51bc70ef21ca5c14f307980a29d8",
		DictionaryMaskDirectoryInput: hcargp.GetStringPtr("./testdata/test_dictionary.txt"),
	})

	if err != nil {
		log.Fatal(err)
	}
}
