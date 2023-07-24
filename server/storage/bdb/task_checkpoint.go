package bdb

import "github.com/mandiant/gocrack/server/storage"

// SaveTaskCheckpoint implements storage.SaveTaskCheckpoint
func (s *BoltBackend) SaveTaskCheckpoint(checkpoint storage.CheckpointFile) error {
	return s.db.From(bucketCheckpoints).Save(&boltCheckpointFile{
		ID:             checkpoint.TaskID,
		DocVersion:     curCheckpointFileVer,
		CheckpointFile: checkpoint,
	})
}

// GetTaskCheckpoint implements storage.GetTaskCheckpoint
func (s *BoltBackend) GetTaskCheckpoint(taskid string) ([]byte, error) {
	var cf boltCheckpointFile
	if err := s.db.From(bucketCheckpoints).One("ID", taskid, &cf); err != nil {
		return nil, convertErr(err)
	}
	return cf.Data, nil
}
