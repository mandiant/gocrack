package web

import (
	"net/http"
	"time"

	"github.com/mandiant/gocrack/server/storage"

	"github.com/gin-gonic/gin"
)

type EntitlementResponseEntry struct {
	UserUUID        string    `json:"user_id"`
	EntitledID      string    `json:"-"`
	GrantedAccessAt time.Time `json:"granted_access_on"`
}

type EntitlementResponseEntries []EntitlementResponseEntry

func convertEntitlementEntryFromStorage(s []storage.EntitlementEntry) []EntitlementResponseEntry {
	if s == nil {
		return []EntitlementResponseEntry{}
	}

	ntries := make([]EntitlementResponseEntry, len(s))
	for i, ent := range s {
		ntries[i] = EntitlementResponseEntry(ent)
	}
	return ntries
}

func (s *Server) webGetTaskEntitlements(c *gin.Context) *WebAPIError {
	var (
		taskid       = c.Param("taskid")
		err          error
		task         *storage.Task
		entitlements []storage.EntitlementEntry
	)

	if task, err = s.stor.GetTaskByID(taskid); err != nil || task == nil {
		if err == storage.ErrNotFound {
			goto NoAccess
		}
		goto ServerError
	}

	if entitlements, err = s.stor.GetEntitlementsForTask(taskid); err != nil {
		goto ServerError
	}

	c.JSON(http.StatusOK, convertEntitlementEntryFromStorage(entitlements))
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
