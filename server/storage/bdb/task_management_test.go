package bdb

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/mandiant/gocrack/server/storage"
	"github.com/mandiant/gocrack/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().Unix())
}

var testUnexpectedValInGetNextTestCmp = "expected next task differs from what was expected to be returned"

func createTestJobDoc(t *testing.T, txn storage.CreateTaskTxn, userDoc *storage.User) (*storage.Task, error) {
	doc := &storage.Task{
		TaskID:        uuid.NewString(),
		TaskName:      "My Awesome Task!",
		FileID:        uuid.NewString(),
		CreatedByUUID: userDoc.UserUUID,
	}
	return doc, txn.CreateTask(doc)
}

func TestTaskManagementChangeStatus(t *testing.T) {
	db := initTest(t)
	defer db.DestroyTest()

	user, err := createTestUser(false, t, db)
	if err != nil {
		assert.FailNow(t, "failed to create user document", err.Error())
	}

	txn, err := db.NewTaskCreateTransaction()
	if err != nil {
		assert.FailNow(t, "failed to create task transaction", err.Error())
	}

	doc, err := createTestJobDoc(t, txn, user)
	if err != nil {
		assert.FailNow(t, "expected document to be created successfully but failed", err.Error())
	}

	if err := txn.GrantEntitlement(user.UserUUID, *doc); err != nil {
		assert.FailNow(t, "failed to add entitlement to user who created the document", err.Error())
	}
	txn.Commit()

	err = db.ChangeTaskStatus(doc.TaskID, storage.TaskStatusDequeued, nil)
	assert.NoErrorf(t, err, "unexpected error changing task status")

	foundDoc, err := db.GetTaskByID(doc.TaskID)
	if err != nil || foundDoc == nil {
		assert.FailNow(t, "expected a document but got an error or nothing", err.Error())
	}
	assert.Equal(t, storage.TaskStatusDequeued, foundDoc.Status)
}

func TestTaskManagementCreateAndGet(t *testing.T) {
	db := initTest(t)
	defer db.DestroyTest()

	user, err := createTestUser(false, t, db)
	if err != nil {
		assert.FailNow(t, "failed to create user document", err.Error())
	}

	txn, err := db.NewTaskCreateTransaction()
	if err != nil {
		assert.FailNow(t, "failed to create task transaction", err.Error())
	}

	doc, err := createTestJobDoc(t, txn, user)
	if err != nil {
		assert.FailNow(t, "expected document to be created successfully but failed", err.Error())
	}

	if err := txn.GrantEntitlement(user.UserUUID, *doc); err != nil {
		assert.FailNow(t, "failed to add entitlement to user who created the document", err.Error())
	}
	txn.Commit()

	foundDoc, err := db.GetTaskByID(doc.TaskID)
	if err != nil || foundDoc == nil {
		assert.FailNow(t, "expected a document but got an error or nothing", err.Error())
	}

	assert.Equal(t, "My Awesome Task!", foundDoc.TaskName)
	assert.WithinDuration(t, time.Now().UTC(), foundDoc.CreatedAt, 5*time.Second)
	assert.Equal(t, user.UserUUID, foundDoc.CreatedByUUID)
}

type testGetTaskItemSearchQ struct {
	Hostname string
	Devices  storage.CLDevices
}

type testGetNextTaskItem struct {
	Tasks                      []storage.Task
	ExpectedErrorOnGetNextTask error
	// index inside "Tasks" where the GetNextTaskForHost should return this element. If ExpectedErrorOnGetNextTask is set, this is ignored
	ExpectedGetNextTask int
	Search              testGetTaskItemSearchQ
}

