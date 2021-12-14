package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fireeye/gocrack/server/authentication"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInternal_checkIfUserIsAdmin(t *testing.T) {
	e := gin.New()
	isDummyAdmin := func(isAdmin bool) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("claim", &authentication.AuthClaim{
				IsAdmin: isAdmin,
			})
			c.Next()
		}
	}

	e.GET("/protected/dummy", isDummyAdmin(true), checkIfUserIsAdmin(), func(c *gin.Context) {
		c.String(http.StatusOK, "Hello from Protected Page!")
	})

	e.GET("/protected", isDummyAdmin(false), checkIfUserIsAdmin(), func(c *gin.Context) {
		c.String(http.StatusOK, "Hello from Protected Page!")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected/dummy", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello from Protected Page!", w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/protected", nil)
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestInternal_checkParamValidUUID(t *testing.T) {
	e := gin.New()
	e.GET("/test/:uuid", checkParamValidUUID("uuid"), func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("Hello %s", c.Param("uuid")))
	})

	// Make a bad test
	req, _ := http.NewRequest("GET", "/test/bad", nil)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var ae apiError
	if err := json.NewDecoder(w.Body).Decode(&ae); err != nil {
		assert.Fail(t, "failed to decode", err.Error())
	}
	assert.Equal(t, "The UUID in the HTTP Path is not valid", ae.Error)

	// Make a good one!
	req, _ = http.NewRequest("GET", fmt.Sprintf("/test/%s", uuid.NewString()), nil)

	w = httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
