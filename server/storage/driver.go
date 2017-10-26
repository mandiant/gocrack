// driver.go provides similar functionality as the sql driver registration.
// see https://golang.org/src/database/sql/sql.go for more info

package storage

import (
	"fmt"
	"sync"
)

var (
	backendsMu sync.RWMutex
	backends   = make(map[string]Driver)
)

type Driver interface {
	Open(cfg Config) (Backend, error)
}

// Register makes a storage backend available to the system
func Register(name string, backend Driver) {
	backendsMu.Lock()
	defer backendsMu.Unlock()

	if backend == nil {
		panic("cannot register a nil backend")
	}

	if _, exists := backends[name]; exists {
		panic(fmt.Sprintf("backend %s already exists", name))
	}

	backends[name] = backend
}

func Open(config Config) (Backend, error) {
	backendsMu.RLock()
	selectedBackend, ok := backends[config.Backend]
	backendsMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unregistered storage backend %s. did you forget to import it?", config.Backend)
	}

	return selectedBackend.Open(config)
}
