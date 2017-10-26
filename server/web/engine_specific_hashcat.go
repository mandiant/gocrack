package web

import (
	"net/http"

	"github.com/fireeye/gocrack/shared"

	"github.com/gin-gonic/gin"
)

func (s *Server) apiHashcatGetTaskModes(c *gin.Context) {
	c.JSON(http.StatusOK, shared.SupportedHashcatModes)
}
