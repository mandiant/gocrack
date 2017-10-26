package database

import (
	"errors"
	"strings"
	"testing"

	"github.com/fireeye/gocrack/server/authentication"
	test "github.com/fireeye/gocrack/server/authentication/test"
	"github.com/fireeye/gocrack/server/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type TestDBAuth struct {
	suite.Suite

	db   authentication.AuthStorageBackend
	auth *DatabaseAuth
}

func (suite *TestDBAuth) SetupTest() {
	var err error
	testdb := test.NewFakeDatabase()
	if suite.auth, err = Init(testdb, nil); err != nil {
		suite.FailNow("could not create auth adapter", err.Error())
	}
}

func (suite *TestDBAuth) TestNewDatabaseAuthWithBadConfig() {
	badConfig := map[string]interface{}{
		"allow_registration": "hello",
	}

	testdb := test.NewFakeDatabase()
	auth, err := Init(testdb, badConfig)
	suite.Nil(auth)
	suite.Error(err)
}

func (suite *TestDBAuth) TestCreateUserWithLogin() {
	err := suite.auth.CreateUser(storage.User{
		Username: "test_user",
		UserUUID: "000000000-0000-0000-000000000000",
		Password: "!Strong1p@ssword",
	})
	suite.Nil(err)

	userRecord, err := suite.auth.Login("test_user", "!Strong1p@ssword")
	suite.Nil(err)
	suite.Equal("test_user", userRecord.Username)

	userRecord, err = suite.auth.Login("test_user", "incorrect_password")
	suite.NotNil(err)
	suite.Nil(userRecord)
	suite.Equal(storage.ErrNotFound, err)
}

func (suite *TestDBAuth) TestUserCanChangePassword() {
	suite.True(suite.auth.UserCanChangePassword())
}

func (suite *TestDBAuth) TestGenerateSecurePassword() {
	hashedValue, err := suite.auth.GenerateSecurePassword("!Strong1p@ssword")
	suite.Nil(err)
	suite.Equal(3, strings.Count(hashedValue, "$"))

	hashedValue, err = suite.auth.GenerateSecurePassword("")
	suite.Equal("", hashedValue)
	suite.Equal(err, authentication.ErrPasswordEmpty)

	hashedValue, err = suite.auth.GenerateSecurePassword("notstrong")
	suite.Equal("", hashedValue)
	suite.Equal(authentication.ErrFailsRequirements, err)
}

func TestDBAuthSuite(t *testing.T) {
	suite.Run(t, new(TestDBAuth))
}

func TestInternal_convertRawConfigToConfig(t *testing.T) {
	for _, test := range []struct {
		Config         map[string]interface{}
		ExpectedError  error
		ExpectedConfig *Config
	}{
		{
			Config: map[string]interface{}{
				"bcrypt_cost": "yolo",
			},
			ExpectedError:  errors.New("authentication.backend_settings.bcrypt_cost is of the wrong type. Expected int, got string"),
			ExpectedConfig: nil,
		},
		{
			Config: map[string]interface{}{
				"bcrypt_cost": 10,
			},
			ExpectedError: nil,
			ExpectedConfig: &Config{
				BCryptCost: 10,
			},
		},
		{
			Config: map[string]interface{}{
				"bcrypt_cost": bcrypt.MaxCost + 10,
			},
			ExpectedError: nil,
			ExpectedConfig: &Config{
				BCryptCost: bcrypt.DefaultCost,
			},
		},
		// Invalid Type on allow_registration
		{
			Config: map[string]interface{}{
				"bcrypt_cost":        bcrypt.DefaultCost,
				"allow_registration": "hello",
			},
			ExpectedError: errors.New("authentication.backend_settings.allow_registration is of the wrong type. Expected bool, got string"),
			ExpectedConfig: &Config{
				BCryptCost: bcrypt.DefaultCost,
			},
		},
		// Good!
		{
			Config: map[string]interface{}{
				"bcrypt_cost":        bcrypt.DefaultCost,
				"allow_registration": true,
			},
			ExpectedError: nil,
			ExpectedConfig: &Config{
				BCryptCost:        bcrypt.DefaultCost,
				AllowRegistration: true,
			},
		},
	} {
		parsedConfig, err := convertRawConfigToConfig(test.Config)
		if err != nil && test.ExpectedError != nil {
			assert.Equal(t, test.ExpectedError, err)
			continue
		} else if err != nil && test.ExpectedError == nil {
			assert.Fail(t, "unexpected error in test", err.Error())
			continue
		}
		assert.Equal(t, test.ExpectedConfig, parsedConfig)
	}
}
