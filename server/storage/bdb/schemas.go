package bdb

import "github.com/mandiant/gocrack/server/storage"

const (
	curCrackTaskVer      float32 = 1.1
	curUserVer           float32 = 1.0
	curEntVer            float32 = 1.0
	curTaskFileVer       float32 = 1.0
	curCrackedHashVer    float32 = 1.0
	curAuditEntryVer     float32 = 1.0
	curEngineFileVer     float32 = 1.0
	curCheckpointFileVer float32 = 1.0
)

var (
	bucketAuditLog    = "audit_log"
	bucketTasks       = "tasks"
	bucketEntName     = "entitlements"
	bucketCheckpoints = "checkpoints"

	bucketTaskFiles   = []string{"files", "task_files"}
	bucketEngineFiles = []string{"files", "engine_files"}

	// Buckets related to entitlements
	bucketEntTaskFiles   = append(bucketTaskFiles, bucketEntName)
	bucketEntTasks       = []string{bucketTasks, bucketEntName}
	bucketEntEngineFiles = append(bucketEngineFiles, bucketEntName)
)

type boltUser struct {
	ID           int64 `storm:"id,increment"`
	DocVersion   float32
	storage.User `storm:"inline"`
}

type boltTaskFile struct {
	ID               int64 `storm:"id,increment"`
	DocVersion       float32
	storage.TaskFile `storm:"inline"`
}

type boltEntitlement struct {
	ID int64 `storm:"id,increment"`
	// UniqueID is the MD5 of UserUUID and EntitleID in storage.EntitlementEntry
	// and is used as a unique check due to uniqueness limitations in bolt/storm
	UniqueID                 string `storm:"unique"`
	DocVersion               float32
	storage.EntitlementEntry `storm:"inline"`
}

type boltCrackTask struct {
	DocVersion   float32
	storage.Task `storm:"inline"`
}

type boltCrackedHash struct {
	ID                  int64 `storm:"id,increment"`
	DocVersion          float32
	storage.CrackedHash `storm:"inline"`
}

type boltAuditLogEntry struct {
	ID                       int64 `storm:"id,increment"`
	DocVersion               float32
	storage.ActivityLogEntry `storm:"inline"`
}

type boltEngineFile struct {
	ID                 int64 `storm:"id,increment"`
	DocVersion         float32
	storage.EngineFile `storm:"inline"`
}

type boltCheckpointFile struct {
	ID                     string `storm:"id,unique"`
	DocVersion             float32
	storage.CheckpointFile `storm:"inline"`
}
