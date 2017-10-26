package bdb

import (
	"time"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/asdine/storm"
)

type EngineFileTransaction struct {
	txn  storm.Node
	root storm.Node
}

func (s *BoltBackend) NewEngineFileTransaction() (storage.EngineFileTxn, error) {
	txn, err := s.db.From(bucketEngineFiles...).Begin(true)
	if err != nil {
		return nil, convertErr(err)
	}
	return &EngineFileTransaction{txn, s.db}, nil
}

func (s *EngineFileTransaction) SaveEngineFile(sf storage.EngineFile) error {
	if sf.UploadedAt.IsZero() {
		sf.UploadedAt = time.Now().UTC()
	}

	if sf.LastUpdatedAt.IsZero() {
		sf.LastUpdatedAt = time.Now().UTC()
	}

	if sf.UploadedBy == "" {
		user, err := getUserFromNode(s.root, sf.UploadedByUUID)
		if err != nil {
			return convertErr(err)
		}
		sf.UploadedBy = user.Username
	}

	if err := s.txn.Save(&boltEngineFile{
		DocVersion: curEngineFileVer,
		EngineFile: sf,
	}); err != nil {
		return convertErr(err)
	}

	return nil
}

func (s *EngineFileTransaction) AddEntitlement(sf storage.EngineFile, userid string) error {
	node := s.txn.From(bucketEntName)
	return grantEntitlement(node, storage.User{
		UserUUID: userid,
	}, sf)
}

// Rollback any writes to the database
func (s *EngineFileTransaction) Rollback() error {
	return s.txn.Rollback()
}

// Commit the transaction
func (s *EngineFileTransaction) Commit() error {
	return s.txn.Commit()
}
