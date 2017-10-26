package bdb

import (
	"testing"
	"time"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/stretchr/testify/assert"
)

func createTestUser(isSuperUser bool, t *testing.T, db *storageTester) (*storage.User, error) {
	doc := &storage.User{
		Username:     "testuser",
		Password:     "secret_password!",
		EmailAddress: "dummyuser@fireeye.com",
		CreatedAt:    time.Now(),
		IsSuperUser:  isSuperUser,
	}
	return doc, db.CreateUser(doc)
}

func TestUserAPIs(t *testing.T) {
	db := initTest(t)
	defer db.DestroyTest()

	testPass := "my_super_secret_password!1"
	err := db.CreateUser(&storage.User{
		Username: "testuser",
		Password: testPass,
	})
	assert.Nil(t, err)

	rec, err := db.SearchForUserByPassword("testuser", func(p string) bool {
		if p == testPass {
			return true
		}
		return false
	})
	if err != nil {
		assert.FailNow(t, "expected to find a user record", err)
	}

	assert.Equal(t, "testuser", rec.Username)
	assert.Equal(t, testPass, rec.Password)
}
