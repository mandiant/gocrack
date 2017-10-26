package gocat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevelString(t *testing.T) {
	for _, test := range []struct {
		ll       LogLevel
		expected string
	}{
		{
			ll:       InfoMessage,
			expected: "INFO",
		},
		{
			ll:       WarnMessage,
			expected: "WARN",
		},
		{
			ll:       ErrorMessage,
			expected: "ERROR",
		},
		{
			ll:       AdviceMessage,
			expected: "ADVICE",
		},
		{
			ll:       LogLevel(9),
			expected: "UNKNOWN",
		},
	} {
		assert.Equal(t, test.expected, test.ll.String())
	}
}

func TestGetCrackedPassword(t *testing.T) {
	pl, err := getCrackedPassword(1, "deadbeefdeadbeefdeadbeefdeadbeef:chris", ":")
	assert.Nilf(t, err, "expected err to be nil")
	assert.Equalf(t, "chris", pl.Value, "expected the value to be chris")
	assert.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeef", pl.Hash, "expected the hash to be deadbeef yo!")

	pl, err = getCrackedPassword(1, "deadbeefdeadbeefdeadbeefdeadbeef:chris", ";")
	assert.Equalf(t, err, ErrCrackedPayload{
		Separator:  ";",
		CrackedMsg: "deadbeefdeadbeefdeadbeefdeadbeef:chris",
	}, "err payload is not correct")

	ecperr, ok := err.(ErrCrackedPayload)
	assert.True(t, ok)
	assert.NotNil(t, ecperr)
	assert.Equal(t, "Could not locate separator `;` in msg", err.Error())
}
