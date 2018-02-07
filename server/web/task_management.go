package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type TaskCrackEngineFancy storage.WorkerCrackEngine

func (s TaskCrackEngineFancy) MarshalJSON() ([]byte, error) {
	switch storage.WorkerCrackEngine(s) {
	case storage.WorkerHashcatEngine:
		return []byte("\"Hashcat\""), nil
	default:
		return []byte("\"Unknown\""), nil
	}
}

type TaskPriorityFancy storage.WorkerPriority

func (s TaskPriorityFancy) MarshalJSON() ([]byte, error) {
	switch storage.WorkerPriority(s) {
	case storage.WorkerPriorityLow:
		return []byte("\"Low\""), nil
	case storage.WorkerPriorityNormal:
		return []byte("\"Normal\""), nil
	case storage.WorkerPriorityHigh:
		return []byte("\"High\""), nil
	default:
		return []byte("\"Unknown\""), nil
	}
}

// CreateTaskRequest defines the request for a new task creation event
type CreateTaskRequest struct {
	TaskName          string                    `json:"task_name"`
	Engine            storage.WorkerCrackEngine `json:"engine"`
	FileID            string                    `json:"file_id"` // FileID is a reference to TaskFile via TaskFile.FileID
	CaseCode          *string                   `json:"case_code,omitempty"`
	AssignedToHost    *string                   `json:"assigned_host,omitempty"`
	AssignedToDevices *storage.CLDevices        `json:"assigned_devices,omitempty"`
	Comment           *string                   `json:"comment,omitempty"`
	EnginePayload     json.RawMessage           `json:"payload"` // The structure of EnginePayload differs based on Engine
	TaskDuration      int                       `json:"task_duration"`
	Priority          *storage.WorkerPriority   `json:"priority,omitempty"`
	AdditionalUsers   *[]string                 `json:"additional_users,omitempty"`
}

// CreateTaskResponse defines response on a successful task creation event
type CreateTaskResponse struct {
	TaskID    string             `json:"taskid"`
	CreatedAt time.Time          `json:"created_at"`
	Status    storage.TaskStatus `json:"status"`
}

// HashcatEnginePayload defines the structure of task.EnginePayload for jobs created for the hashcat engine
type HashcatEnginePayload struct {
	HashType         string          `json:"hash_type"`
	AttackMode       string          `json:"attack_mode"`
	Masks            *EngineFileItem `json:"masks,omitempty"`
	DictionaryFile   *EngineFileItem `json:"dictionary_file,omitempty"`
	ManglingRuleFile *EngineFileItem `json:"mangling_file,omitempty"`
}

// TaskInfoResponseItem defines the response for all the information possible about a given task
type TaskInfoResponseItem struct {
	TaskID            string               `json:"task_id"`
	TaskName          string               `json:"task_name"`
	CaseCode          *string              `json:"case_code,omitempty"`
	Comment           *string              `json:"comment,omitempty"`
	AssignedToHost    string               `json:"assigned_host,omitempty"`
	AssignedToDevices *storage.CLDevices   `json:"assigned_devices,omitempty"`
	Status            storage.TaskStatus   `json:"status"`
	CreatedBy         string               `json:"created_by"`
	CreatedByUUID     string               `json:"created_by_uuid"`
	CreatedAt         time.Time            `json:"created_at"`
	Engine            TaskCrackEngineFancy `json:"engine"`
	FileID            string               `json:"-"` // FileID is a reference to TaskFile via TaskFile.FileID
	Priority          TaskPriorityFancy    `json:"priority"`
	EnginePayload     interface{}          `json:"engine_options"`
	TaskDuration      int                  `json:"task_duration"`
	FileInfo          *TaskFileItem        `json:"password_file"`
	Error             *string              `json:"error,omitempty"`
}

