package bdb

import (
	"testing"
	"time"

	"github.com/fireeye/gocrack/server/storage"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestEntitlement(t *testing.T) {
	db := initTest(t)
	defer db.DestroyTest()

	dummyUUID := uuid.NewV4().String()
	taskFile := storage.TaskFile{
		FileID:         dummyUUID,
		SavedAt:        "/tmp/foobaz",
		UploadedAt:     time.Now(),
		UploadedBy:     "testing",
		UploadedByUUID: uuid.NewV4().String(),
	}

	txn, err := db.NewTaskFileTransaction()
	if err != nil {
		assert.FailNow(t, "could not start task file txn", err)
	}

	if err := txn.SaveTaskFile(taskFile); err != nil {
		txn.Rollback()
		assert.FailNow(t, "could not save record", err)
	}
	txn.Commit()

	// We don't actually need to save this record
	dummyUserRec := storage.User{
		UserUUID: uuid.NewV4().String(),
	}

	// User shouldnt be entitled to this document...yet
	hasAccess, err := db.CheckEntitlement(dummyUserRec.UserUUID, taskFile.FileID, storage.EntitlementTaskFile)
	assert.Nil(t, err)
	assert.False(t, hasAccess)

	err = db.GrantEntitlement(dummyUserRec, taskFile)
	if err != nil {
		assert.FailNow(t, "expected a successful entitlement grant", err)
	}

	hasAccess, err = db.CheckEntitlement(dummyUserRec.UserUUID, taskFile.FileID, storage.EntitlementTaskFile)
	assert.Nil(t, err)
	assert.True(t, hasAccess)

	// Now let's revoke it!
	err = db.RevokeEntitlement(dummyUserRec, taskFile)
	assert.Nil(t, err)

	hasAccess, err = db.CheckEntitlement(dummyUserRec.UserUUID, taskFile.FileID, storage.EntitlementTaskFile)
	assert.Nil(t, err)
	assert.False(t, hasAccess)
}
