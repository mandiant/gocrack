package bdb

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

func getEntitlement(node storm.Node, userUUID, entID string) (bet *boltEntitlement, err error) {
	var tmp boltEntitlement

	if err = node.Select(
		q.Eq("UserUUID", userUUID),
		q.Eq("EntitledID", entID),
	).Limit(1).First(&tmp); err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}
		return nil, convertErr(err)
	}

	return &tmp, nil
}

// grantEntitlement builds the entitlement document and inserts it into the database
func grantEntitlement(node storm.Node, user storage.User, entitledTo interface{}) error {
	var entitledID string

	// Build the proper entitlement doc based on the input of entitledTo
	switch rec := entitledTo.(type) {
	case storage.TaskFile:
		entitledID = rec.FileID
	case storage.Task:
		entitledID = rec.TaskID
	case storage.EngineFile:
		entitledID = rec.FileID
	default:
		return fmt.Errorf("unknown object type passed into entitledTo")
	}

	if err := node.Save(&boltEntitlement{
		EntitlementEntry: storage.EntitlementEntry{
			UserUUID:        user.UserUUID,
			EntitledID:      entitledID,
			GrantedAccessAt: time.Now().UTC(),
		},
		UniqueID:   fmt.Sprintf("%x", md5.New().Sum([]byte(fmt.Sprintf("%s_%s", user.UserUUID, entitledID)))),
		DocVersion: curEntVer,
	}); err != nil {
		// If they are already entitled, do not unnecessarily create another record
		if err == storm.ErrAlreadyExists {
			return nil
		}
		return convertErr(err)
	}
	return nil
}

// CheckEntitlement implements storage.CheckEntitlement
func (s *BoltBackend) CheckEntitlement(userUUID, entityID string, entType storage.EntitlementType) (bool, error) {
	var node storm.Node
	switch entType {
	case storage.EntitlementTaskFile:
		node = s.db.From(bucketEntTaskFiles...)
	case storage.EntitlementTask:
		node = s.db.From(bucketEntTasks...)
	case storage.EntitlementEngineFile:
		node = s.db.From(bucketEntEngineFiles...)
	default:
		return false, fmt.Errorf("unknown entType of %d", entType)
	}

	if r, err := getEntitlement(node, userUUID, entityID); r == nil || err != nil {
		return false, convertErr(err)
	}

	return true, nil
}

// GrantEntitlement implements storage.GrantEntitlement
func (s *BoltBackend) GrantEntitlement(user storage.User, entitledTo interface{}) (err error) {
	if user.UserUUID == "" {
		return fmt.Errorf("user record must have a UserUUID to check entitlement. is %s", user.UserUUID)
	}

	var node storm.Node
	// determine the proper sub-bucket based on entitledTo
	switch entitledTo.(type) {
	case storage.TaskFile:
		node = s.db.From(bucketEntTaskFiles...)
	case storage.Task:
		node = s.db.From(bucketEntTasks...)
	case storage.EngineFile:
		node = s.db.From(bucketEntEngineFiles...)
	default:
		err = fmt.Errorf("unknown object type passed into GrantEntitlement")
	}

	if err != nil {
		return
	}
	// XXX(cschmitt): Do we want to check if the user exists? there's really no point...
	return grantEntitlement(node, user, entitledTo)
}

// RevokeEntitlement removes the users access to the document
func (s *BoltBackend) RevokeEntitlement(user storage.User, document interface{}) error {
	if user.UserUUID == "" {
		return fmt.Errorf("user record must have a UserUUID to check entitlement. is %s", user.UserUUID)
	}

	var node storm.Node
	var entitledID string

	switch rec := document.(type) {
	case storage.TaskFile:
		node = s.db.From(bucketEntTaskFiles...)
		entitledID = rec.FileID
	case storage.Task:
		node = s.db.From(bucketEntTasks...)
		entitledID = rec.TaskID
	case storage.EngineFile:
		node = s.db.From(bucketEntEngineFiles...)
		entitledID = rec.FileID
	default:
		return fmt.Errorf("unknown object type passed into entitledTo")
	}

	r, err := getEntitlement(node, user.UserUUID, entitledID)
	if r == nil || err != nil {
		return convertErr(err)
	}

	if err = node.DeleteStruct(r); err != nil {
		return convertErr(err)
	}

	return nil
}

// GetEntitlementsForTask implements storage.GetEntitlementsForTask
func (s *BoltBackend) GetEntitlementsForTask(entityID string) ([]storage.EntitlementEntry, error) {
	var ents []storage.EntitlementEntry

	if err := s.db.From(bucketEntTasks...).Select(
		q.Eq("EntitledID", entityID),
	).Each(new(boltEntitlement), func(record interface{}) error {
		be := record.(*boltEntitlement)
		ents = append(ents, storage.EntitlementEntry{
			UserUUID:        be.UserUUID,
			EntitledID:      be.EntitledID,
			GrantedAccessAt: be.GrantedAccessAt,
		})
		return nil
	}); err != nil {
		return nil, convertErr(err)
	}
	return ents, nil
}

// RemoveEntitlements implements storage.RemoveEntitlements
func (s *BoltBackend) RemoveEntitlements(entityID string, entType storage.EntitlementType) error {
	var node storm.Node

	switch entType {
	case storage.EntitlementTaskFile:
		node = s.db.From(bucketEntTaskFiles...)
	case storage.EntitlementTask:
		node = s.db.From(bucketEntTasks...)
	case storage.EntitlementEngineFile:
		node = s.db.From(bucketEntEngineFiles...)
	default:
		return fmt.Errorf("unknown entType of %d", entType)
	}

	return convertErr(node.Select(q.Eq("EntitledID", entityID)).Each(new(boltEntitlement), func(record interface{}) error {
		return node.DeleteStruct(record.(*boltEntitlement))
	}))
}
