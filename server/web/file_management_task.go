package web

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fireeye/gocat"
	"github.com/fireeye/gocrack/server/storage"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type APIEngine storage.TaskFileEngine

func (s APIEngine) MarshalJSON() ([]byte, error) {
	switch storage.TaskFileEngine(s) {
	case storage.TaskFileEngineAll:
		return []byte("\"All\""), nil
	case storage.TaskFileEngineHashcat:
		return []byte("\"Hashcat\""), nil
	}
	return []byte("\"Unknown\""), nil
}

// UploadedFileResponse is returned on a successful upload
type UploadedFileResponse struct {
	SHA1       string    `json:"sha1"`
	FileUUID   string    `json:"file_uuid"`
	FileSize   int64     `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// TaskFileItem describes a file that is used for tasks
type TaskFileItem struct {
	FileID            string    `json:"file_id"`
	UploadedAt        time.Time `json:"uploaded_at"`
	UploadedBy        string    `json:"uploaded_by"`
	UploadedByUUID    string    `json:"uploaded_by_uuid"`
	FileSize          int64     `json:"file_size"`
	FileName          string    `json:"filename"`
	SHA1Hash          string    `json:"sha1"`
	ForEngine         APIEngine `json:"use_in_engine"`
	NumberOfPasswords int       `json:"num_passwords"`
	NumberOfSalts     int       `json:"num_salts"`
}

type TaskFileLintError struct {
	Errors  []string `json:"errors"`
	Message string   `json:"msg"`
}

func convStorTaskFileItem(i storage.TaskFile) *TaskFileItem {
	return &TaskFileItem{
		FileID:            i.FileID,
		UploadedAt:        i.UploadedAt,
		UploadedBy:        i.UploadedBy,
		UploadedByUUID:    i.UploadedByUUID,
		FileSize:          i.FileSize,
		FileName:          i.FileName,
		SHA1Hash:          i.SHA1Hash,
		ForEngine:         APIEngine(i.ForEngine),
		NumberOfPasswords: i.NumberOfPasswords,
		NumberOfSalts:     i.NumberOfSalts,
	}
}

// getFileFromContext attempts to locate the file in a PUT request whether it may be raw bytes or a multipart/form-data submission
func getFileFromContext(c *gin.Context) (f io.ReadCloser, filename string, e *WebAPIError) {
	var err error

	contentType := strings.ToLower(c.Request.Header.Get("Content-Type"))
	// separate the content-type for multi-part
	contentTypeSep := strings.Index(contentType, ";")
	if contentTypeSep != -1 {
		contentType = contentType[:contentTypeSep]
	}

	switch contentType {
	case "multipart/form-data":
		var headers *multipart.FileHeader
		if f, headers, err = c.Request.FormFile("file"); err != nil {
			e = &WebAPIError{
				StatusCode: http.StatusBadRequest,
				Err:        err,
				UserError:  "Could not save your file",
			}
			return
		}
		filename = headers.Filename
	default:
		f = c.Request.Body
	}
	return
}

func (s *Server) webUploadTaskFile(c *gin.Context) *WebAPIError {
	var err error
	var txn storage.TaskFileTxn

	claim := getClaimInformation(c)

	tf := storage.TaskFile{
		FileName:       c.Param("filename"),
		UploadedAt:     time.Now().UTC(),
		UploadedByUUID: claim.UserUUID,
		FileID:         uuid.NewV4().String(),
	}

	engineTypeStr, _ := c.GetQuery("engine")
	switch strings.ToLower(engineTypeStr) {
	case "1", "hashcat":
		tf.ForEngine = storage.TaskFileEngineHashcat
	default:
		tf.ForEngine = storage.TaskFileEngineAll
	}

	f, filename, e := getFileFromContext(c)
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	if e != nil {
		return e
	}

	if filename != "" {
		tf.FileName = filename
	}

	fresp, err := s.fm.SaveFile(f, filename, tf.FileID, tf.ForEngine)
	if err != nil {
		goto ServerError
	}

	switch tf.ForEngine {
	case storage.TaskFileEngineHashcat:
		fileType, _ := c.GetQuery("filetype")
		ftint, err := strconv.Atoi(fileType)
		if err != nil {
			os.Remove(fresp.SavedTo)
			return &WebAPIError{
				StatusCode: http.StatusBadRequest,
				UserError:  "filetype must be an integer",
			}
		}

		vresults, err := gocat.ValidateHashes(fresp.SavedTo, uint32(ftint))
		if err != nil {
			os.Remove(fresp.SavedTo)
			return &WebAPIError{
				StatusCode: http.StatusInternalServerError,
				Err:        err,
				UserError:  "Could not validate your file",
			}
		}

		if vresults != nil && len(vresults.Errors) > 0 {
			os.Remove(fresp.SavedTo)
			c.JSON(http.StatusBadRequest, &TaskFileLintError{
				Errors:  vresults.Errors,
				Message: "One or more hashes in the file are not valid for the filetype you have selected",
			})
			return nil
		}
		tf.NumberOfPasswords = int(vresults.NumHashesUnique)
		tf.NumberOfSalts = int(vresults.NumSalts)
	}

	if txn, err = s.stor.NewTaskFileTransaction(); err != nil {
		goto ServerError
	}

	tf.SHA1Hash = fresp.SHA1
	tf.FileSize = fresp.Size
	tf.SavedAt = fresp.SavedTo

	if err = txn.SaveTaskFile(tf); err != nil {
		goto ServerError
	}

	if err = txn.AddEntitlement(tf, claim.UserUUID); err != nil {
		goto ServerError
	}

	if err = txn.Commit(); err != nil {
		goto ServerError
	}

	c.JSON(http.StatusCreated, &UploadedFileResponse{
		SHA1:       fresp.SHA1,
		FileUUID:   tf.FileID,
		FileSize:   fresp.Size,
		UploadedAt: tf.UploadedAt,
	})
	return nil

ServerError:
	if fresp.SavedTo != "" {
		os.Remove(fresp.SavedTo)
	}

	if txn != nil {
		txn.Rollback()
	}
	return &WebAPIError{
		StatusCode: http.StatusInternalServerError,
		Err:        err,
		UserError:  "Could not save your file",
	}
}

func (s *Server) webListAvailableTaskFiles(c *gin.Context) *WebAPIError {
	claim := getClaimInformation(c)

	tasks, err := s.stor.ListTasksForUser(storage.User{
		UserUUID:    claim.UserUUID,
		IsSuperUser: claim.IsAdmin,
	})

	if err != nil {
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "Error retrieving list of task files",
		}
	}

	// Days since I complained about go's type system: 0
	resp := []*TaskFileItem{}
	for _, task := range tasks {
		resp = append(resp, convStorTaskFileItem(task))
	}

	c.JSON(http.StatusOK, resp)
	return nil
}

// webDownloadTask is responsible for downloading task files (uncracked passwords)
func (s *Server) webDownloadTaskFile(c *gin.Context) *WebAPIError {
	var (
		tf     *storage.TaskFile
		fileid = c.Param("fileid")
		err    error
	)

	if tf, err = s.stor.GetTaskFileByID(fileid); err != nil {
		if err == storage.ErrNotFound {
			goto NoAccess
		}
		goto ServerError
	}

	c.Header("X-Hash-SHA1", tf.SHA1Hash)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileid))

	if fd, err := os.Open(tf.SavedAt); err != nil {
		goto ServerError
	} else {
		if _, err = io.Copy(c.Writer, fd); err != nil {
			goto ServerError
		}
	}

	return nil
NoAccess:
	return &WebAPIError{
		StatusCode: http.StatusNotFound,
		Err:        err,
		UserError:  "The requested file does not exist or you do not have permissions to it",
	}
ServerError:
	return &WebAPIError{
		StatusCode: http.StatusInternalServerError,
		Err:        err,
		UserError:  "The server was unable to process your request. Please try again later",
	}
}
