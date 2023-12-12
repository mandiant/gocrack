package authentication

import (
	"strings"
	"testing"
	"time"

	test "github.com/mandiant/gocrack/server/authentication/test"
	"github.com/mandiant/gocrack/server/storage"
	"github.com/mandiant/gocrack/shared"

	"github.com/stretchr/testify/assert"
)

var testSecretKey = "aw3som3_Security!@"

func TestConfigValidation(t *testing.T) {
	for _, test := range []struct {
		cfg           *AuthSettings
		ExpectedError error
	}{
		// Valid with default expiration
		{
			cfg: &AuthSettings{
				Backend:   "database",
				SecretKey: &testSecretKey,
			},
		},
		// Valid with user set token duration
		{
			cfg: &AuthSettings{
				Backend:     "database",
				SecretKey:   &testSecretKey,
				TokenExpiry: &shared.HumanDuration{Duration: 1 * time.Hour},
			},
		},
		// Invalid with negative token expiry
		{
			cfg: &AuthSettings{
				Backend:     "database",
				SecretKey:   &testSecretKey,
				TokenExpiry: &shared.HumanDuration{Duration: -1 * time.Hour},
			},
			ExpectedError: errTokenExpiryNegative,
		},
		// Invalid with no backend
		{
			cfg: &AuthSettings{
				Backend:     "",
				SecretKey:   &testSecretKey,
				TokenExpiry: &shared.HumanDuration{Duration: 1 * time.Hour},
			},
			ExpectedError: errNoBackend,
		},
		// Invalid secret key
		{
			cfg: &AuthSettings{
				Backend:     "database",
				SecretKey:   nil,
				TokenExpiry: &shared.HumanDuration{Duration: 1 * time.Hour},
			},
			ExpectedError: errSecretKeyBad,
		},
	} {
		err := test.cfg.Validate()
		assert.Equal(t, test.ExpectedError, err)
	}
}

func TestAuthWrapper_Login(t *testing.T) {
	fakedb := test.NewFakeDatabase()
	dbauth := test.NewFakeAuthProv(fakedb)

	dbauth.CreateUser(storage.User{
		UserUUID: "013337-deadbeef",
		Username: "test_user",
		Password: "myawesomepassword",
	})

	wrapper := WrapProvider(dbauth, AuthSettings{
		SecretKey:   &testSecretKey,
		TokenExpiry: &DefaultTokenExpiry,
	})

	claim, err := wrapper.Login("test_user", "myawesomepassword", true)
	assert.Nil(t, err)
	assert.True(t, strings.Contains(claim, "."))

	claim, err = wrapper.Login("test_user", "no", true)
	assert.NotNil(t, err)
	assert.Empty(t, claim)
}

func TestAuthWrapper_VerifyClaim(t *testing.T) {
	fakedb := test.NewFakeDatabase()
	dbauth := test.NewFakeAuthProv(fakedb)

	dbauth.CreateUser(storage.User{
		UserUUID: "013337-deadbeef",
		Username: "test_user",
		Password: "myawesomepassword",
	})

	wrapper := WrapProvider(dbauth, AuthSettings{
		SecretKey:   &testSecretKey,
		TokenExpiry: &DefaultTokenExpiry,
	})

	rawClaim, err := wrapper.Login("test_user", "myawesomepassword", true)
	assert.Nil(t, err)
	assert.NotEmpty(t, rawClaim)

	parsedClaim, err := wrapper.VerifyClaim(rawClaim, "gocrack")
	assert.Nil(t, err)
	assert.Equal(t, "test_user", parsedClaim.Username)
	assert.Equal(t, "013337-deadbeef", parsedClaim.UserUUID)

	// Validate a missing audience
	parsedClaim, err = wrapper.VerifyClaim(rawClaim, "gocrack", "NotAValidAudience")
	assert.EqualError(t, err, "missing aud `NotAValidAudience` in claim")
	assert.Nil(t, parsedClaim)
}

func TestAuthWrapper_VerifyExpiredClaim(t *testing.T) {
	fakedb := test.NewFakeDatabase()
	dbauth := test.NewFakeAuthProv(fakedb)
	dbauth.CreateUser(storage.User{
		UserUUID: "013337-deadbeef",
		Username: "test_user",
		Password: "myawesomepassword",
	})

	wrapper := WrapProvider(dbauth, AuthSettings{
		SecretKey:   &testSecretKey,
		TokenExpiry: &shared.HumanDuration{Duration: -2 * time.Minute},
	})

	rawClaim, err := wrapper.Login("test_user", "myawesomepassword", true)
	assert.Nil(t, err)
	assert.NotEmpty(t, rawClaim)

	parsedClaim, err := wrapper.VerifyClaim(rawClaim, "gocrack")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "authentication has expired")
	assert.Nil(t, parsedClaim)
}
