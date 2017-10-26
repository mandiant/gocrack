package bdb

import (
	"github.com/fireeye/gocrack/server/storage"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

// GetEngineFileByID returns the engine file given it's unique ID
func (s *BoltBackend) GetEngineFileByID(storageID string) (*storage.EngineFile, error) {
	var bsf boltEngineFile
	if err := s.db.From(bucketEngineFiles...).One("FileID", storageID, &bsf); err != nil {
		return nil, convertErr(err)
	}
	sf := storage.EngineFile(bsf.EngineFile)
	return &sf, nil
}

// GetEngineFilesForUser returns a list of engine files available to the user
func (s *BoltBackend) GetEngineFilesForUser(user storage.User) (sfs []storage.EngineFile, err error) {
	node := s.db.From(bucketEngineFiles...)
	var fileids []string
	var baseQuery storm.Query

	if !user.IsSuperUser {
		if err := node.From(bucketEntName).Select(
			q.Eq("UserUUID", user.UserUUID),
		).Each(new(boltEntitlement), func(record interface{}) error {
			be := record.(*boltEntitlement)
			fileids = append(fileids, be.EntitledID)
			return nil
		}); err != nil {
			return nil, err
		}
		baseQuery = node.Select(q.Or(
			q.Eq("IsShared", true),
			q.In("FileID", fileids),
		))
	} else {
		baseQuery = node.Select()
	}

	if err = convertErr(baseQuery.Each(new(boltEngineFile), func(record interface{}) error {
		sfs = append(sfs, storage.EngineFile(record.(*boltEngineFile).EngineFile))
		return nil
	})); err != nil {
		return nil, err
	}
	return
}

// DeleteEngineFile file implements storage.DeleteEngineFile
func (s *BoltBackend) DeleteEngineFile(fileID string) error {
	return s.deleteFile(fileID, deleteTaskEngineFile)
}
