package bdb

import (
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestBDBStatsExporter(t *testing.T) {
	db := initTest(t)
	exporter := NewExporter(db.BoltBackend.db)

	metrics := make([]prometheus.Metric, 0)
	descriptions := make([]*prometheus.Desc, 0)

	chMetrics := make(chan prometheus.Metric, 100)
	chDescriptions := make(chan *prometheus.Desc, 100)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(metrics *[]prometheus.Metric) {
		defer wg.Done()

		for m := range chMetrics {
			*metrics = append(*metrics, m)
		}
	}(&metrics)

	wg.Add(1)
	go func(descs *[]*prometheus.Desc) {
		defer wg.Done()

		for m := range chDescriptions {
			*descs = append(*descs, m)
		}
	}(&descriptions)

	exporter.Collect(chMetrics)
	exporter.Describe(chDescriptions)

	close(chMetrics)
	close(chDescriptions)

	wg.Wait()

	GaugesHaveAtLeast1SetMetrics := false
	for _, metric := range metrics {
		m := &dto.Metric{}
		if err := metric.Write(m); err != nil {
			t.Fatal(err)
		}

		if val := m.Gauge.GetValue(); val > 0 {
			GaugesHaveAtLeast1SetMetrics = true
		}
	}

	if !GaugesHaveAtLeast1SetMetrics {
		t.Fatal("Expected at least one of the gauges to have a set metric")
	}

	for _, description := range descriptions {
		if description.String() == "" {
			t.Fatal("description isnt set")
		}
	}
}
