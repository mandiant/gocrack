package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	for _, test := range []struct {
		TestConfig Config
		Error      error
	}{
		{
			TestConfig: Config{
				Backend: "",
			},
			Error: noBackend,
		},
		{
			TestConfig: Config{
				Backend: "hello",
			},
			Error: noConnectionString,
		},
		{
			TestConfig: Config{
				Backend:          "hello",
				ConnectionString: "world",
			},
		},
	} {
		if err := test.TestConfig.Validate(); err != nil {
			assert.Equal(t, test.Error, err)
		}
	}
}
