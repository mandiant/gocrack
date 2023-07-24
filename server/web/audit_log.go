package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mandiant/gocrack/server/storage"

	"github.com/gin-gonic/gin"
)

type storActivityTypeString storage.ActivityType

func (e storActivityTypeString) MarshalJSON() ([]byte, error) {
	var tmp string

	switch storage.ActivityType(e) {
	case storage.ActivtyLogin:
		tmp = "ActivtyLogin"
	case storage.ActivityCreatedTask:
		tmp = "ActivityCreatedTask"
	case storage.ActivityModifiedTask:
		tmp = "ActivityModifiedTask"
	case storage.ActivityDeletedTask:
		tmp = "ActivityDeletedTask"
	case storage.ActivityViewTask:
		tmp = "ActivityViewTask"
	case storage.ActivityViewPasswords:
		tmp = "ActivityViewPasswords"
	case storage.ActivityEntitlementRequest:
		tmp = "ActivityEntitlementRequest"
	case storage.ActivityEntitlementModification:
		tmp = "ActivityEntitlementModification"
	default:
		tmp = fmt.Sprintf("Unknown Action %d", e)
	}

	return json.Marshal(tmp)
}

// ActivityLogEntry should mimick storage.ActivityLogEntry
type ActivityLogEntry struct {
	OccuredAt  time.Time              `json:"occurred_at"`
	UserUUID   string                 `json:"user_id"`
	Username   string                 `json:"username"`
	StatusCode string                 `json:"status_code"`
	Type       storActivityTypeString `json:"type"`
	IPAddress  string                 `json:"ip_address"`
}

func convertStorageActivityLogEntry(entry storage.ActivityLogEntry) ActivityLogEntry {
	return ActivityLogEntry{
		OccuredAt:  entry.OccuredAt,
		UserUUID:   entry.UserUUID,
		Username:   entry.Username,
		StatusCode: http.StatusText(entry.StatusCode),
		Type:       storActivityTypeString(entry.Type),
		IPAddress:  entry.IPAddress,
	}
}

func (s *Server) webGetAuditLog(c *gin.Context) *WebAPIError {
	entityID := c.Param("entityid")

	logs, err := s.stor.GetActivityLog(entityID)
	if err != nil && err != storage.ErrNotFound {
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	}

	if logs == nil {
		c.JSON(http.StatusOK, &[]ActivityLogEntry{})
		return nil
	}

	out := make([]ActivityLogEntry, len(logs))
	for i, log := range logs {
		out[i] = convertStorageActivityLogEntry(log)
	}

	c.JSON(http.StatusOK, &out)
	return nil
}
