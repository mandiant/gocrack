package bdb

import (
	"github.com/fireeye/gocrack/server/storage"

	"github.com/asdine/storm/q"
)

// LogActivity implements storage.LogActivity
func (s *BoltBackend) LogActivity(entry storage.ActivityLogEntry) error {
	if entry.Username == "" {
		user, err := getUserFromNode(s.db, entry.UserUUID)
		if err != nil {
			return err
		}
		entry.Username = user.Username
	}

	if err := s.db.From(bucketAuditLog).Save(&boltAuditLogEntry{
		ActivityLogEntry: entry,
		DocVersion:       curAuditEntryVer,
	}); err != nil {
		return convertErr(err)
	}
	return nil
}

// GetActivityLog implements storage.GetActivityLog
func (s *BoltBackend) GetActivityLog(entityID string) ([]storage.ActivityLogEntry, error) {
	items := make([]storage.ActivityLogEntry, 0)

	if err := convertErr(s.db.From(bucketAuditLog).Select(q.Eq("EntityID", entityID)).OrderBy("OccuredAt").Each(new(boltAuditLogEntry), func(record interface{}) error {
		entry := record.(*boltAuditLogEntry)
		items = append(items, storage.ActivityLogEntry(entry.ActivityLogEntry))
		return nil
	})); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return items, nil
}

// RemoveActivityEntries implements storage.RemoveActivityEntries
func (s *BoltBackend) RemoveActivityEntries(entityID string) error {
	var entries []*boltAuditLogEntry

	if err := s.db.From(bucketAuditLog).Select(q.Eq("EntityID", entityID)).Each(new(boltAuditLogEntry), func(record interface{}) error {
		entries = append(entries, record.(*boltAuditLogEntry))
		return nil
	}); err != nil {
		return convertErr(err)
	}

	for _, ent := range entries {
		s.db.From(bucketAuditLog).Remove(ent)
	}
	return nil
}
