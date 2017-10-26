package workmgr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/fireeye/gocrack/opencl"
	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"
)

type TestWorkManagerSuite struct {
	suite.Suite
	*WorkerManager
}

func (suite *TestWorkManagerSuite) SetupTest() {
	suite.WorkerManager = NewWorkerManager()
}

func (suite *TestWorkManagerSuite) TearDownTest() {
	suite.Stop()
	suite.WorkerManager = nil
}

func (suite *TestWorkManagerSuite) TestHostCheckingIn() {
	suite.HostCheckingIn(shared.Beacon{
		Hostname:       "testcase",
		RequestNewTask: false,
		Devices: shared.DeviceMap{
			1: &shared.Device{
				ID:     1,
				Name:   "My Awesome GPU",
				Type:   opencl.DeviceTypeGPU,
				IsBusy: false,
			},
		},
	})

	hosts := suite.GetCurrentWorkers()
	suite.Len(hosts, 1)
	suite.Contains(hosts, "testcase")

	testRecord := suite.GetCurrentHostRecord("testcase")
	if testRecord == nil {
		suite.FailNow("testRecord should not be nil")
	}
	suite.Equal("testcase", testRecord.LastBeacon.Hostname)

	nothingShouldBeHere := testRecord.GetRunningTaskIDs()
	suite.Len(nothingShouldBeHere, 0)
}

func (suite *TestWorkManagerSuite) TestGetCurrentHostRecordWithBadHost() {
	testRecord := suite.GetCurrentHostRecord("testcase")
	suite.Nil(testRecord)
}

func (suite *TestWorkManagerSuite) TestBroadcastEngineStatusUpdate() {
	exampleBroadcastMsg := TaskEngineStatusBroadcast{
		TaskID: "1337",
		Status: "testing",
	}

	hndl, err := suite.Subscribe(EngineStatusTopic, func(payload interface{}) {
		taskStatus, ok := payload.(TaskEngineStatusBroadcast)
		suite.True(ok)
		suite.Equal(exampleBroadcastMsg.TaskID, taskStatus.TaskID)
	})
	suite.Nil(err)
	defer suite.Unsubscribe(hndl)

	err = suite.BroadcastEngineStatusUpdate("1337", exampleBroadcastMsg)
	suite.Nil(err)
}

func (suite *TestWorkManagerSuite) TestBroadcastCrackedPassword() {
	exampleCrackedPassword := CrackedPasswordBroadcast{
		TaskID:    "1337",
		Hash:      "deadbeefdeadbeefdeadbeefdeadbeef",
		Value:     "deadbeef",
		CrackedAt: time.Now().UTC(),
	}

	hndl, err := suite.Subscribe(CrackedTopic, func(payload interface{}) {
		password, ok := payload.(CrackedPasswordBroadcast)
		suite.True(ok)
		suite.Equal(password.TaskID, exampleCrackedPassword.TaskID)
		suite.Equal(password.Hash, exampleCrackedPassword.Hash)
		suite.Equal(password.Value, exampleCrackedPassword.Value)
	})
	suite.Nil(err)
	defer suite.Unsubscribe(hndl)

	err = suite.BroadcastCrackedPassword(
		exampleCrackedPassword.TaskID,
		exampleCrackedPassword.Hash,
		exampleCrackedPassword.Value,
		exampleCrackedPassword.CrackedAt,
	)
	suite.Nil(err)
}

func (suite *TestWorkManagerSuite) TestBroadcastTaskStatusChange() {
	var taskID = "1337"
	var status = storage.TaskStatusStopped

	hndl, err := suite.Subscribe(TaskStatusTopic, func(payload interface{}) {
		taskstatus, ok := payload.(TaskStatusChangeBroadcast)
		suite.True(ok)
		suite.Equal(taskstatus.TaskID, taskID)
		suite.Equal(taskstatus.Status, status)
	})
	suite.Nil(err)
	defer suite.Unsubscribe(hndl)

	err = suite.BroadcastTaskStatusChange(taskID, status)
	suite.Nil(err)
}

func TestWorkerManager(t *testing.T) {
	suite.Run(t, new(TestWorkManagerSuite))
}

func TestWorkManager(t *testing.T) {
	mgr := NewWorkerManager()

	mgr.HostCheckingIn(shared.Beacon{
		Hostname: "test.local",
		Devices: shared.DeviceMap{
			0: &shared.Device{
				ID:     0,
				Name:   "NVIDIA",
				IsBusy: false,
			},
		},
	})
}