func TestTaskManagementGetNextTaskForHost(t *testing.T) {
	for i, test := range []testGetNextTaskItem{
		// This should return the document because the hostname matches and devices 1 & 2 are not being used when the task was requested to specifically
		// use this machine and devices 4 & 5
		{
			Tasks: []storage.Task{
				{
					FileID:            uuid.NewString(),
					TaskID:            uuid.NewString(),
					TaskName:          "Testing",
					CaseCode:          shared.GetStrPtr("CC-1337"),
					CreatedBy:         "testing",
					CreatedByUUID:     uuid.NewString(),
					CreatedAt:         time.Now().UTC(),
					AssignedToHost:    "my-hostname",
					AssignedToDevices: &storage.CLDevices{4, 5},
				},
			},
			ExpectedGetNextTask: 0,
			Search: testGetTaskItemSearchQ{
				Hostname: "my-hostname",
				Devices:  storage.CLDevices{1, 2},
			},
		},
		// expecting a storage.ErrNotFound because the device is already running a task using devices 4 & 5
		// and the next task in-line uses devices 4 & 5
		{
			Tasks: []storage.Task{
				{
					FileID:            uuid.NewString(),
					TaskID:            uuid.NewString(),
					TaskName:          "Testing",
					CaseCode:          shared.GetStrPtr("CC-1337"),
					CreatedBy:         "testing",
					CreatedByUUID:     uuid.NewString(),
					CreatedAt:         time.Now().UTC(),
					AssignedToHost:    "my-hostname",
					AssignedToDevices: &storage.CLDevices{4, 5},
				},
			},
			ExpectedErrorOnGetNextTask: storage.ErrNotFound,
			Search: testGetTaskItemSearchQ{
				Hostname: "my-hostname",
				Devices:  storage.CLDevices{4, 5},
			},
		},
		// Expecting the task to successfully return because we arent looking for a specific host match
		{
			Tasks: []storage.Task{
				{
					FileID:            uuid.NewString(),
					TaskID:            uuid.NewString(),
					TaskName:          "Testing",
					CaseCode:          shared.GetStrPtr("CC-1337"),
					CreatedBy:         "testing",
					CreatedByUUID:     uuid.NewString(),
					CreatedAt:         time.Now().UTC(),
					AssignedToDevices: &storage.CLDevices{4, 5},
				},
			},
			ExpectedGetNextTask: 0,
			Search: testGetTaskItemSearchQ{
				Hostname: "",
				Devices:  storage.CLDevices{1, 2},
			},
		},
		// Expecting the first one to return in the list as it was created before the second
		{
			Tasks: []storage.Task{
				{
					FileID:        uuid.NewString(),
					TaskID:        uuid.NewString(),
					TaskName:      "Testing",
					CaseCode:      shared.GetStrPtr("CC-1337"),
					CreatedBy:     "testing",
					CreatedByUUID: uuid.NewString(),
					CreatedAt:     time.Now().UTC().Add(-time.Duration(time.Hour * 4)),
				},
				{
					FileID:        uuid.NewString(),
					TaskID:        uuid.NewString(),
					TaskName:      "Testing2",
					CaseCode:      shared.GetStrPtr("CC-1338"),
					CreatedBy:     "testing",
					CreatedByUUID: uuid.NewString(),
					CreatedAt:     time.Now().UTC(),
				},
			},
			ExpectedGetNextTask: 0,
			Search: testGetTaskItemSearchQ{
				Hostname: "",
				Devices:  nil,
			},
		},
		// Not expecting anything to be returned as this task has "completed"
		{
			Tasks: []storage.Task{
				{
					FileID:        uuid.NewString(),
					TaskID:        uuid.NewString(),
					TaskName:      "Testing",
					CaseCode:      shared.GetStrPtr("CC-1337"),
					CreatedBy:     "testing",
					CreatedByUUID: uuid.NewString(),
					CreatedAt:     time.Now().UTC().Add(-time.Duration(time.Hour * 4)),
					Status:        storage.TaskStatusFinished,
				},
			},
			ExpectedGetNextTask:        0,
			ExpectedErrorOnGetNextTask: storage.ErrNotFound,
		},
		// Expecting the 2nd one to return first due to its priority even though it was created after the 1st one
		{
			Tasks: []storage.Task{
				{
					FileID:        uuid.NewString(),
					TaskID:        uuid.NewString(),
					TaskName:      "Testing",
					CaseCode:      shared.GetStrPtr("CC-1337"),
					CreatedByUUID: uuid.NewString(),
					CreatedBy:     "testing",
					CreatedAt:     time.Now().UTC().Add(-time.Duration(time.Hour * 4)),
					Priority:      storage.WorkerPriorityNormal,
				},
				{
					FileID:        uuid.NewString(),
					TaskID:        uuid.NewString(),
					TaskName:      "Testing2",
					CaseCode:      shared.GetStrPtr("CC-1338"),
					CreatedBy:     "testing",
					CreatedByUUID: uuid.NewString(),
					CreatedAt:     time.Now().UTC(),
					Priority:      storage.WorkerPriorityHigh,
				},
			},
			ExpectedGetNextTask: 1,
			Search: testGetTaskItemSearchQ{
				Hostname: "",
				Devices:  nil,
			},
		},
	} {
		var task *storage.Task
		db := initTest(t)

		txn, err := db.NewTaskCreateTransaction()
		if err != nil {
			assert.Fail(t, fmt.Sprintf("unexpected error creating transaction in test %d", i), err.Error())
			goto CleanupTestIteration
		}

		for taskIdx, taskToCreate := range test.Tasks {
			if err := txn.CreateTask(&taskToCreate); err != nil {
				assert.Fail(t, fmt.Sprintf("unexpected error creating task in test %d within test.Tasks[%d]", i, taskIdx), err.Error())
				goto CleanupTestIteration
			}
		}

		if err := txn.Commit(); err != nil {
			assert.Fail(t, fmt.Sprintf("unexpected error committing txn in test %d", i), err.Error())
			goto CleanupTestIteration
		}

		task, err = db.getNextTaskForHost(test.Search.Hostname, test.Search.Devices)
		if test.ExpectedErrorOnGetNextTask == nil && err != nil {
			assert.Fail(t, fmt.Sprintf("unexpected error getting next task for host in test %d", i), err.Error())
			goto CleanupTestIteration
		}

		if test.ExpectedErrorOnGetNextTask != nil {
			if err == nil {
				assert.Fail(t, fmt.Sprintf("expected error in test %d while getting next task for a host but got nil. expected error is included", i),
					test.ExpectedErrorOnGetNextTask.Error())
				goto CleanupTestIteration
			}
			assert.EqualErrorf(t, test.ExpectedErrorOnGetNextTask, err.Error(), "expected an error and got something else")
			goto CleanupTestIteration
		}

		assert.Equalf(t, test.Tasks[test.ExpectedGetNextTask].TaskName, task.TaskName, testUnexpectedValInGetNextTestCmp)
		assert.Equalf(t, test.Tasks[test.ExpectedGetNextTask].TaskID, task.TaskID, testUnexpectedValInGetNextTestCmp)
		assert.Equalf(t, test.Tasks[test.ExpectedGetNextTask].AssignedToHost, task.AssignedToHost, testUnexpectedValInGetNextTestCmp)
		assert.Equalf(t, test.Tasks[test.ExpectedGetNextTask].AssignedToHost, test.Search.Hostname, testUnexpectedValInGetNextTestCmp)

	CleanupTestIteration:
		db.DestroyTest()
		continue
	}
}

