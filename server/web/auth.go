package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mandiant/gocrack/server/authentication"
	"github.com/mandiant/gocrack/server/storage"

	"github.com/gin-gonic/gin"
)

type expiredAuth struct {
	Expired bool   `json:"expired"`
	Error   string `json:"error"`
}

// userIsAdmin will prevent a call to the API if the user is not logged in or is not an admin
func userIsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claim := getClaimInformation(c)
		if claim == nil || (claim != nil && !claim.IsAdmin) {
			c.JSON(http.StatusUnauthorized, &WebAPIError{
				StatusCode: http.StatusUnauthorized,
				UserError:  "You are not authorized to view this resource",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) requestHasValidAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCookie, err := c.Cookie("Auth")

		if authCookie == "" || err != nil {
			c.JSON(http.StatusUnauthorized, &WebAPIError{
				StatusCode: http.StatusUnauthorized,
				UserError:  "Username and/or password is incorrect",
			})
			c.Abort()
			return
		}

		claim, err := s.auth.VerifyClaim(authCookie, "gocrack", "api")
		if err != nil {
			if err == authentication.ErrExpired {
				c.JSON(http.StatusUnauthorized, &expiredAuth{
					Expired: true,
					Error:   "Your API key has expired and must be refreshed.",
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusUnauthorized, &WebAPIError{
				StatusCode: http.StatusUnauthorized,
				Err:        err,
				UserError:  "Username and/or password is incorrect",
			})
			c.Abort()
			return
		}

		c.Set("claim", claim)
		c.Next()
		return
	}
}

// LoginRequest defines the structure of a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAPI    bool   `json:"api_only"`
}

func (s LoginRequest) validate() []string {
	errs := make([]string, 0)

	if s.Username == "" {
		errs = append(errs, "username must not be empty")
	}

	if s.Password == "" {
		errs = append(errs, "password must not be empty")
	}
	return errs

}

func (s *Server) webSubmitLogin(c *gin.Context) *WebAPIError {
	var req LoginRequest

	if err := c.BindJSON(&req); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
			UserError:  "Invalid login data",
		}
	}

	if errs := req.validate(); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, APIValidationErrors{
			Valid:  false,
			Errors: errs,
		})
		return nil
	}

	claim, err := s.auth.Login(req.Username, req.Password, req.IsAPI)
	if err != nil || claim == "" {
		invalidLoginCounter.Inc()
		if err == storage.ErrNotFound {
			err = fmt.Errorf("failed login from %s", req.Username)
		}
		return &WebAPIError{
			StatusCode: http.StatusUnauthorized,
			Err:        err,
			UserError:  "Username and/or password is incorrect",
		}
	}

	cookie := http.Cookie{
		Name:     "Auth",
		Value:    claim,
		Expires:  time.Now().UTC().Add(time.Hour * 24),
		HttpOnly: true,
	}

	http.SetCookie(c.Writer, &cookie)
	c.JSON(http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: claim,
	})
	return nil
}
