package workmgr

import (
	"sync"
	"time"

	"github.com/mandiant/gocrack/server/storage"
	"github.com/mandiant/gocrack/shared"

	"github.com/tchap/go-exchange/exchange"
)

// ChannelTopic is a PUB/SUB topic that defines a type of event
type ChannelTopic exchange.Topic

var (
	// EngineStatusTopic is the topic for all engine statuses
	EngineStatusTopic = ChannelTopic("EngineStatusTopic")
	// TaskStatusTopic is the topic for all task status changes like Queued, Running, Error, etc.
	TaskStatusTopic = ChannelTopic("TaskStatusTopic")
	// CrackedTopic is the topic for all cracked passwords
	CrackedTopic = ChannelTopic("CrackedTopic")
	// LogTopic is the topic for all log messages from a worker
	LogTopic = ChannelTopic("LogTopic")
	// FinalStatusTopic is the topic that indicates when a task is finished on a worker and includes its final status payload
	FinalStatusTopic = ChannelTopic("FinalStatusTopic")
)

// ConnectedHost is an active, connected host to the WorkManager
type ConnectedHost struct {
	LastCheckin time.Time
	LastBeacon  shared.Beacon
}

// GetRunningTaskIDs returns a list of running TaskIDs on a connected host
func (s ConnectedHost) GetRunningTaskIDs() []string {
	var taskids []string

	for taskid := range s.LastBeacon.Processes {
		taskids = append(taskids, taskid)
	}
	return taskids
}

// CallbackFunc defines the function called whenever we get a message from a subscription
type CallbackFunc func(payload interface{})

// WorkerManager manages all connected hosts to the GoCrack server
type WorkerManager struct {
	mu               *sync.RWMutex
	connectedWorkers map[string]*ConnectedHost
	exch             *exchange.Exchange
	hndls            map[uint]ChannelTopic
}

// TaskEngineStatusBroadcast contains the engine status of a task
type TaskEngineStatusBroadcast struct {
	TaskID string      `json:"task_id"`
	Status interface{} `json:"status"`
}

// TaskStatusFinalBroadcast contains the final engine status of a task
type TaskStatusFinalBroadcast TaskEngineStatusBroadcast

// CrackedPasswordBroadcast contains information about a cracked password
type CrackedPasswordBroadcast struct {
	TaskID    string    `json:"task_id"`
	Hash      string    `json:"hash"`
	Value     string    `json:"value"`
	CrackedAt time.Time `json:"cracked_at"`
}

// TaskStatusChangeBroadcast contains information about a recent task status change from a worker
type TaskStatusChangeBroadcast struct {
	TaskID string             `json:"task_id"`
	Status storage.TaskStatus `json:"status"`
}

// NewWorkerManager creates a new remote worker manager
func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		mu:               &sync.RWMutex{},
		exch:             exchange.New(),
		connectedWorkers: make(map[string]*ConnectedHost),
		hndls:            make(map[uint]ChannelTopic),
	}
}

// HostCheckingIn records a beacon from a connected worker
func (s *WorkerManager) HostCheckingIn(beacon shared.Beacon) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add it to the connected list if it does not exist
	if _, ok := s.connectedWorkers[beacon.Hostname]; !ok {
		s.connectedWorkers[beacon.Hostname] = &ConnectedHost{}
		connectedWorkers.Inc()
	}

	s.connectedWorkers[beacon.Hostname].LastCheckin = time.Now().UTC()
	s.connectedWorkers[beacon.Hostname].LastBeacon = beacon
}

// GetCurrentHostRecord returns the latest host record from a given hostname
func (s WorkerManager) GetCurrentHostRecord(hostname string) *ConnectedHost {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.connectedWorkers[hostname]
	if !ok {
		return nil
	}
	return val
}

// Stop the worker manager and terminate all subscribers
func (s *WorkerManager) Stop() {
	s.exch.Terminate()
}

// GetCurrentWorkers returns a copy of the current workers
func (s WorkerManager) GetCurrentWorkers() map[string]ConnectedHost {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]ConnectedHost, len(s.connectedWorkers))
	for hostname, worker := range s.connectedWorkers {
		out[hostname] = ConnectedHost{
			LastBeacon:  worker.LastBeacon,
			LastCheckin: worker.LastCheckin,
		}
	}

	return out
}

// BroadcastEngineStatusUpdate notifies all subscribers that a task has just sent us a new status payload
func (s *WorkerManager) BroadcastEngineStatusUpdate(taskID string, payload interface{}) error {
	broadcastsSent.WithLabelValues(string(EngineStatusTopic)).Inc()
	return s.exch.Publish(exchange.Topic(EngineStatusTopic), TaskEngineStatusBroadcast{
		TaskID: taskID,
		Status: payload,
	})
}

// BroadcastFinalStatus notifies all subscribers that a task has finished and sent its final task status message
func (s *WorkerManager) BroadcastFinalStatus(taskID string, payload interface{}) error {
	broadcastsSent.WithLabelValues(string(FinalStatusTopic)).Inc()
	return s.exch.Publish(exchange.Topic(FinalStatusTopic), TaskStatusFinalBroadcast{
		TaskID: taskID,
		Status: payload,
	})
}

// BroadcastCrackedPassword notifies all subscribers that a cracked password event has just occurred
func (s *WorkerManager) BroadcastCrackedPassword(taskID, hash, value string, crackedAt time.Time) error {
	broadcastsSent.WithLabelValues(string(CrackedTopic)).Inc()
	return s.exch.Publish(exchange.Topic(CrackedTopic), CrackedPasswordBroadcast{
		TaskID:    taskID,
		Hash:      hash,
		Value:     value,
		CrackedAt: crackedAt,
	})
}

// BroadcastTaskStatusChange notifies all subscribers that the actual task status has changed
func (s *WorkerManager) BroadcastTaskStatusChange(taskid string, status storage.TaskStatus) error {
	broadcastsSent.WithLabelValues(string(TaskStatusTopic)).Inc()
	return s.exch.Publish(exchange.Topic(TaskStatusTopic), TaskStatusChangeBroadcast{
		TaskID: taskid,
		Status: status,
	})
}

// Subscribe to a channel topic and get called asynchronously everytime a new event occurs. If successful, the handle is returned.
func (s *WorkerManager) Subscribe(topic ChannelTopic, f CallbackFunc) (uint, error) {
	hndl, err := s.exch.Subscribe(exchange.Topic(topic), func(t exchange.Topic, e exchange.Event) {
		f(e)
	})

	s.mu.Lock()
	if err == nil {
		s.hndls[uint(hndl)] = topic
		activeSubscriptions.WithLabelValues(string(topic)).Inc()
	}
	s.mu.Unlock()

	return uint(hndl), err
}

// Unsubscribe given a function
func (s *WorkerManager) Unsubscribe(hndl uint) error {
	s.mu.Lock()
	if topic, ok := s.hndls[hndl]; ok {
		activeSubscriptions.WithLabelValues(string(topic)).Dec()
		delete(s.hndls, hndl)
	}
	s.mu.Unlock()

	return s.exch.Unsubscribe(exchange.Handle(hndl))
}
