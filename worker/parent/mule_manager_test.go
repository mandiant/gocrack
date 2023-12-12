package parent

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var pythonCommand = "python3"

type remoteProcessManagerTest struct {
	*RemoteProcessManager

	outputCh RemoteOutput
	stopCh   chan bool
	wg       *sync.WaitGroup
	output   []PipeResponse
}

func (s *remoteProcessManagerTest) outputRecorder() {
	defer s.wg.Done()
DebugLoop:
	for {
		select {
		case <-s.stopCh:
			break DebugLoop
		case msg := <-s.outputCh:
			s.output = append(s.output, msg)
		}
	}
}

func (s *remoteProcessManagerTest) CleanupTest() {
	close(s.outputCh)
	close(s.stopCh)
	s.wg.Wait()
}

func newRemoteProcessManagerTest() *remoteProcessManagerTest {
	outCh := make(RemoteOutput, 100)
	test := &remoteProcessManagerTest{
		outputCh:             outCh,
		stopCh:               make(chan bool, 1),
		RemoteProcessManager: NewRemoteProcess(outCh),
		wg:                   &sync.WaitGroup{},
		output:               make([]PipeResponse, 0),
	}

	test.wg.Add(1)
	go test.outputRecorder()
	return test
}

func TestRemoteProcessManager(t *testing.T) {
	rpmt := newRemoteProcessManagerTest()
	defer rpmt.CleanupTest()

	// pid should be unknown at this point because the process hasnt started
	assert.Equal(t, PIDUNKN, rpmt.GetRunningPID())

	err := rpmt.Start(pythonCommand, "testdata/printloop.py", "--rc", "25")
	assert.Nil(t, err)
	rpmt.Wait()

	for i, output := range rpmt.output {
		assert.Equal(t, fmt.Sprintf("I: %d", i), output.Data)
	}

	// Shouldn't be able to get memory here
	assert.Equal(t, uint64(0), rpmt.GetMemoryRSS())
	// PID should be > 0 even though the process has exited by this point
	assert.True(t, rpmt.GetRunningPID() > 0)
	assert.Equal(t, 25, rpmt.GetReturnCode())
}

func TestRemoteProcessManagerWithBadProcess(t *testing.T) {
	rpmt := newRemoteProcessManagerTest()
	defer rpmt.CleanupTest()

	err := rpmt.Start("notAValidProcess")
	rpmt.Wait()
	assert.Error(t, err)
}

func rpmKillTest(t *testing.T, sigKill bool) {
	rpmt := newRemoteProcessManagerTest()
	defer rpmt.CleanupTest()

	err := rpmt.Start(pythonCommand, "testdata/printloop.py", "--max", "25")
	assert.Nil(t, err)
	time.Sleep(2 * time.Second)

	if sigKill {
		rpmt.StopKill()
	} else {
		rpmt.StopTerm()
	}
	rpmt.Wait()
	// there should really only be 1 item in here but lets allow for a really slow computer
	assert.True(t, len(rpmt.output) < 4)
}

func TestRemoteProcessManagerStopKill(t *testing.T) {
	rpmKillTest(t, true)
}

func TestRemoteProcessManagerStopTerminate(t *testing.T) {
	rpmKillTest(t, false)
}
