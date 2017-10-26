package storage

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// WorkerCrackEngine defines the engine the worker uses to crack the password(s)
type WorkerCrackEngine uint8

const (
	// WorkerHashcatEngine indicates the task should use the hashcat engine
	WorkerHashcatEngine WorkerCrackEngine = 1 << iota
)

// PendingTaskPayloadType describes the payload structure & contents
type PendingTaskPayloadType uint8

const (
	// PendingTaskNewRequest indicates a new task
	PendingTaskNewRequest PendingTaskPayloadType = 1 << iota
	// PendingTaskStatusChange indicates a change in task status
	PendingTaskStatusChange
)

// TaskFileEngine indicates the engine that this task file is for
type TaskFileEngine uint8

const (
	// TaskFileEngineAll indicates that this file should work on all engines
	TaskFileEngineAll = 0
	// TaskFileEngineHashcat indicates that this task file only works on the hashcat engine
	TaskFileEngineHashcat = TaskFileEngine(WorkerHashcatEngine)
)

// ActivityType describes an action taken within the system
type ActivityType uint8

const (
	// ActivtyLogin indicates a logon activity to the system
	ActivtyLogin ActivityType = iota
	// ActivityCreatedTask indicates a task was created to the system
	ActivityCreatedTask
	// ActivityModifiedTask indicates a task was modified in the system
	ActivityModifiedTask
	// ActivityDeletedTask indicates a task was deleted in the system
	ActivityDeletedTask
	// ActivityViewTask indicates a task was viewed in the system
	ActivityViewTask
	// ActivityViewPasswords indicates passwords were viewed
	ActivityViewPasswords
	// ActivityEntitlementRequest indicates an entitled request was requsted in the system
	ActivityEntitlementRequest
	// ActivityEntitlementModification indicates a user attempted to modify an entitled entity in the system
	ActivityEntitlementModification
)

// EngineFileType indicates the type of engine file
type EngineFileType uint8

const (
	// EngineFileDictionary indicates the shared file is a list of dictionary words
	EngineFileDictionary EngineFileType = iota
	// EngineFileMasks indicates the file is a list of passwords masks (combinations to try)
	EngineFileMasks
	// EngineFileRules indicates the file is a mangling rule set and is used to modify dictionary words
	EngineFileRules
)

// TaskStatus indicates the processing status of a Task
type TaskStatus string

const (
	TaskStatusQueued    TaskStatus = "Queued"
	TaskStatusDequeued  TaskStatus = "Dequeued"
	TaskStatusRunning   TaskStatus = "Running"
	TaskStatusStopping  TaskStatus = "Stopping"
	TaskStatusStopped   TaskStatus = "Stopped"
	TaskStatusError     TaskStatus = "Error"
	TaskStatusExhausted TaskStatus = "Exhausted"
	TaskStatusFinished  TaskStatus = "Finished"
)

func (s *TaskStatus) UnmarshalJSON(data []byte) error {
	var tmp string
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	switch strings.ToLower(tmp) {
	case "queued":
		*s = TaskStatusQueued
	case "dequeued":
		*s = TaskStatusDequeued
	case "running":
		*s = TaskStatusRunning
	case "stopping":
		*s = TaskStatusStopping
	case "stopped":
		*s = TaskStatusStopped
	case "error":
		*s = TaskStatusError
	case "exhausted":
		*s = TaskStatusExhausted
	case "finished":
		*s = TaskStatusFinished
	default:
		return fmt.Errorf("`%s` is not a valid task status", tmp)
	}
	return nil
}

// WorkerPriority describes the priority of the task in relative to the position in queue
type WorkerPriority int

const (
	WorkerPriorityHigh WorkerPriority = iota
	WorkerPriorityNormal
	WorkerPriorityLow
)

// EntitlementType indicates the document type for the entitlement record
type EntitlementType uint8

const (
	EntitlementTask EntitlementType = iota
	EntitlementTaskFile
	EntitlementEngineFile
)

// TaskFile describes all the properties of a task file which is a file that contains one or more hashes
// that can be cracked by a GoCrack engine
type TaskFile struct {
	FileID            string `storm:"unique"`
	SavedAt           string // The physical location on the server where the file is located
	UploadedAt        time.Time
	UploadedBy        string
	UploadedByUUID    string // UUID of the user who initially uploaded the file
	FileSize          int64
	FileName          string
	SHA1Hash          string
	ForEngine         TaskFileEngine
	NumberOfPasswords int
	NumberOfSalts     int
}

// User describes all the properties of a GoCrack user
type User struct {
	UserUUID     string `storm:"unique"`
	Username     string `storm:"unique"`
	Password     string
	Enabled      *bool
	EmailAddress string
	IsSuperUser  bool
	CreatedAt    time.Time
}

// Task describes all the properties of a GoCrack cracking task
type Task struct {
	TaskID            string `storm:"id,unique"`
	TaskName          string
	Status            TaskStatus
	Engine            WorkerCrackEngine
	EnginePayload     interface{} // As of now, this can only be shared.HashcatUserOptions
	Priority          WorkerPriority
	FileID            string // FileID is a reference to TaskFile via TaskFile.FileID
	CreatedBy         string
	CreatedByUUID     string // CreatedBy is a reference to User via User.UserUUID
	CreatedAt         time.Time
	LastUpdatedAt     time.Time
	AssignedToHost    string
	AssignedToDevices *CLDevices
	Comment           *string
	CaseCode          *string
	NumberCracked     int
	NumberPasswords   int
	Error             *string
}

// EngineFile describes a file that is either a dictionary, list of masks, or a rule file for GoCrack
type EngineFile struct {
	FileID          string
	FileName        string
	FileSize        int64
	Description     *string
	UploadedBy      string
	UploadedByUUID  string // UUID of the user who initially uploaded the file
	UploadedAt      time.Time
	LastUpdatedAt   time.Time
	FileType        EngineFileType
	NumberOfEntries int64
	IsShared        bool
	SHA1Hash        string
	SavedAt         string // The physical location on the server where the file is located
}

// CrackedHash is a cracked password from a task
type CrackedHash struct {
	Hash      string `storm:"unique"`
	Value     string
	CrackedAt time.Time
}

// EntitlementEntry is created when a user is granted access to a task, file, etc.
type EntitlementEntry struct {
	UserUUID        string
	EntitledID      string
	GrantedAccessAt time.Time
}

// ActivityLogEntry describes a change in the system to an entity by a user
type ActivityLogEntry struct {
	OccuredAt  time.Time
	UserUUID   string
	Username   string
	EntityID   string
	StatusCode int
	Type       ActivityType
	Path       string
	IPAddress  string
}

// CheckpointFile is a file used to restore a task's state within the engine.
// Note: We may need to revisit this if the files grow in size but as of now, they are only a few hundred bytes.
type CheckpointFile struct {
	TaskID string
	Data   []byte
}
