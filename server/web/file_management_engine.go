package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fireeye/gocrack/server/filemanager"
	"github.com/fireeye/gocrack/server/storage"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type FileType storage.EngineFileType

func (s FileType) MarshalJSON() ([]byte, error) {
	switch storage.EngineFileType(s) {
	case storage.EngineFileDictionary:
		return []byte("\"Dictionary\""), nil
	case storage.EngineFileMasks:
		return []byte("\"Mask(s)\""), nil
	case storage.EngineFileRules:
		return []byte("\"Rule(s)\""), nil
	}
	return []byte("\"Unknown\""), nil
}

// EngineFileMetadataUpdateRequest defines the structure of the metadata update request of an engine file
type EngineFileMetadataUpdateRequest struct {
	IsShared    *bool   `json:"shared,omitempty"`
	FileName    *string `json:"file_name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UploadEngineFileResponse is returned to the user on a successful shared file upload
type UploadEngineFileResponse struct {
	FileID          string `json:"file_id"`
	FileSize        int64  `json:"file_sz"`
	NumberOfEntries int64  `json:"num_of_lines"`
	SHA1            string `json:"sha1"`
}

// EngineFileItem is returned to the user on a shared file listing should mimic storage.EngineFile
type EngineFileItem struct {
	FileID          string    `json:"file_id"`
	FileName        string    `json:"filename"`
	FileSize        int64     `json:"file_sz"`
	Description     *string   `json:"description,omitempty"`
	UploadedAt      time.Time `json:"uploaded_at"`
	UploadedBy      string    `json:"uploaded_by"`
	LastUpdatedAt   time.Time `json:"last_modified"`
	FileType        FileType  `json:"file_type"`
	NumberOfEntries int64     `json:"num_entries"`
	SHA1Hash        string    `json:"sha1"`
}

func convStorageEngineFile(sf storage.EngineFile) EngineFileItem {
	return EngineFileItem{
		FileID:          sf.FileID,
		FileName:        sf.FileName,
		FileSize:        sf.FileSize,
		Description:     sf.Description,
		UploadedBy:      sf.UploadedBy,
		UploadedAt:      sf.UploadedAt,
		LastUpdatedAt:   sf.LastUpdatedAt,
		FileType:        FileType(sf.FileType),
		NumberOfEntries: sf.NumberOfEntries,
		SHA1Hash:        sf.SHA1Hash,
	}
}

// webUploadEngineFile is responsible for handing PUT requests of engine files (dictionaries, masks, rules, etc) for GoCrack tasks
func (s *Server) webUploadEngineFile(c *gin.Context) *WebAPIError {
	var err error
	var fresp *filemanager.FileSaveResponse
	var txn storage.EngineFileTxn

	claim := getClaimInformation(c)
	// if there's an error, isShared is false anyways
	isShared, _ := strconv.ParseBool(c.DefaultQuery("shared", "false"))
	sf := storage.EngineFile{
		FileName:       c.Param("filename"),
		UploadedByUUID: claim.UserUUID,
		FileID:         uuid.NewV4().String(),
		IsShared:       isShared,
	}

	// we can ignore exists here as we'll hit the catchall
	sharedFileType, _ := c.GetQuery("type")
	switch strings.ToLower(sharedFileType) {
	case "0", "dictionary":
		sf.FileType = storage.EngineFileDictionary
	case "1", "masks", "mask":
		sf.FileType = storage.EngineFileMasks
	case "2", "rules", "rule":
		sf.FileType = storage.EngineFileRules
	default:
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			UserError:  "type must be either dictionary (0), masks (1), or rules (2)",
		}
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
		sf.FileName = filename
	}

	if fresp, err = s.fm.SaveFile(f, filename, sf.FileID, sf.FileType); err != nil {
		goto ServerError
	}

	if txn, err = s.stor.NewEngineFileTransaction(); err != nil {
		goto ServerError
	}
	defer txn.Rollback() // wont be rolled back if the commit occurs

	sf.FileSize = fresp.Size
	sf.NumberOfEntries = fresp.NumberOfLines
	sf.SavedAt = fresp.SavedTo
	sf.SHA1Hash = fresp.SHA1

	if err = txn.SaveEngineFile(sf); err != nil {
		goto ServerError
	}

	if err = txn.AddEntitlement(sf, claim.UserUUID); err != nil {
		goto ServerError
	}

	if err = txn.Commit(); err != nil {
		goto ServerError
	}

	c.JSON(http.StatusCreated, &UploadEngineFileResponse{
		FileID:          sf.FileID,
		FileSize:        sf.FileSize,
		NumberOfEntries: sf.NumberOfEntries,
		SHA1:            sf.SHA1Hash,
	})
	return nil

ServerError:
	if fresp != nil && fresp.SavedTo != "" {
		os.Remove(sf.SavedAt)
	}

	return &WebAPIError{
		StatusCode: http.StatusInternalServerError,
		Err:        err,
		UserError:  "The server was unable to process your request. Please try again later",
	}
}

func (s *Server) webDownloadEngineFile(c *gin.Context) *WebAPIError {
	var (
		fileid    = c.Param("fileid")
		err       error
		etf       *storage.EngineFile
		canAccess bool
	)
	claim := getClaimInformation(c)

	if etf, err = s.stor.GetEngineFileByID(fileid); err != nil {
		if err == storage.ErrNotFound {
			goto NoAccess
		}
		goto ServerError
	}

	// Allow the user access to the shared file they're admin
	if claim.IsAdmin {
		goto DownloadFile
	}

	if canAccess, err = s.stor.CheckEntitlement(claim.UserUUID, etf.FileID, storage.EntitlementEngineFile); err != nil {
		goto ServerError
	}

	if canAccess {
		goto DownloadFile
	}

	goto NoAccess

DownloadFile:
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", etf.FileName))

	if fd, err := os.Open(etf.SavedAt); err != nil {
		goto ServerError
	} else {
		if _, err = io.Copy(c.Writer, fd); err != nil {
			goto ServerError
		}
	}

	return nil
NoAccess:
	// Throw a forbidden error if it's a public file as they're allowed to use it in tasks but not
	// download it
	if etf != nil && etf.IsShared {
		return &WebAPIError{
			StatusCode: http.StatusForbidden,
			UserError:  "Cannot download public files unless you're granted access to it",
		}
	}
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

func (s *Server) webGetEngineFiles(c *gin.Context) *WebAPIError {
	var resp []EngineFileItem

	claim := getClaimInformation(c)
	files, err := s.stor.GetEngineFilesForUser(storage.User{
		UserUUID:    claim.UserUUID,
		IsSuperUser: claim.IsAdmin,
	})

	if err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusOK, []EngineFileItem{})
			return nil
		}
		goto ServerError
	}

	resp = make([]EngineFileItem, len(files))
	for i, file := range files {
		resp[i] = convStorageEngineFile(file)
	}
	c.JSON(http.StatusOK, resp)
	return nil

ServerError:
	return &WebAPIError{
		StatusCode: http.StatusInternalServerError,
		Err:        err,
		UserError:  "The server was unable to process your request. Please try again later",
	}
}
