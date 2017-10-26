package shared

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func createMetrics() (prometheus.Summary, *prometheus.CounterVec) {
	return prometheus.NewSummary(
			prometheus.SummaryOpts{
				Name: "request_duration_microseconds",
				Help: "The HTTP request latencies in microseconds.",
			},
		),
		prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "requests_total",
				Help: "How many HTTP requests processed, partitioned by status code, HTTP method, and APIMetrics.",
			},
			[]string{"status", "method", "path"},
		)
}

func createTestEngine(middlewares ...gin.HandlerFunc) *gin.Engine {
	e := gin.New()
	e.Use(middlewares...)
	e.GET("/test/:testid", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("Hello Test %s", c.Param("testid")))
	})
	return e
}

func TestRecordAPIMetrics(t *testing.T) {
	sum, counter := createMetrics()
	server := createTestEngine(RecordAPIMetrics(sum, counter))
	req, _ := http.NewRequest("GET", "/test/1337-w00t", nil)

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	sumMetric := &dto.Metric{}
	sum.Write(sumMetric)
	assert.Equal(t, uint64(1), sumMetric.GetSummary().GetSampleCount())

	counterMetric := &dto.Metric{}
	counter.WithLabelValues("OK", "GET", "/test/:testid").Write(counterMetric)
	assert.Equal(t, float64(1), counterMetric.Counter.GetValue())
}
