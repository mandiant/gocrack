package shared

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// RecordAPIMetrics logs the duration of an API request, minus websockets as well as a summary of request partitioned by their status code and method
func RecordAPIMetrics(duration prometheus.Summary, counter *prometheus.CounterVec) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// If the request isn't a websocket, observe the duration otherwise the results
		// get skewed
		if c.Request.Header.Get("Upgrade") != "websocket" {
			duration.Observe(float64(time.Since(start)) / float64(time.Microsecond))
		}

		// c.HandlerName() or c.Request.URL.Path doesn't really work well here because it'll contain IDs and pollute the prometheus namespace.
		// We're going to build the apiPath by replacing the parameter value with it's key so that it matches the router definition.
		// Example: /api/v2/task/04a58d59-bc91-48e6-94ea-e47c7fd2a802 -> /api/v2/task/:taskid
		apiPath := c.Request.URL.Path
		for _, param := range c.Params {
			apiPath = strings.Replace(apiPath, param.Value, fmt.Sprintf(":%s", param.Key), 1)
		}

		counter.WithLabelValues(http.StatusText(c.Writer.Status()), c.Request.Method, apiPath).Inc()
	}
}
