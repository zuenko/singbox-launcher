package xray

import (
	"fmt"
	"sync"
)

// SidecarRegistry maps sing-box outbound tags to sidecar configuration.
type SidecarRegistry struct {
	mu      sync.RWMutex
	entries map[string]*RegistryEntry
	basePort int
	nextOffset int
}

// RegistryEntry holds the original outbound and assigned port for one sidecar.
type RegistryEntry struct {
	Tag              string
	OriginalOutbound map[string]interface{}
	Port             int
}

// NewRegistry creates a new registry with the given base port (e.g. 15080).
func NewRegistry(basePort int) *SidecarRegistry {
	return &SidecarRegistry{
		entries:  make(map[string]*RegistryEntry),
		basePort: basePort,
	}
}

// Register adds an outbound to the registry and returns the assigned port
// plus a SOCKS outbound map that can be inserted into sing-box config.
func (r *SidecarRegistry) Register(tag string, original map[string]interface{}) (int, map[string]interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.entries[tag]; ok {
		// Re-use existing port if already registered
		return existing.Port, r.socksOutbound(tag, existing.Port)
	}

	port := r.basePort + r.nextOffset
	r.nextOffset++

	r.entries[tag] = &RegistryEntry{
		Tag:              tag,
		OriginalOutbound: original,
		Port:             port,
	}

	return port, r.socksOutbound(tag, port)
}

// Get returns the registry entry for a tag, or nil if not registered.
func (r *SidecarRegistry) Get(tag string) (*RegistryEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[tag]
	return e, ok
}

// HasXHTTP reports whether any registered outbound uses xhttp transport.
func (r *SidecarRegistry) HasXHTTP() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.entries {
		if tr, ok := e.OriginalOutbound["transport"].(map[string]interface{}); ok {
			if registryMapGetString(tr, "type") == "xhttp" {
				return true
			}
		}
	}
	return false
}

// Tags returns all registered tags.
func (r *SidecarRegistry) Tags() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tags := make([]string, 0, len(r.entries))
	for t := range r.entries {
		tags = append(tags, t)
	}
	return tags
}

// Clear removes all entries.
func (r *SidecarRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = make(map[string]*RegistryEntry)
	r.nextOffset = 0
}

// Len returns the number of registered entries.
func (r *SidecarRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// socksOutbound builds a sing-box socks outbound pointing to the sidecar.
func (r *SidecarRegistry) socksOutbound(tag string, port int) map[string]interface{} {
	return map[string]interface{}{
		"tag":         tag,
		"type":        "socks",
		"server":      "127.0.0.1",
		"server_port": port,
		"version":     "5",
	}
}

// registryMapGetString is a local copy to avoid duplicate symbol.
func registryMapGetString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch s := v.(type) {
		case string:
			return s
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}
