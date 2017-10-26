package rpc

import "github.com/prometheus/client_golang/prometheus"

var (
	crackedCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "gocrack",
			Name:      "cracked_passwords_total",
			Help:      "Number of Cracked Passwords",
		},
	)

	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gocrack",
			Subsystem: "rpc",
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code, HTTP method, and APIMetrics.",
		},
		[]string{"status", "method", "path"},
	)

	requestDuration = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "gocrack",
			Subsystem: "rpc",
			Name:      "request_duration_microseconds",
			Help:      "The HTTP request latencies in microseconds.",
		},
	)
)

func init() {
	prometheus.MustRegister(crackedCounter)
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
}
