package bdb

import (
	"errors"

	"github.com/mandiant/gocrack/server/storage"

	"github.com/asdine/storm"
)

type deleteType int

const (
	deleteTaskFile deleteType = 1 << iota
	deleteTaskEngineFile
)

func convertErr(err error) error {
	if err == nil {
		return nil
	}

	switch err {
	case storm.ErrNotFound:
		return storage.ErrNotFound
	case storm.ErrAlreadyExists:
		return storage.ErrAlreadyExists
	}

	return storage.StorageError{DriverError: err}
}

func (s *BoltBackend) deleteFile(fileid string, filetype deleteType) error {
	switch filetype {
	case deleteTaskFile:
		var btf boltTaskFile
		if err := s.db.From(bucketTaskFiles...).One("FileID", fileid, &btf); err != nil {
			return err
		}

		return s.db.From(bucketTaskFiles...).DeleteStruct(&btf)
	case deleteTaskEngineFile:
		var bsf boltEngineFile
		if err := s.db.From(bucketEngineFiles...).One("FileID", fileid, &bsf); err != nil {
			return err
		}

		return s.db.From(bucketEngineFiles...).DeleteStruct(&bsf)
	}
	return errors.New("unknown filetype")
}
