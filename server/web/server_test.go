package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Suppress the gin out while testing
	gin.SetMode(gin.ReleaseMode)
}

func TestWrapAPIForError(t *testing.T) {
	e := gin.New()
	e.GET("/test", WrapAPIForError(func(c *gin.Context) *WebAPIError {
		c.String(http.StatusOK, "Hello World")
		return nil
	}))
	e.GET("/broken", WrapAPIForError(func(c *gin.Context) *WebAPIError {
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        errors.New("i am an error"),
			CanErrorBeShownToUser: true,
		}
	}))

	req, _ := http.NewRequest("GET", "/test", nil)

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello World", w.Body.String())

	req, _ = http.NewRequest("GET", "/broken", nil)

	w = httptest.NewRecorder()
	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var ae apiError
	if err := json.NewDecoder(w.Body).Decode(&ae); err != nil {
		assert.Fail(t, "failed to decode", err.Error())
	}
	assert.Equal(t, "i am an error", ae.Error)
}
