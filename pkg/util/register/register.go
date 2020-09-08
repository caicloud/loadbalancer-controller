package register

import (
	"sync"
)

var (
	defaultConfig = &Config{
		OverrideAllowed: false,
	}
)

// Register is a struct binding name and interface such as Constructor
type Register struct {
	data            map[string]interface{}
	mu              sync.RWMutex
	overrideAllowed bool
}

// Config is a struct containing all config for register
type Config struct {
	// OverrideAllowed allows the register to override
	// an already registered interface by name if it is true,
	// otherwise register will panic.
	OverrideAllowed bool
}

// New returns a new register
func New(config *Config) *Register {
	if config == nil {
		config = defaultConfig
	}

	return &Register{
		data:            make(map[string]interface{}),
		overrideAllowed: config.OverrideAllowed,
	}
}

// Register registers a interface by name.
// It will panic if name corresponds to an already registered interface
// and the register does not allow user to override the interface.
func (r *Register) Register(name string, v interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[name]; ok && !r.overrideAllowed {
		panic("[Register] Repeated registration key: " + name)
	}
	r.data[name] = v
}

// Get returns an interface registered with the given name
func (r *Register) Get(name string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.data[name]
	return v, ok
}

// Contains returns true if name corresponds to an already registered interface
func (r *Register) Contains(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.data[name]
	return ok
}

// Range calls f sequentially for each key and value present in the register.
// If f returns false, range stops the iteration.
func (r *Register) Range(f func(key string, value interface{}) bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, v := range r.data {
		if ok := f(k, v); !ok {
			return
		}
	}
}

// Keys returns the name of all registered interfaces
func (r *Register) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.data {
		names = append(names, name)
	}
	return names
}

// Values returns all registered interfaces
func (r *Register) Values() []interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var ret []interface{}
	for _, v := range r.data {
		ret = append(ret, v)
	}
	return ret
}

// KeyValues returns an iterable map for <name, interface> pair
func (r *Register) KeyValues() map[string]interface{} {
	return r.data
}

// Clear cleans up the registered items
func (r *Register) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data = make(map[string]interface{})
}
