package bdb

import (
	"github.com/mandiant/gocrack/server/storage"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/asdine/storm"
)

// CurrentStorageVersion describes what storage version we're on and can be used to determine
// if the underlying database structure needs to change
const CurrentStorageVersion = 1.0

var bucketInternalConfig = []byte("config")

func init() {
	storage.Register("bdb", &Driver{})
}

type Driver struct{}

func (s *Driver) Open(cfg storage.Config) (storage.Backend, error) {
	return Init(cfg)
}

// BoltBackend is a storage backend for GoCrack built ontop of Bolt/LMDB
type BoltBackend struct {
	// contains filtered or unexported fields
	db       *storm.DB
	expstats *StatsExporter
}

// Init creates a database instance backed by boltdb
func Init(cfg storage.Config) (storage.Backend, error) {
	db, err := storm.Open(cfg.ConnectionString)
	if err != nil {
		return nil, err
	}

	bb := &BoltBackend{
		db:       db,
		expstats: NewExporter(db),
	}
	if err = bb.checkSchema(); err != nil {
		return nil, err
	}

	// this shouldnt panic but it will if two storage engines are initialized (which we dont want anyways)
	prometheus.MustRegister(bb.expstats)

	return bb, nil
}

// Close the resources used by bolt
func (s *BoltBackend) Close() error {
	if s.expstats != nil {
		// Remove the boltdb exporter if it was enabled
		prometheus.Unregister(s.expstats)
	}
	return s.db.Close()
}

func (s *BoltBackend) checkSchema() error {
	var opt interface{}
	// check our internal version
	if err := s.db.Get("options", "version", &opt); err != nil {
		if err == storm.ErrNotFound {
			if err := s.db.Set("options", "version", CurrentStorageVersion); err != nil {
				return err
			}
			opt = CurrentStorageVersion
			return nil
		}
	}
	// XXX(cschmitt): What do if version changes?
	return nil
}
