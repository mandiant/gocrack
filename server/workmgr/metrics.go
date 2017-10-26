package workmgr

import "github.com/prometheus/client_golang/prometheus"

var (
	activeSubscriptions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gocrack",
			Subsystem: "workmgr",
			Name:      "subscriptions_total",
			Help:      "Number of subscribers to the internal pub/sub system by topic",
		},
		[]string{"topic"},
	)

	broadcastsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gocrack",
			Subsystem: "workmgr",
			Name:      "broadcasts_sent",
			Help:      "Number of broadcasts sent to the internal pub/sub system by topic",
		},
		[]string{"topic"},
	)

	connectedWorkers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "gocrack",
			Subsystem: "workmgr",
			Name:      "workers_total",
			Help:      "Number of connected and active workers",
		},
	)
)

func init() {
	prometheus.MustRegister(activeSubscriptions)
	prometheus.MustRegister(broadcastsSent)
	prometheus.MustRegister(connectedWorkers)
}
