package bdb

import (
	"time"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/asdine/storm"
)

type TaskCreateTransaction struct {
	txn  storm.Node
	root storm.Node
}

// NewTaskCreateTransaction creates a new transaction used by the task creation PUT API
func (s *BoltBackend) NewTaskCreateTransaction() (storage.CreateTaskTxn, error) {
	txn, err := s.db.From(bucketTasks).Begin(true)
	if err != nil {
		return nil, err
	}
	return &TaskCreateTransaction{txn, s.db}, nil
}

// CreateTask creates the task within the context of a database transaction
func (s *TaskCreateTransaction) CreateTask(t *storage.Task) error {
	if t.CreatedByUUID == "" {
		return errExpectedUser
	}

	if t.CreatedBy == "" {
		user, err := getUserFromNode(s.root, t.CreatedByUUID)
		if err != nil {
			return err
		}
		t.CreatedBy = user.Username
	}

	if t.Status == "" {
		t.Status = storage.TaskStatusQueued
	}

	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}

	if t.LastUpdatedAt.IsZero() {
		t.LastUpdatedAt = time.Now().UTC()
	}

	if err := s.txn.Save(&boltCrackTask{Task: *t, DocVersion: curCrackTaskVer}); err != nil {
		return convertErr(err)
	}

	if err := s.GrantEntitlement(t.CreatedByUUID, *t); err != nil {
		convertErr(err)
	}
	return nil
}

// GrantEntitlement grants additional users to the task within the context of a database transaction
func (s *TaskCreateTransaction) GrantEntitlement(userUUID string, t storage.Task) error {
	return grantEntitlement(s.txn.From(bucketEntName), storage.User{
		UserUUID: userUUID,
	}, t)
}

// Rollback any writes to the database
func (s *TaskCreateTransaction) Rollback() error {
	return s.txn.Rollback()
}

// Commit the transaction
func (s *TaskCreateTransaction) Commit() error {
	return s.txn.Commit()
}
