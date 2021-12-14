package bdb

import (
	"net/http"
	"testing"
	"time"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func generateRandomAuditEntry(entID string, userID string) storage.ActivityLogEntry {
	if userID == "" {
		userID = uuid.NewString()
	}

	return storage.ActivityLogEntry{
		OccuredAt:  time.Now(),
		UserUUID:   userID,
		EntityID:   entID,
		StatusCode: http.StatusOK,
		Path:       "/testing/path/1234/",
		IPAddress:  "13.37.13.37",
	}
}

func TestLogActivity(t *testing.T) {
	db := initTest(t)
	defer db.DestroyTest()

	user, err := createTestUser(false, t, db)
	if err != nil {
		assert.FailNow(t, "unexpected error in createTestUser", err.Error())
	}

	err = db.LogActivity(generateRandomAuditEntry("some-entity-id", user.UserUUID))
	assert.Nil(t, err)

	entries, err := db.GetActivityLog("some-entity-id")
	if err != nil {
		assert.FailNow(t, "unexpected error in GetActivityLog", err.Error())
	}
	assert.Len(t, entries, 1)
}

func TestRemoveActivityEntries(t *testing.T) {
	db := initTest(t)
	defer db.DestroyTest()

	user, err := createTestUser(false, t, db)
	if err != nil {
		assert.FailNow(t, "unexpected error in createTestUser", err.Error())
	}

	// Seed some entries that we're going to delete
	for i := 0; i < 25; i++ {
		db.LogActivity(generateRandomAuditEntry("some-entity-id-that-we-will-delete", user.UserUUID))
	}

	// Seed some entries that we're not going to delete
	for i := 0; i < 25; i++ {
		db.LogActivity(generateRandomAuditEntry("some-entity-id-that-will-remain", user.UserUUID))
	}

	entries, err := db.GetActivityLog("some-entity-id-that-we-will-delete")
	if err != nil {
		assert.FailNow(t, "unexpected error in GetActivityLog", err.Error())
	}
	assert.Len(t, entries, 25)

	if err := db.RemoveActivityEntries("some-entity-id-that-we-will-delete"); err != nil {
		assert.FailNow(t, "unexpected error in GetActivityLog", err.Error())
	}

	entries, err = db.GetActivityLog("some-entity-id-that-we-will-delete")
	if err != nil {
		assert.FailNow(t, "unexpected error in GetActivityLog", err.Error())
	}
	assert.Len(t, entries, 0)

	entries, err = db.GetActivityLog("some-entity-id-that-will-remain")
	if err != nil {
		assert.FailNow(t, "unexpected error in GetActivityLog", err.Error())
	}
	assert.Len(t, entries, 25)
}
