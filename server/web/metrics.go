package web

import "github.com/prometheus/client_golang/prometheus"

var (
	invalidLoginCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "gocrack",
			Subsystem: "api",
			Name:      "login_failures_total",
			Help:      "Number of failed logins",
		},
	)

	totalRealtimeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gocrack",
			Subsystem: "realtime",
			Name:      "connections_total",
			Help:      "Total number of active connections to the realtime endpoint ",
		},
	)

	totalRealtimeMessages = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "gocrack",
			Subsystem: "realtime",
			Name:      "messages_sent",
			Help:      "Total number of messages sent to connected realtime users",
		},
	)

	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gocrack",
			Subsystem: "api",
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code, HTTP method, and APIMetrics.",
		},
		[]string{"status", "method", "path"},
	)

	requestDuration = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "gocrack",
			Subsystem: "api",
			Name:      "request_duration_microseconds",
			Help:      "The HTTP request latencies in microseconds.",
		},
	)
)

func init() {
	prometheus.MustRegister(invalidLoginCounter)
	prometheus.MustRegister(totalRealtimeConnections)
	prometheus.MustRegister(totalRealtimeMessages)
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
}
