package bdb

import (
	"sync"

	"github.com/asdine/storm"
	"github.com/prometheus/client_golang/prometheus"
)

type gauges struct {
	help      string
	subsystem string
}

var metrics = map[string]*gauges{
	"free_pages_total": {
		help: "Total number of free pages on the freelist",
	},
	"pending_pages_total": {
		help: "Total number of pending pages on the freelist",
	},
	"free_alloc_bytes": {
		help: "Total number of bytes allocated in the free pages",
	},
	"free_list_inuse_bytes": {
		help: "Total number of bytes used by the freelist",
	},
	"started_total": {
		help:      "Total number of started read transactions",
		subsystem: "tx",
	},
	"open_total": {
		help:      "Total number of open read transactions",
		subsystem: "tx",
	},
	"page_count_total": {
		help:      "Total number of pages allocated",
		subsystem: "tx",
	},
	"page_alloc_total": {
		help:      "Total number of bytes allocated",
		subsystem: "tx",
	},
	"cursor_count_total": {
		help:      "Total number of cursors created",
		subsystem: "tx",
	},
	"node_count_total": {
		help:      "Total number of node allocations",
		subsystem: "tx",
	},
	"node_deref_total": {
		help:      "Total number of node dereferences",
		subsystem: "tx",
	},
	"rebalances_total": {
		help:      "Total number of node rebalances",
		subsystem: "tx",
	},
	"rebalances_time_seconds": {
		help:      "Total number of seconds spent rebalancing nodes",
		subsystem: "tx",
	},
	"split_count_total": {
		help:      "Total number of nodes split",
		subsystem: "tx",
	},
	"spill_count_total": {
		help:      "Total number of nodes spilled",
		subsystem: "tx",
	},
	"spill_time_seconds": {
		help:      "Total number of seconds spent spilling nodes",
		subsystem: "tx",
	},
	"writes_total": {
		help:      "Total number of writes to disk performed",
		subsystem: "tx",
	},
	"write_time_seconds": {
		help:      "Total number of seconds spent writing to disk",
		subsystem: "tx",
	},
}

// StatsExporter is used to export BoltDB stats to Prometheus
type StatsExporter struct {
	// unexported fields below
	mu     sync.Mutex
	db     *storm.DB
	gauges map[string]prometheus.Gauge
}

// NewExporter creates a new StatsExporter for prometheus to allow us to peek at how BoltDB is performing
func NewExporter(db *storm.DB) *StatsExporter {
	exporter := &StatsExporter{
		db:     db,
		gauges: make(map[string]prometheus.Gauge, 0),
	}

	for name, info := range metrics {
		exporter.gauges[name] = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "storage_bdb",
			Subsystem: info.subsystem,
			Name:      name,
			Help:      info.help,
		})
	}
	return exporter
}

// Describe describes all the metrics exported by BoltDB
func (s *StatsExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, vec := range s.gauges {
		vec.Describe(ch)
	}
}

func (s *StatsExporter) updateMetrics() {
	globalStats := s.db.Bolt.Stats()

	s.gauges["free_pages_total"].Set(float64(globalStats.FreePageN))
	s.gauges["pending_pages_total"].Set(float64(globalStats.PendingPageN))
	s.gauges["free_alloc_bytes"].Set(float64(globalStats.FreeAlloc))
	s.gauges["free_list_inuse_bytes"].Set(float64(globalStats.FreelistInuse))
	s.gauges["started_total"].Set(float64(globalStats.TxN))
	s.gauges["open_total"].Set(float64(globalStats.OpenTxN))
	s.gauges["page_count_total"].Set(float64(globalStats.TxStats.PageCount))
	s.gauges["page_alloc_total"].Set(float64(globalStats.TxStats.PageAlloc))
	s.gauges["cursor_count_total"].Set(float64(globalStats.TxStats.CursorCount))
	s.gauges["node_count_total"].Set(float64(globalStats.TxStats.NodeCount))
	s.gauges["node_deref_total"].Set(float64(globalStats.TxStats.NodeDeref))
	s.gauges["rebalances_total"].Set(float64(globalStats.TxStats.Rebalance))
	s.gauges["rebalances_time_seconds"].Set(globalStats.TxStats.RebalanceTime.Seconds())
	s.gauges["split_count_total"].Set(float64(globalStats.TxStats.Split))
	s.gauges["spill_count_total"].Set(float64(globalStats.TxStats.Spill))
	s.gauges["spill_time_seconds"].Set(globalStats.TxStats.SpillTime.Seconds())
	s.gauges["writes_total"].Set(float64(globalStats.TxStats.Write))
	s.gauges["write_time_seconds"].Set(globalStats.TxStats.WriteTime.Seconds())
}

// Collect retrieves the current BoltDB statistics
func (s *StatsExporter) Collect(ch chan<- prometheus.Metric) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.updateMetrics()

	for _, vec := range s.gauges {
		vec.Collect(ch)
	}
}