func TestTaskManagementProperlyQueueingTasks(t *testing.T) {
	db := initTest(t)
	defer db.DestroyTest()

	firstTaskID := uuid.NewString()
	tasks := []storage.Task{
		{
			FileID:            uuid.NewString(),
			TaskID:            firstTaskID,
			TaskName:          "Testing",
			CaseCode:          shared.GetStrPtr("CC-1337"),
			CreatedBy:         "testing",
			CreatedByUUID:     uuid.NewString(),
			CreatedAt:         time.Now().UTC(),
			AssignedToHost:    "my-hostname",
			AssignedToDevices: &storage.CLDevices{4, 5},
		},
		{
			FileID:            uuid.NewString(),
			TaskID:            uuid.NewString(),
			TaskName:          "Testing 2",
			CaseCode:          shared.GetStrPtr("CC-1337"),
			CreatedBy:         "testing",
			CreatedByUUID:     uuid.NewString(),
			CreatedAt:         time.Now().UTC().Add(1 * time.Minute),
			AssignedToDevices: &storage.CLDevices{5},
		},
	}

	for _, task := range tasks {
		txn, err := db.NewTaskCreateTransaction()
		if err != nil {
			assert.Nil(t, err)
			return
		}

		if err := txn.CreateTask(&task); err != nil {
			assert.Nil(t, err)
			return
		}

		if err := txn.Commit(); err != nil {
			assert.Nil(t, err)
			return
		}
	}

	task, err := db.getNextTaskForHost("my-hostname", nil)
	if err != nil {
		assert.Nil(t, err, "an error should not be present here")
		return
	}

	assert.NotNil(t, task)
	assert.Equal(t, firstTaskID, task.TaskID)

	// This should return nothing as the devices for the 2nd task are "in-use"
	task, err = db.getNextTaskForHost("my-hostname", storage.CLDevices{4, 5})
	assert.Equal(t, err, storage.ErrNotFound)
	assert.Nil(t, task)
}

func TestDeleteTask(t *testing.T) {
	db := initTest(t)
	defer db.DestroyTest()

	doc := storage.Task{
		FileID:            uuid.NewString(),
		TaskID:            uuid.NewString(),
		TaskName:          "Testing",
		CaseCode:          shared.GetStrPtr("CC-1337"),
		CreatedBy:         "testing",
		CreatedByUUID:     uuid.NewString(),
		CreatedAt:         time.Now().UTC(),
		AssignedToHost:    "my-hostname",
		AssignedToDevices: &storage.CLDevices{4, 5},
	}

	txn, err := db.NewTaskCreateTransaction()
	if err != nil {
		assert.Nil(t, err)
		return
	}

	if err := txn.CreateTask(&doc); err != nil {
		assert.Nil(t, err)
		return
	}

	if err := txn.Commit(); err != nil {
		assert.Nil(t, err)
		return
	}

	if err := db.DeleteTask(doc.TaskID); err != nil {
		assert.Nil(t, err)
		return
	}

}
