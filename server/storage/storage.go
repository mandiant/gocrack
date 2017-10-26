package storage

import (
	"errors"
	"time"
)

var (
	// ErrNotFound is raised whenever a database backend raises it's driver specific ErrNotFound
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists is raised whenever a database backend raises a driver specific document already exists error
	ErrAlreadyExists = errors.New("already exists")
	// ErrViolateConstraint is raised whenever a database raises a driver specific constraint (unique) error
	ErrViolateConstraint = errors.New("constraint violation")
)

// PasswordCheckFunc defines a function that drivers use for validating a password from a found record.
// If the password stored in the driver is correct, this function should return true. Otherwise the login will fail
type PasswordCheckFunc func(password string) (ok bool)

// StorageError is a generic error raised by a backend
type StorageError struct {
	DriverError error
}

func (e StorageError) Error() string {
	return e.DriverError.Error()
}

// TaskFileTxn describes all the methods needed for a task file transaction
type TaskFileTxn interface {
	SaveTaskFile(tf TaskFile) error
	AddEntitlement(tf TaskFile, userid string) error
	Rollback() error
	Commit() error
}

type EngineFileTxn interface {
	SaveEngineFile(tf EngineFile) error
	AddEntitlement(tf EngineFile, userid string) error
	Rollback() error
	Commit() error
}

// CreateTaskTxn describes all the methods needed for the transaction tasked with creating a task
type CreateTaskTxn interface {
	CreateTask(t *Task) error
	GrantEntitlement(userUUID string, t Task) (err error)
	Rollback() error
	Commit() error
}

// SearchResults includes the results of a search request along with the total number of documents before limits were applied
type SearchResults struct {
	Results interface{}
	Total   int
}

// GetPendingTasksRequest is used in the GetPendingTasks API to build a search query and return all actions
// that a host should take
type GetPendingTasksRequest struct {
	Hostname        string
	DevicesInUse    CLDevices
	RunningTasks    []string
	CheckForNewTask bool
}

type PendingTaskStatusChangeItem struct {
	TaskID    string
	NewStatus TaskStatus
}

type GetPendingTasksResponseItem struct {
	Type    PendingTaskPayloadType
	Payload interface{}
}

// Backend describes all APIs that a backend storage driver should implement
type Backend interface {
	Close() error

	LogActivity(ActivityLogEntry) error
	GetActivityLog(string) ([]ActivityLogEntry, error)
	RemoveActivityEntries(string) error

	// File Management APIs

	NewTaskFileTransaction() (TaskFileTxn, error)
	GetTaskFileByID(string) (*TaskFile, error)
	ListTasksForUser(User) ([]TaskFile, error)
	NewEngineFileTransaction() (EngineFileTxn, error)
	GetEngineFileByID(storageID string) (*EngineFile, error)
	GetEngineFilesForUser(User) ([]EngineFile, error)
	DeleteEngineFile(string) error
	DeleteTaskFile(string) error

	// Task Management APIs
	NewTaskCreateTransaction() (CreateTaskTxn, error)
	GetTaskByID(string) (*Task, error)
	ChangeTaskStatus(string, TaskStatus, *string) error
	TasksSearch(page, limit int, orderby, searchQuery string, isAscending bool, user User) (*SearchResults, error)
	SaveCrackedHash(taskid, hash, value string, crackedAt time.Time) error
	GetCrackedPasswords(string) (*[]CrackedHash, error)
	GetPendingTasks(GetPendingTasksRequest) ([]GetPendingTasksResponseItem, error)
	UpdateTask(string, ModifiableTaskRequest) error
	SaveTaskCheckpoint(CheckpointFile) error
	GetTaskCheckpoint(string) ([]byte, error)
	DeleteTask(string) error

	// Rights Management APIs
	CheckEntitlement(userUUID, entityID string, entType EntitlementType) (bool, error)
	// GrantEntitlement grants the user access to the record passed in via document
	GrantEntitlement(user User, document interface{}) (err error)
	// RevokeEntitlement removes the user's access to the document
	RevokeEntitlement(user User, document interface{}) (err error)
	GetEntitlementsForTask(entityID string) ([]EntitlementEntry, error)
	RemoveEntitlements(string, EntitlementType) error

	// User Management APIs
	SearchForUserByPassword(username string, passcheck PasswordCheckFunc) (userRecord *User, err error)
	CreateUser(user *User) (err error)
	GetUserByID(userUUID string) (user *User, err error)
	GetUsers() ([]User, error)
	EditUser(string, UserModifyRequest) error
}
