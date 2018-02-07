package bdb

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"time"

	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

var errExpectedUser = errors.New("expected CreatedBy to be set")

type contains struct {
	list storage.CLDevices
}

func (i *contains) MatchField(v interface{}) (bool, error) {
	ref := reflect.ValueOf(i.list)
	if ref.Kind() != reflect.Slice {
		return false, nil
	}

	devicesFromRecord := v.(*storage.CLDevices)
	if devicesFromRecord == nil {
		return false, nil
	}

	sort.Ints(*devicesFromRecord)
	for i := 0; i < ref.Len(); i++ {
		devIDInSearch := int(ref.Index(i).Int())
		// We only need to find one match in here and then we can bomb out.
		foundIdx := sort.Search(len(*devicesFromRecord), func(idFromRecord int) bool {
			tmp := *devicesFromRecord
			return tmp[idFromRecord] != devIDInSearch
		})
		if foundIdx > 0 {
			return true, nil
		}
	}

	return false, nil
}

// DeviceMatch returns true if *any* element inside v is inside "AssignedToDevices" in a task document
func DeviceMatch(v storage.CLDevices) q.Matcher {
	return q.NewFieldMatcher("AssignedToDevices", &contains{list: v})
}

// GetTaskByID returns a task record based on it's ID
func (s *BoltBackend) GetTaskByID(taskID string) (task *storage.Task, err error) {
	var tmp boltCrackTask

	if err = s.db.From("tasks").One("TaskID", taskID, &tmp); err != nil {
		return nil, convertErr(err)
	}

	task = &tmp.Task
	if err := convertTaskFromMap(task); err != nil {
		return nil, err
	}
	return
}

func convertTaskFromMap(task *storage.Task) error {
	if task.EnginePayload == nil {
		return nil
	}
	if _, ok := task.EnginePayload.(map[string]interface{}); !ok {
		return errors.New("expected t to be map[string]interface{}")
	}

	b, err := json.Marshal(task.EnginePayload)
	if err != nil {
		return err
	}
	switch task.Engine {
	case storage.WorkerHashcatEngine:
		var hcp shared.HashcatUserOptions
		if err := json.Unmarshal(b, &hcp); err != nil {
			return err
		}
		task.EnginePayload = hcp
	default:
		return errors.New("unexpected engine")
	}
	return nil
}

func (s *BoltBackend) getNextTaskForHost(workerHostname string, devicesInUse storage.CLDevices) (*storage.Task, error) {
	var tmp boltCrackTask

	searchQuery := q.And(
		q.Or(
			q.Eq("AssignedToHost", workerHostname),
			q.Eq("AssignedToHost", ""),
		),
		q.Not(
			DeviceMatch(devicesInUse),
		),
		q.Eq("Status", storage.TaskStatusQueued),
	)

	baseQuery := s.db.
		From("tasks").
		Select(searchQuery).
		Limit(1).
		OrderBy("Priority", "CreatedAt")

	if err := baseQuery.First(&tmp); err != nil {
		return nil, convertErr(err)
	}

	v := storage.Task(tmp.Task)
	if err := convertTaskFromMap(&v); err != nil {
		return nil, convertErr(err)
	}
	return &v, nil
}

func (s *BoltBackend) GetPendingTasks(req storage.GetPendingTasksRequest) ([]storage.GetPendingTasksResponseItem, error) {
	var items []storage.GetPendingTasksResponseItem

	if req.CheckForNewTask {
		newTask, err := s.getNextTaskForHost(req.Hostname, req.DevicesInUse)
		if err != nil {
			if err == storage.ErrNotFound {
				goto GetPaused
			}
			return nil, err
		}

		if newTask != nil {
			items = append(items, storage.GetPendingTasksResponseItem{
				Type:    storage.PendingTaskNewRequest,
				Payload: newTask,
			})
		}
	}

GetPaused:
	searchQuery := q.And(
		q.In("TaskID", req.RunningTasks),
		q.Eq("Status", storage.TaskStatusStopping),
	)

	if err := convertErr(s.db.From("tasks").Select(searchQuery).Each(new(boltCrackTask), func(record interface{}) error {
		taskToPause := record.(*boltCrackTask)
		items = append(items, storage.GetPendingTasksResponseItem{
			Type: storage.PendingTaskStatusChange,
			Payload: storage.PendingTaskStatusChangeItem{
				TaskID:    taskToPause.TaskID,
				NewStatus: taskToPause.Status,
			},
		})
		return nil
	})); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return items, nil
}

func (s *BoltBackend) ChangeTaskStatus(taskID string, status storage.TaskStatus, potentialError *string) error {
	var tmp boltCrackTask

	txn, err := s.db.From("tasks").Begin(true)
	if err != nil {
		return convertErr(err)
	}
	defer txn.Rollback()

	if err = txn.One("TaskID", taskID, &tmp); err != nil {
		return convertErr(err)
	}

	if potentialError != nil {
		tmp.Error = potentialError
	}

	tmp.Status = status

	if err = txn.Update(&tmp); err != nil {
		return convertErr(err)
	}
	return txn.Commit()
}

