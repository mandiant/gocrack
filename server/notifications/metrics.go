package notifications

import "github.com/prometheus/client_golang/prometheus"

var (
	notificationsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gocrack",
			Subsystem: "notifications",
			Name:      "sent_total",
			Help:      "Number of notifications sent, partitioned by type",
		},
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(notificationsSent)
}
