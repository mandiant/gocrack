package test

import "github.com/mandiant/gocrack/server/storage"

type TestEngineFileTxn struct {
	doc   *storage.EngineFile
	saved bool
}

func (s *TestEngineFileTxn) SaveEngineFile(ef storage.EngineFile) error {
	s.doc = &ef
	s.saved = false
	return nil
}

func (s *TestEngineFileTxn) AddEntitlement(ef storage.EngineFile, userID string) error {
	return nil
}

func (s *TestEngineFileTxn) Rollback() error {
	s.doc = nil
	s.saved = false
	return nil
}

func (s *TestEngineFileTxn) Commit() error {
	s.saved = true
	return nil
}

type TestStorageImpl struct{}

func (s *TestStorageImpl) NewEngineFileTransaction() (storage.EngineFileTxn, error) {
	return &TestEngineFileTxn{}, nil
}

func (s *TestStorageImpl) DeleteEngineFile(fileid string) error {
	return nil
}

func (s *TestStorageImpl) DeleteTaskFile(fileid string) error {
	return nil
}