func (s *BoltBackend) TasksSearch(page, limit int, orderby, searchQuery string, isAscending bool, user storage.User) (*storage.SearchResults, error) {
	var baseQuery storm.Query
	sr := &storage.SearchResults{
		Results: make([]storage.Task, 0),
	}

	// If the user is not a super user, we'll grab a list of taskids they are entitled to
	if !user.IsSuperUser {
		var taskids []string

		if err := s.db.From("tasks", bucketEntName).Select(
			q.Eq("UserUUID", user.UserUUID),
		).Each(new(boltEntitlement), func(record interface{}) error {
			be := record.(*boltEntitlement)
			taskids = append(taskids, be.EntitledID)
			return nil
		}); err != nil {
			return nil, convertErr(err)
		}

		if searchQuery != "" {
			baseQuery = s.db.From("tasks").Select(
				q.And(
					q.In("TaskID", taskids),
					q.Or(
						StringContains("TaskName", searchQuery),
						StringContains("CaseCode", searchQuery),
						StringContains("TaskID", searchQuery),
						StringContains("CreatedBy", searchQuery),
						StringContains("Status", searchQuery),
					),
				),
			)
		} else {
			baseQuery = s.db.From("tasks").Select(q.In("TaskID", taskids))
		}
	} else {
		if searchQuery != "" {
			baseQuery = s.db.From("tasks").Select(q.Or(
				StringContains("TaskName", searchQuery),
				StringContains("CaseCode", searchQuery),
				StringContains("TaskID", searchQuery),
				StringContains("CreatedBy", searchQuery),
				StringContains("Status", searchQuery),
			))
		} else {
			baseQuery = s.db.From("tasks").Select()
		}
	}

	if orderby != "" {
		switch orderby {
		case "created_at":
			orderby = "CreatedAt"
		case "status":
			orderby = "Status"
		case "task_id":
			orderby = "TaskID"
		default:
			orderby = "CreatedAt"
		}
		baseQuery = baseQuery.OrderBy(orderby)
	}

	if !isAscending {
		baseQuery = baseQuery.Reverse()
	}

	// Before we limit, we need to determine how many documents exist in this bucket

	total, err := baseQuery.Count(&boltCrackTask{})
	if err != nil {
		return nil, convertErr(err)
	}
	sr.Total = total

	if page != 1 {
		baseQuery = baseQuery.Skip((page - 1) * limit)
	}

	// Now apply the limit & query
	baseQuery = baseQuery.Limit(limit)

	if err := baseQuery.Each(new(boltCrackTask), func(record interface{}) error {
		bt := record.(*boltCrackTask)

		// Storm has no joins...
		numCracked, err := s.countCrackedPasswords(bt.TaskID)
		if err == nil {
			bt.NumberCracked = numCracked
		}

		tf, err := s.GetTaskFileByID(bt.FileID)
		if err == nil {
			bt.NumberPasswords = tf.NumberOfPasswords
		}

		sr.Results = append(sr.Results.([]storage.Task), storage.Task(bt.Task))
		return nil
	}); err != nil {
		return nil, convertErr(err)
	}

	return sr, nil
}

func (s *BoltBackend) countCrackedPasswords(taskid string) (int, error) {
	return s.db.From("tasks", taskid, "results").Count(&boltCrackedHash{})
}

func (s *BoltBackend) SaveCrackedHash(taskid, hash, value string, crackedAt time.Time) error {
	node := s.db.From("tasks", taskid, "results")
	return convertErr(node.Save(&boltCrackedHash{
		DocVersion: curCrackedHashVer,
		CrackedHash: storage.CrackedHash{
			Hash:      hash,
			Value:     value,
			CrackedAt: crackedAt,
		},
	}))
}

func (s *BoltBackend) GetCrackedPasswords(taskid string) (*[]storage.CrackedHash, error) {
	var tmp []boltCrackedHash

	if err := s.db.From("tasks", taskid, "results").All(&tmp); err != nil {
		return nil, convertErr(err)
	}

	out := make([]storage.CrackedHash, len(tmp))
	for i, doc := range tmp {
		out[i] = storage.CrackedHash(doc.CrackedHash)
	}
	return &out, nil
}

func (s *BoltBackend) UpdateTask(taskid string, modifiedFields storage.ModifiableTaskRequest) error {
	txn, err := s.db.From("tasks").Begin(true)
	if err != nil {
		return convertErr(err)
	}
	defer txn.Rollback()

	var tmp boltCrackTask
	if err = txn.One("TaskID", taskid, &tmp); err != nil {
		return convertErr(err)
	}

	if tmp.AssignedToDevices != nil && modifiedFields.AssignedToDevices == nil {
		tmp.AssignedToDevices = nil
	}

	if tmp.AssignedToHost != "" && modifiedFields.AssignedToHost == nil {
		tmp.AssignedToHost = ""
	}

	if modifiedFields.AssignedToDevices != nil {
		tmp.AssignedToDevices = modifiedFields.AssignedToDevices
	}

	if modifiedFields.AssignedToHost != nil {
		tmp.AssignedToHost = *modifiedFields.AssignedToHost
	}

	if modifiedFields.Status != nil {
		tmp.Status = *modifiedFields.Status
	}

	if err = txn.Update(&tmp); err != nil {
		return convertErr(err)
	}
	txn.Commit()

	return nil
}

// DeleteTask implements storage.DeleteTask
func (s *BoltBackend) DeleteTask(taskid string) error {
	var bt boltCrackTask
	if err := s.db.From(bucketTasks).One("TaskID", taskid, &bt); err != nil {
		return convertErr(err)
	}

	return convertErr(s.db.From(bucketTasks).Remove(&bt))
}
