package ginlog

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// LogRequests logs requests using zerolog
func LogRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().UTC()
		path := c.Request.URL.Path

		c.Next()

		rawQ := c.Request.URL.RawQuery

		if rawQ != "" {
			path = path + "?" + rawQ
		}

		log.Info().
			Str("client_ip", c.ClientIP()).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Str("latency", time.Now().UTC().Sub(start).String()).
			Msg("Processed HTTP Request")
	}
}
