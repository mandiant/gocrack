package bdb

import (
	"fmt"
	"testing"
	"time"

	"github.com/mandiant/gocrack/server/storage"

	"github.com/stretchr/testify/assert"
)

type testTaskFile struct {
	UploadedUser              *storage.User
	QueryUser                 *storage.User
	TaskFile                  storage.TaskFile
	GrantQueryUserEntitlement bool
}

func TestCreatingTaskFilePermission(t *testing.T) {
	var err error
	var txn storage.TaskFileTxn
	var found []storage.TaskFile

	for i, test := range []testTaskFile{
		// Unprivileged user (user1) successfully uploads a file and can verify it can see it
		{
			UploadedUser: &storage.User{
				Username:    "user1",
				UserUUID:    "user1_uuid",
				IsSuperUser: false,
			},
			QueryUser: nil,
			TaskFile: storage.TaskFile{
				UploadedByUUID: "user1_uuid",
				UploadedAt:     time.Now().UTC(),
				FileSize:       1337,
				FileName:       "user1's awesome file",
				SHA1Hash:       "e53b01ec18dfb13cac90ed8b4f802a3a",
			},
		},
		// Unprivileged user (user1) successfully uploads a file and another unprivileged user (user2) cant see it
		{
			UploadedUser: &storage.User{
				Username:    "user1",
				UserUUID:    "user1_uuid",
				IsSuperUser: false,
			},
			QueryUser: &storage.User{
				Username:    "user2",
				UserUUID:    "user2_uuid",
				IsSuperUser: false,
			},
			TaskFile: storage.TaskFile{
				UploadedByUUID: "user1_uuid",
				UploadedAt:     time.Now().UTC(),
				FileSize:       1337,
				FileName:       "user1's awesome file",
				SHA1Hash:       "e53b01ec18dfb13cac90ed8b4f802a3a",
			},
		},
		// Unprivileged (user1) user uploads a file and admin (user2) verifies he can see it
		{
			UploadedUser: &storage.User{
				Username:    "user1",
				UserUUID:    "user1_uuid",
				IsSuperUser: false,
			},
			QueryUser: &storage.User{
				Username:    "user2",
				UserUUID:    "user2_uuid",
				IsSuperUser: true,
			},
			TaskFile: storage.TaskFile{
				UploadedByUUID: "user1_uuid",
				UploadedAt:     time.Now().UTC(),
				FileSize:       1337,
				FileName:       "user1's awesome file",
				SHA1Hash:       "e53b01ec18dfb13cac90ed8b4f802a3a",
			},
		},
		// Admin (user1) uploads a file and grants a basic unprivileged user (user2) access to the file
		{
			UploadedUser: &storage.User{
				Username:    "user1",
				UserUUID:    "user1_uuid",
				IsSuperUser: true,
			},
			QueryUser: &storage.User{
				Username:    "user2",
				UserUUID:    "user2_uuid",
				IsSuperUser: false,
			},
			GrantQueryUserEntitlement: true,
			TaskFile: storage.TaskFile{
				UploadedByUUID: "user1_uuid",
				UploadedAt:     time.Now().UTC(),
				FileSize:       1337,
				FileName:       "user1's awesome file",
				SHA1Hash:       "e53b01ec18dfb13cac90ed8b4f802a3a",
			},
		},
	} {
		db := initTest(t)
		if err := db.CreateUser(test.UploadedUser); err != nil {
			assert.Fail(t, fmt.Sprintf("could not create upload user in test %d", i), err.Error())
			goto CleanupTest
		}

		if test.QueryUser != nil {
			if err := db.CreateUser(test.QueryUser); err != nil {
				assert.Fail(t, fmt.Sprintf("could not create query user in test %d", i), err.Error())
				goto CleanupTest
			}
		}

		if txn, err = db.NewTaskFileTransaction(); err != nil {
			assert.Fail(t, fmt.Sprintf("unexpected error creating a transaction in test %d", i), err.Error())
			goto CleanupTest
		}

		if err = txn.SaveTaskFile(test.TaskFile); err != nil {
			assert.Fail(t, fmt.Sprintf("failed to create task file in test %d", i), err.Error())
			goto CleanupTest
		}

		if err = txn.AddEntitlement(test.TaskFile, test.UploadedUser.UserUUID); err != nil {
			assert.Fail(t, fmt.Sprintf("failed to add entitlement to the user who uploaded the file %d", i), err.Error())
			goto CleanupTest
		}

		if test.GrantQueryUserEntitlement {
			if err = txn.AddEntitlement(test.TaskFile, test.QueryUser.UserUUID); err != nil {
				assert.Fail(t, fmt.Sprintf("failed to add entitlement to the query user in test %d", i), err.Error())
				goto CleanupTest
			}
		}

		// Alright, we should be set to write these changes and do some basic validation
		txn.Commit()

		// query to make sure our user who uploaded it can see it
		if found, err = db.ListTasksForUser(*test.UploadedUser); err != nil {
			assert.Fail(t, fmt.Sprintf("failed to list tasks file for user who uploaded the file in test %d", i), err.Error())
			goto CleanupTest
		}
		assert.Len(t, found, 1)

		if test.QueryUser != nil {
			// query to make sure our user who uploaded it can see it
			if found, err = db.ListTasksForUser(*test.QueryUser); err != nil {
				assert.Fail(t, fmt.Sprintf("failed to list tasks file for a normal user %d", i), err.Error())
				goto CleanupTest
			}

			if test.QueryUser.IsSuperUser || test.GrantQueryUserEntitlement {
				assert.Len(t, found, 1)
			} else {
				assert.Len(t, found, 0)
			}
		}

	CleanupTest:
		if err != nil && txn != nil {
			fmt.Println("attempting rollback", err)
			if err := txn.Rollback(); err != nil {
				assert.FailNow(t, "failed to rollback on error", err.Error())
			}
		}
		db.DestroyTest()
	}
}
