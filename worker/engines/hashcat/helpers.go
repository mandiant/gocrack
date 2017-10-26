package hashcat

import (
	"io/ioutil"
	"os"

	"github.com/fireeye/gocrack/gocat/restoreutil"
	"github.com/fireeye/gocrack/server/rpc"

	"github.com/rs/zerolog/log"
)

// saveCheckpoint will upload the hashcat restore file to the server
func saveCheckpoint(up rpc.GoCrackRPC, taskid, checkpointPath string) error {
	if _, err := os.Stat(checkpointPath); !os.IsNotExist(err) {
		b, err := ioutil.ReadFile(checkpointPath)
		if err != nil {
			return nil
		}

		if err := up.SendCheckpointFile(rpc.TaskCheckpointSaveRequest{
			TaskID: taskid,
			Data:   b,
		}); err != nil {
			return err
		}

		return os.Remove(checkpointPath)
	}

	return nil
}

// getCheckpoint downloads a checkpoint from the RPC server if one is present
func getCheckpoint(api rpc.GoCrackRPC, taskid, checkpointPath string) error {
	var fd *os.File
	var err error

	bytes, err := api.GetCheckpointFile(taskid)
	if err != nil {
		return err
	}

	if fd, err = os.Create(checkpointPath); err != nil {
		goto Cleanup
	}

	if _, err = fd.Write(bytes); err != nil {
		goto Cleanup
	}

	log.Debug().Str("restore_point", checkpointPath).Msg("Wrote restore file from server to disk")

Cleanup:
	if fd != nil {
		fd.Close()
	}
	return err
}

func changeLocalCheckpoint(rd restoreutil.RestoreData, path string) error {
	var err error
	var fd *os.File

	if fd, err = os.Create(path); err != nil {
		goto Cleanup
	}

	if err = rd.Write(fd); err != nil {
		goto Cleanup
	}

	if err = fd.Sync(); err != nil {
		goto Cleanup
	}

	log.Debug().
		Strs("args", rd.Args).
		Msg("Modified devices in Restore file due to change")

Cleanup:
	if err != nil {
		fd.Close()
	}
	return err
}
