package hcargp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleHashcatSessionOptions_MarshalArgs() {
	opts := HashcatSessionOptions{
		AttackMode:     GetIntPtr(0),
		HashType:       GetIntPtr(0),
		SessionName:    GetStringPtr("example_args_session"),
		PotfileDisable: GetBoolPtr(true),
		InputFile:      "deadbeefdeadbeefdeadbeefdeadbeef",
	}

	args, err := opts.MarshalArgs()
	if err != nil {
		fmt.Printf("Failed to marshal args: %s\n", err)
		return
	}

	fmt.Println(strings.Join(args, " "))
	// Output: --hash-type=0 --attack-mode=0 --session=example_args_session --potfile-disable deadbeefdeadbeefdeadbeefdeadbeef
}

func TestInternalParseTag(t *testing.T) {
	tag, opts := parseTag("--example,omitempty,required")

	assert.Equal(t, "--example", tag)
	assert.Equal(t, "omitempty,required", opts)

	tag, opts = parseTag("--example2,")
	assert.Equal(t, "--example2", tag)
	assert.Equal(t, "", opts)
}

func TestHashcatSessionOptionsMarshalArgs(t *testing.T) {
	for _, test := range []struct {
		opts          HashcatSessionOptions
		expectedError error
		expectedArgs  []string
	}{
		{
			opts: HashcatSessionOptions{
				AttackMode: GetIntPtr(0),
				HashType:   nil,
				InputFile:  "deadbeefdeadbeefdeadbeefdeadbeef",
			},
			expectedError: nil,
			expectedArgs:  []string{"--attack-mode=0", "deadbeefdeadbeefdeadbeefdeadbeef"},
		},
		{
			opts: HashcatSessionOptions{
				IsHexCharset: GetBoolPtr(true),
				InputFile:    "deadbeefdeadbeefdeadbeefdeadbeef",
			},
			expectedError: nil,
			expectedArgs:  []string{"--hex-charset", "deadbeefdeadbeefdeadbeefdeadbeef"},
		},
		{
			opts: HashcatSessionOptions{
				IsHexCharset: GetBoolPtr(false),
				InputFile:    "deadbeefdeadbeefdeadbeefdeadbeef",
			},
			expectedError: nil,
			expectedArgs:  []string{"deadbeefdeadbeefdeadbeefdeadbeef"},
		},
		{
			opts: HashcatSessionOptions{
				AttackMode:                   GetIntPtr(0),
				InputFile:                    "deadbeefdeadbeefdeadbeefdeadbeef",
				DictionaryMaskDirectoryInput: GetStringPtr("./testdata/test_dictionary.txt"),
			},
			expectedError: nil,
			expectedArgs:  []string{"--attack-mode=0", "deadbeefdeadbeefdeadbeefdeadbeef", "./testdata/test_dictionary.txt"},
		},
	} {
		args, err := test.opts.MarshalArgs()

		assert.Equal(t, test.expectedError, err)
		assert.Equal(t, test.expectedArgs, args)
	}
}
