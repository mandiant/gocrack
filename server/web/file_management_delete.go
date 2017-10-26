package web

import (
	"errors"
	"net/http"

	"github.com/fireeye/gocrack/server/storage"
	"github.com/fireeye/gocrack/shared"

	"github.com/gin-gonic/gin"
)

type deletableFile interface {
	GetOwner() string
	GetDocument() interface{}
}

type deletableTaskFile storage.TaskFile

func (s deletableTaskFile) GetOwner() string {
	return s.UploadedByUUID
}

func (s deletableTaskFile) GetDocument() interface{} {
	return storage.TaskFile(s)
}

type deletableEngineFile storage.EngineFile

func (s deletableEngineFile) GetOwner() string {
	return s.UploadedByUUID
}

func (s deletableEngineFile) GetDocument() interface{} {
	return storage.EngineFile(s)
}

// DeleteFileResponse is returned whenever the user deletes a file successfully
type DeleteFileResponse struct {
	Deleted bool   `json:"deleted"`
	FileID  string `json:"file_id"`
}

type deleteFileAPIType uint8

const (
	deleteEngineFileAPI deleteFileAPIType = 1 << iota
	deleteTaskFileAPI
)

func (s *Server) webDeleteFile(ftype deleteFileAPIType) WebAPI {
	return func(c *gin.Context) *WebAPIError {
		requestedFileID := c.Param("fileid")
		claim := getClaimInformation(c)

		var err error
		var f deletableFile

		switch ftype {
		case deleteEngineFileAPI:
			var ef *storage.EngineFile
			ef, err = s.stor.GetEngineFileByID(requestedFileID)
			if err != nil {
				goto ErrorCheck
			}
			f = deletableEngineFile(*ef)
		case deleteTaskFileAPI:
			var tf *storage.TaskFile
			tf, err = s.stor.GetTaskFileByID(requestedFileID)
			if err != nil {
				goto ErrorCheck
			}
			f = deletableTaskFile(*tf)
		}

		// At this time, only admins or users who uploaded the file can delete it
		if !claim.IsAdmin && f.GetOwner() != claim.UserUUID {
			goto NoAccess
		}

		if err := s.fm.DeleteFile(f.GetDocument()); err != nil {
			return &WebAPIError{
				StatusCode: http.StatusInternalServerError,
				Err:        err,
				UserError:  "The server was unable to process your request. Please try again later",
			}
		}

		c.JSON(http.StatusOK, &DeleteFileResponse{
			Deleted: true,
			FileID:  requestedFileID,
		})

		return nil
	ErrorCheck:
		if err == storage.ErrNotFound {
			goto NoAccess
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	NoAccess:
		return &WebAPIError{
			StatusCode: http.StatusNotFound,
			Err:        err,
			UserError:  "The requested file does not exist or you do not have permissions to it",
		}
	}
}

// checkIfTaskFilesArePresent is used by webChangeTaskStatus to verify that all the files are present before we start the task
func (s *Server) checkIfTaskFilesArePresent(task *storage.Task) (bool, error) {
	var err error

	if _, err = s.stor.GetTaskFileByID(task.FileID); err != nil {
		goto CheckError
	}

	switch ep := task.EnginePayload.(type) {
	case shared.HashcatUserOptions:
		if ep.DictionaryFile != nil {
			if _, err = s.stor.GetEngineFileByID(*ep.DictionaryFile); err != nil {
				goto CheckError
			}
		}

		if ep.ManglingRuleFile != nil {
			if _, err = s.stor.GetEngineFileByID(*ep.ManglingRuleFile); err != nil {
				goto CheckError
			}
		}

		if ep.Masks != nil {
			if _, err = s.stor.GetEngineFileByID(*ep.Masks); err != nil {
				goto CheckError
			}
		}
	default:
		return false, errors.New("unknown task.EnginePayload")
	}
	return true, nil

CheckError:
	if err == storage.ErrNotFound {
		return false, nil
	}
	return false, err
}
