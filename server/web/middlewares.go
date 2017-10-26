package web

import (
	"net/http"
	"time"

	"github.com/fireeye/gocrack/server/storage"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

// logAction records potentially sensitive actions to the database for auditing purposes
func (s *Server) logAction(action storage.ActivityType, entityID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claim := getClaimInformation(c)
		record := storage.ActivityLogEntry{
			OccuredAt: time.Now().UTC(),
			UserUUID:  claim.UserUUID,
			Type:      action,
			EntityID:  c.Param(entityID),
			Path:      c.Request.URL.EscapedPath(),
			IPAddress: c.ClientIP(),
		}

		c.Next()

		record.StatusCode = c.Writer.Status()
		if err := s.stor.LogActivity(record); err != nil {
			log.Error().Interface("record", record).Err(err).Msg("Failed to write activity log to database")
		}
	}
}

// checkIfUserIsAdmin ensures the user is an administrator before allowing the rest of the chain to continue
func checkIfUserIsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claim := getClaimInformation(c)
		if !claim.IsAdmin {
			c.JSON(http.StatusUnauthorized, &WebAPIError{
				StatusCode: http.StatusUnauthorized,
				UserError:  "You do not have permissions to access this route",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// setSecureHeaders sets some standard secure headers to the responses
func setSecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Option", "nosniff")

		c.Next()
	}
}

// setXSRFTokenIfNecessary will grab the CSRF token for a request and set it as a response cookie if needed
func setXSRFTokenIfNecessary(enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			csrf.UnsafeSkipCheck(c.Request)
			c.Next()
			return
		}

		claim := getClaimInformation(c)
		if claim == nil {
			goto SetCookie
		}

		// If the claim is strictly API use, enable a CSRF bypass
		if claim.APIOnly {
			csrf.UnsafeSkipCheck(c.Request)
			c.Next()
			return
		}

	SetCookie:
		// Set the XSRF token so that the UI can get access to it
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "XSRF-TOKEN",
			Value:    csrf.Token(c.Request),
			HttpOnly: false,
			Path:     "/",
			Secure:   c.Request.URL.Scheme == "https",
		})
		c.Next()
	}
}

// checkIfUserIsEntitled ensures the user is able to access the document.
func (s *Server) checkIfUserIsEntitled(entityIDLookup string, documentType storage.EntitlementType) gin.HandlerFunc {
	return func(c *gin.Context) {
		claim := getClaimInformation(c)

		// If they're not an admin, let's look them up to ensure they're entitled to the document
		if !claim.IsAdmin {
			canAccess, err := s.stor.CheckEntitlement(claim.UserUUID, c.Param(entityIDLookup), documentType)
			if err != nil {
				if err == storage.ErrNotFound {
					goto CantAccess
				}
				c.JSON(http.StatusInternalServerError, &WebAPIError{
					StatusCode: http.StatusInternalServerError,
					Err:        err,
				})
				c.Abort()
				return
			}

			if canAccess {
				c.Next()
				return
			}

			goto CantAccess
		CantAccess:
			c.JSON(http.StatusNotFound, &WebAPIError{
				StatusCode: http.StatusNotFound,
				Err:        err,
				UserError:  "The requested file does not exist or you do not have permissions to it",
			})
			c.Abort()
			return
		}

		// Continue if all is well
		c.Next()
	}
}

// checkParamValidUUID ensures the UUID in the HTTP route is valid
func checkParamValidUUID(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid := c.Param(paramName); uid != "" {
			if _, err := uuid.FromString(uid); err != nil {
				goto Error
			}
			// UUID checks out
			c.Next()
			return
		}

		// Will fallthrough here if uid is ""
	Error:
		c.JSON(http.StatusBadRequest, &WebAPIError{
			StatusCode:            http.StatusBadRequest,
			CanErrorBeShownToUser: false,
			UserError:             "The UUID in the HTTP Path is not valid",
		})
		c.Abort()

	}
}
