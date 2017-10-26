package bdb

import (
	"github.com/fireeye/gocrack/server/storage"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

// TaskFileTransaction is used in the creation of a task file and all the APIs it defines are executed under a bolt transaction
type TaskFileTransaction struct {
	txn  storm.Node
	root storm.Node
}

// NewTaskFileTransaction creates a new transaction used by the task file PUT APIs
func (s *BoltBackend) NewTaskFileTransaction() (storage.TaskFileTxn, error) {
	txn, err := s.db.From(bucketTaskFiles...).Begin(true)
	if err != nil {
		return nil, convertErr(err)
	}
	return &TaskFileTransaction{txn, s.db}, nil
}

// SaveTaskFile saves metadata regarding a task file to the database
func (s *TaskFileTransaction) SaveTaskFile(tf storage.TaskFile) error {
	if tf.UploadedBy == "" {
		user, err := getUserFromNode(s.root, tf.UploadedByUUID)
		if err != nil {
			return convertErr(err)
		}
		tf.UploadedBy = user.Username
	}

	if err := s.txn.Save(&boltTaskFile{
		TaskFile:   tf,
		DocVersion: curTaskFileVer,
	}); err != nil {
		return convertErr(err)
	}
	return nil
}

// AddEntitlement creates a record giving the user access to the task file
func (s *TaskFileTransaction) AddEntitlement(tf storage.TaskFile, userid string) error {
	node := s.txn.From(bucketEntName)
	return grantEntitlement(node, storage.User{
		UserUUID: userid,
	}, tf)
}

// Rollback any writes to the database
func (s *TaskFileTransaction) Rollback() error {
	return s.txn.Rollback()
}

// Commit the transaction
func (s *TaskFileTransaction) Commit() error {
	return s.txn.Commit()
}

// GetTaskFileByID implements the storage.GetTaskFileByID API
func (s *BoltBackend) GetTaskFileByID(storageID string) (*storage.TaskFile, error) {
	var btf boltTaskFile

	if err := s.db.From(bucketTaskFiles...).One("FileID", storageID, &btf); err != nil {
		return nil, convertErr(err)
	}

	out := storage.TaskFile(btf.TaskFile)
	return &out, nil
}

// ListTasksForUser implements the storage.ListTasksForUser API
func (s *BoltBackend) ListTasksForUser(user storage.User) ([]storage.TaskFile, error) {
	node := s.db.From(bucketTaskFiles...)
	bq := node.Select()

	// Build a cache of all the FileID's that the user is entitled to aid in
	// getting a list of all tasks the user has access to if they are not admin
	if !user.IsSuperUser {
		var fileids []string
		if err := node.From(bucketEntName).Select(
			q.Eq("UserUUID", user.UserUUID),
		).Each(new(boltEntitlement), func(record interface{}) error {
			be := record.(*boltEntitlement)
			fileids = append(fileids, be.EntitledID)
			// returning an error here stops the each callback
			return nil
		}); err != nil {
			return nil, convertErr(err)
		}
		bq = node.Select(q.In("FileID", fileids))
	}

	bq = bq.OrderBy("UploadedAt").Reverse()
	tfs := make([]storage.TaskFile, 0)
	if err := bq.Each(new(boltTaskFile), func(record interface{}) error {
		boltTaskFile := record.(*boltTaskFile)
		tfs = append(tfs, storage.TaskFile(boltTaskFile.TaskFile))
		return nil
	}); err != nil {
		return nil, convertErr(err)
	}
	return tfs, nil
}

// DeleteTaskFile implements storage.DeleteTaskFile
func (s *BoltBackend) DeleteTaskFile(fileID string) error {
	return convertErr(s.deleteFile(fileID, deleteTaskFile))
}
