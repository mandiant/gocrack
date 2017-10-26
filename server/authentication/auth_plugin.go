// auth_plugin.go provides similar functionality as the sql driver registration.
// see https://golang.org/src/database/sql/sql.go for more info

package authentication

import (
	"fmt"
	"sync"
)

var (
	authPluginsMu sync.RWMutex
	authPlugins   = make(map[string]AuthDriver)
)

// PluginSettings contains settings related to a specific authentication plugin
type PluginSettings map[string]interface{}

type AuthDriver interface {
	Open(AuthStorageBackend, PluginSettings) (AuthAPI, error)
}

// Register makes a storage backend available to the system
func Register(name string, plugin AuthDriver) {
	authPluginsMu.Lock()
	defer authPluginsMu.Unlock()

	if plugin == nil {
		panic("cannot register a nil authentication plugin")
	}

	if _, exists := authPlugins[name]; exists {
		panic(fmt.Sprintf("authentication plugin %s already exists", name))
	}

	authPlugins[name] = plugin
}

// Open creates the authentication plugin
func Open(db AuthStorageBackend, cfg AuthSettings) (*AuthWrapper, error) {
	authPluginsMu.RLock()
	selectedPlugin, ok := authPlugins[cfg.Backend]
	authPluginsMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unregistered authentication plugin %s. did you forget to import it?", cfg.Backend)
	}

	authPlugin, err := selectedPlugin.Open(db, cfg.Settings)
	if err != nil {
		return nil, err
	}

	return WrapProvider(authPlugin, cfg), nil
}
