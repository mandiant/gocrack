package web

import (
	"fmt"
	"net/http"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/gin-gonic/gin"
)

func (s *Server) webDeleteTask(c *gin.Context) *WebAPIError {
	taskid := c.Param("taskid")

	if err := s.stor.DeleteTask(taskid); err != nil {
		if err == storage.ErrNotFound {
			return &WebAPIError{
				UserError:  "Could not delete task as it does not exist",
				StatusCode: http.StatusNotFound,
				Err:        err,
			}
		}

		return &WebAPIError{
			UserError:  "An error occured while trying to delete your task",
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	if err := s.stor.RemoveEntitlements(taskid, storage.EntitlementTask); err != nil {
		return &WebAPIError{
			UserError:  "An error occured while trying to delete your task",
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to clear entitlement entries for %s: %v", taskid, err),
		}
	}

	if err := s.stor.RemoveActivityEntries(taskid); err != nil {
		return &WebAPIError{
			UserError:  "An error occured while trying to delete your task",
			StatusCode: http.StatusInternalServerError,
			Err:        fmt.Errorf("failed to clear activity entries for %s: %v", taskid, err),
		}
	}

	c.Status(http.StatusNoContent)
	return nil
}
