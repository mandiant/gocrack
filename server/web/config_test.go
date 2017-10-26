package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig(t *testing.T) {
	for _, test := range []struct {
		cfg           Config
		expectedError error
	}{
		// Valid
		{
			cfg: Config{
				Listener: listener{},
				CORS: corsSettings{
					AllowedOrigins: []string{"http://mytest.com"},
				},
				UserInterface: uiSettings{
					CSRFKey: "testing",
				},
			},
			expectedError: nil,
		},
		// Bad
		{
			cfg: Config{
				Listener: listener{},
				CORS: corsSettings{
					AllowedOrigins: []string{},
				},
				UserInterface: uiSettings{
					CSRFKey: "testing",
				},
			},
			expectedError: errMissingCORSOrigins,
		},
		{
			cfg: Config{
				Listener: listener{},
				CORS: corsSettings{
					AllowedOrigins: []string{
						"http://localhost",
					},
				},
				UserInterface: uiSettings{
					CSRFKey: "",
				},
			},
			expectedError: errMissingCSRFKey,
		},
	} {
		err := test.cfg.Validate()
		if err != test.expectedError {
			assert.Fail(t, "expected errors do not match", err, test.expectedError)
		}
	}
}
