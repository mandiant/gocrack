package bdb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type StringContainsTestCase struct {
	Field string
}

func TestStringContains(t *testing.T) {
	for i, test := range []struct {
		Field         interface{}
		ExpectedError error
		Match         bool
	}{
		{
			Field: "Hello",
			Match: true,
		},
		{
			Field: "not a match",
			Match: false,
		},
		{
			Field:         []byte("hello"), // wrong type
			Match:         false,
			ExpectedError: fmt.Errorf("Only string is supported for StringContains matcher, got []byte"),
		},
	} {
		q := StringContains("Field", "Hello")
		match, err := q.Match(&test)

		if test.ExpectedError != nil && err == nil {
			assert.Fail(t, fmt.Sprintf("test %d expected an error but err was nil in match", i))
			continue
		} else if test.ExpectedError == nil && err != nil {
			assert.Fail(t, fmt.Sprintf("test %d did not expect an error but got one", i), err.Error())
		}

		assert.Equal(t, test.Match, match)
	}
}