// TaskListingResponseItem includes the "bare minimum" information about a task for listing purposes
type TaskListingResponseItem struct {
	TaskID         string             `json:"task_id"`
	TaskName       string             `json:"task_name"`
	CaseCode       string             `json:"case_code"`
	Status         storage.TaskStatus `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
	CreatedBy      string             `json:"created_by"`
	PasswordsTotal int                `json:"passwords_total"`
	CrackedTotal   int                `json:"cracked_total"`
}

type TaskListingResponse struct {
	Data  []TaskListingResponseItem `json:"data"`
	Count int                       `json:"count"`
}

type PasswordResponseItem struct {
	Hash      string    `json:"hash"`
	Value     string    `json:"value"`
	CrackedAt time.Time `json:"cracked_at"`
}

type PasswordListResponse struct {
	Data  []PasswordResponseItem `json:"data"`
	Count int                    `json:"count"`
}

type ChangeTaskStatusRequest struct {
	Status string `json:"state"`
}

type ModifyTaskRequest struct {
	AssignedToHost    *string             `json:"assigned_host,omitempty"`
	AssignedToDevices *storage.CLDevices  `json:"assigned_devices,omitempty"`
	Status            *storage.TaskStatus `json:"task_status,omitempty"`
	TaskDuration      *int                `json:"task_duration,omitempty"`
}

// HashcatTaskPayload defines the structure of a Task request which should be executed in a worker with the hashcat engine
type HashcatTaskPayload shared.HashcatUserOptions

func (hcp HashcatTaskPayload) validate() []string {
	errs := make([]string, 0)

	switch hcp.AttackMode {
	case shared.AttackModeStraight:
		if hcp.DictionaryFile == nil || *hcp.DictionaryFile == "" {
			errs = append(errs, "dictionary_file must be set on a straight/dictionary attack mode")
		}
	case shared.AttackModeBruteForce:
		if hcp.Masks == nil || *hcp.Masks == "" {
			errs = append(errs, "masks must be set on a brute force attack mode")
		}
	}

	return errs
}

func (s CreateTaskRequest) validate() []string {
	errs := make([]string, 0)

	if s.TaskName == "" {
		errs = append(errs, "task_name must not be empty")
	}

	if _, err := uuid.FromString(s.FileID); err != nil {
		errs = append(errs, "file_id must be a valid UUID")
	}

	switch s.Engine {
	case storage.WorkerHashcatEngine:
	// do nothing for supported engines
	default:
		errs = append(errs, "engine must be hashcat")
	}

	return errs
}

func (s *Server) webCreateTask(c *gin.Context) *WebAPIError {
	var (
		request CreateTaskRequest
		err     error
		tf      *storage.TaskFile
		payload interface{}
		task    storage.Task
		now     = time.Now().UTC()
	)

	claim := getClaimInformation(c)

	if err := c.BindJSON(&request); err != nil {
		goto BadRequest
	}

	if errs := request.validate(); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, APIValidationErrors{
			Valid:  false,
			Errors: errs,
		})
		return nil
	}

	// Check if the file exists
	if tf, err = s.stor.GetTaskFileByID(request.FileID); err != nil || tf == nil {
		if err == storage.ErrNotFound {
			goto NoAccess
		}
		goto ServerError
	}

	// Check if the user has access to this document if they are not an admin
	if !claim.IsAdmin {
		if canAccess, e := s.stor.CheckEntitlement(claim.UserUUID, tf.FileID, storage.EntitlementTaskFile); e != nil {
			if e == storage.ErrNotFound {
				goto NoAccess
			}
			err = e
			goto ServerError
		} else if !canAccess {
			goto NoAccess
		}
	}

	switch request.Engine {
	case storage.WorkerHashcatEngine:
		var hcp HashcatTaskPayload
		if err := json.Unmarshal(request.EnginePayload, &hcp); err != nil {
			goto BadRequest
		}

		if errs := hcp.validate(); len(errs) > 0 {
			c.JSON(http.StatusBadRequest, APIValidationErrors{
				Valid:  false,
				Errors: errs,
			})
			return nil
		}
		// XXX(cschmitt): Check the entitlement on any of the fields that are a remote file
		payload = hcp
	default:
		goto BadRequest
	}

	task = storage.Task{
		TaskName:      request.TaskName,
		TaskID:        uuid.NewV4().String(),
		CreatedAt:     now,
		CreatedByUUID: claim.UserUUID,
		LastUpdatedAt: now,
		Engine:        request.Engine,
		TaskDuration:  request.TaskDuration,
		EnginePayload: payload,
		FileID:        tf.FileID,
		CaseCode:      request.CaseCode,
		Comment:       request.Comment,
	}

	// Set the device affinity if the request contains valid data
	if request.AssignedToHost != nil && request.AssignedToDevices != nil {
		task.AssignedToDevices = request.AssignedToDevices
		task.AssignedToHost = *request.AssignedToHost
	}

	// Default priority if none was specified
	if request.Priority == nil {
		task.Priority = storage.WorkerPriorityNormal
	} else {
		task.Priority = *request.Priority
	}

	// Save the task, it's creator, and any additional users
	if txn, e := s.stor.NewTaskCreateTransaction(); e != nil {
		err = e
		goto ServerError
	} else {
		if err = txn.CreateTask(&task); err != nil {
			goto ServerError
		}
		defer txn.Rollback() // wont be rolled back if the commit occurs

		if request.AdditionalUsers != nil && len(*request.AdditionalUsers) > 0 {
			for _, userid := range *request.AdditionalUsers {
				if err = txn.GrantEntitlement(userid, task); err != nil {
					goto ServerError
				}
			}
		}
		txn.Commit()
	}

	c.JSON(http.StatusCreated, &CreateTaskResponse{
		TaskID:    task.TaskID,
		CreatedAt: task.CreatedAt,
		Status:    task.Status,
	})
	return nil

NoAccess:
	return &WebAPIError{
		StatusCode: http.StatusNotFound,
		Err:        err,
		UserError:  "The requested file does not exist or you do not have permissions to it",
	}
BadRequest:
	return &WebAPIError{
		StatusCode: http.StatusBadRequest,
		Err:        err,
		CanErrorBeShownToUser: true,
		UserError:             "Your request is malformed",
	}
ServerError:
	return &WebAPIError{
		StatusCode: http.StatusInternalServerError,
		Err:        err,
		UserError:  "The server was unable to process your request. Please try again later",
	}
}

func (s *Server) webGetTaskInfo(c *gin.Context) *WebAPIError {
	var (
		taskid   = c.Param("taskid")
		err      error
		task     *storage.Task
		resp     TaskInfoResponseItem
		taskfile *storage.TaskFile
	)

	if task, err = s.stor.GetTaskByID(taskid); err != nil || task == nil {
		if err == storage.ErrNotFound {
			return &WebAPIError{
				StatusCode: http.StatusNotFound,
				Err:        err,
				UserError:  "The requested file does not exist or you do not have permissions to it",
			}
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	}

	resp = convertStorageTaskToItem(s.stor, *task)
	// Get the task file information
	if taskfile, err = s.stor.GetTaskFileByID(task.FileID); err != nil {
		// the file might not exist. If that's the case, we'll just keep it null
		if err != storage.ErrNotFound {
			return &WebAPIError{
				StatusCode: http.StatusInternalServerError,
				Err:        err,
				UserError:  "The server was unable to process your request. Please try again later",
			}
		}
	} else if taskfile != nil {
		resp.FileInfo = convStorTaskFileItem(*taskfile)
	}

	c.JSON(http.StatusOK, &resp)
	return nil
}

func (s *Server) getAvailableTasks(c *gin.Context) *WebAPIError {
	var (
		pageNum, limit int
		orderBy        = c.DefaultQuery("orderBy", "")
		ascendingOrder bool
		err            error
		searchResults  *storage.SearchResults
	)

	claim := getClaimInformation(c)
	searchQuery := c.Query("query")

	if pageNum, err = strconv.Atoi(c.DefaultQuery("page", "0")); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        errors.New("page must be an integer"),
			CanErrorBeShownToUser: true,
		}
	}

	if limit, err = strconv.Atoi(c.DefaultQuery("limit", "20")); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        errors.New("limit must be an integer"),
			CanErrorBeShownToUser: true,
		}
	}

	if ascendingOrder, err = strconv.ParseBool(c.DefaultQuery("ascending", "0")); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        errors.New("ascending must be a boolean"),
			CanErrorBeShownToUser: true,
		}
	}

	if searchResults, err = s.stor.TasksSearch(pageNum, limit, orderBy, searchQuery, ascendingOrder, storage.User{
		UserUUID:    claim.UserUUID,
		IsSuperUser: claim.IsAdmin,
	}); err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusOK, &TaskListingResponse{
				Data:  []TaskListingResponseItem{},
				Count: 0,
			})
			return nil
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	}

	tasks := searchResults.Results.([]storage.Task)
	items := make([]TaskListingResponseItem, len(tasks))
	for i, task := range tasks {
		item := TaskListingResponseItem{
			TaskID:         task.TaskID,
			TaskName:       task.TaskName,
			Status:         task.Status,
			CreatedAt:      task.CreatedAt,
			CreatedBy:      task.CreatedBy,
			PasswordsTotal: task.NumberPasswords,
			CrackedTotal:   task.NumberCracked,
		}
		// Add the case code if the pointer is not nil
		if task.CaseCode != nil {
			item.CaseCode = *task.CaseCode
		}
		items[i] = item
	}

	c.JSON(http.StatusOK, &TaskListingResponse{
		Data:  items,
		Count: searchResults.Total,
	})
	return nil
}

func (s *Server) webGetTaskPasswords(c *gin.Context) *WebAPIError {
	var (
		taskid    = c.Param("taskid")
		err       error
		task      *storage.Task
		passwords []PasswordResponseItem
	)

	if task, err = s.stor.GetTaskByID(taskid); err != nil || task == nil {
		if err == storage.ErrNotFound {
			goto NoAccess
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	}

	if storpass, err := s.stor.GetCrackedPasswords(taskid); err != nil {
		if err == storage.ErrNotFound {
			goto NoAccess
		}
	} else {
		if storpass == nil || (storpass != nil && len(*storpass) == 0) {
			c.JSON(http.StatusOK, &PasswordListResponse{
				Data:  []PasswordResponseItem{},
				Count: 0,
			})
			return nil
		}

		for _, pass := range *storpass {
			passwords = append(passwords, PasswordResponseItem(pass))
		}
		c.JSON(http.StatusOK, &PasswordListResponse{
			Data:  passwords,
			Count: len(passwords),
		})
	}
	return nil

NoAccess:
	return &WebAPIError{
		StatusCode: http.StatusNotFound,
		Err:        err,
		UserError:  "The requested file does not exist or you do not have permissions to it",
	}
}

func (s *Server) webChangeTaskStatus(c *gin.Context) *WebAPIError {
	var (
		taskid    = c.Param("taskid")
		err       error
		task      *storage.Task
		req       ChangeTaskStatusRequest
		newStatus storage.TaskStatus
	)

	if err := c.BindJSON(&req); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
			CanErrorBeShownToUser: true,
		}
	}

	switch strings.ToLower(req.Status) {
	case "stop", "pause":
		newStatus = storage.TaskStatusStopping
	case "start", "resume":
		newStatus = storage.TaskStatusQueued
	default:
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			UserError:  "The request status must either be `start` or `stop`",
		}
	}

	if task, err = s.stor.GetTaskByID(taskid); err != nil || task == nil {
		if err == storage.ErrNotFound {
			return &WebAPIError{
				StatusCode: http.StatusNotFound,
				Err:        err,
				UserError:  "The requested file does not exist or you do not have permissions to it",
			}
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	}

	switch task.Status {
	case storage.TaskStatusRunning, storage.TaskStatusDequeued:
		if newStatus == storage.TaskStatusQueued {
			return &WebAPIError{
				StatusCode: http.StatusBadRequest,
				UserError:  "The task is already in the processing of starting",
			}
		}
	case storage.TaskStatusStopped, storage.TaskStatusError:
		// do nothing
	case storage.TaskStatusStopping:
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			UserError:  "The task is already in the processing of stopping",
		}
	default:
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			UserError:  "The task status can only be changed if the task is Stopped, Running, or Error",
		}
	}

	if newStatus == storage.TaskStatusQueued {
		ok, err := s.checkIfTaskFilesArePresent(task)
		if err != nil {
			return &WebAPIError{
				StatusCode: http.StatusInternalServerError,
				Err:        err,
				UserError:  "The server was unable to process your request. Please try again later",
			}
		}

		if !ok {
			return &WebAPIError{
				StatusCode: http.StatusBadRequest,
				UserError:  "One or more files required by this job are no longer present. The task can no longer be resumed",
			}
		}
	}

	if err := s.stor.ChangeTaskStatus(taskid, newStatus, nil); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}

func (s *Server) webModifyTask(c *gin.Context) *WebAPIError {
	var taskid = c.Param("taskid")
	var request ModifyTaskRequest

	if err := c.BindJSON(&request); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
			CanErrorBeShownToUser: true,
		}
	}

	claim := getClaimInformation(c)

	// Only admins can override the task status
	if !claim.IsAdmin && request.Status != nil {
		request.Status = nil
	}

	if err := s.stor.UpdateTask(taskid, storage.ModifiableTaskRequest(request)); err != nil {
		if err == storage.ErrNotFound {
			return &WebAPIError{
				StatusCode: http.StatusNotFound,
				Err:        err,
				UserError:  "The requested file does not exist or you do not have permissions to it",
			}
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}
