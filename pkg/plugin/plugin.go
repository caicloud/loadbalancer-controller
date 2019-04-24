package plugin

import (
	"github.com/caicloud/clientset/informers"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/zoumo/golib/register"
)

// Interface defines a pluggable proxy interface
type Interface interface {
	Init(config.Configuration, informers.SharedInformerFactory)
	Run(stopCh <-chan struct{})
	OnSync(*lbapi.LoadBalancer)
}

// Registry ...
type Registry struct {
	register *register.Register
}

// NewRegistry ...
func NewRegistry() *Registry {
	return &Registry{
		register: register.New(nil),
	}
}

// Register registers a Interface into registry by name.
// Register does not allow user to override an existing Interface.
// This is expected to happen during app startup
func (p *Registry) Register(name string, plugin Interface) {
	p.register.Register(name, plugin)
}

// Get returns a registered Interface, or nil if not
func (p *Registry) Get(name string) (Interface, bool) {
	v, found := p.register.Get(name)
	if !found {
		return nil, false
	}
	return v.(Interface), true
}

// Contains checks if the plugin's name is already registered
func (p *Registry) Contains(name string) bool {
	return p.register.Contains(name)
}

// AllInterfaces returns all registered plugins' names
func (p *Registry) AllInterfaces() []string {
	return p.register.Keys()
}

// InitAll calls all registered plugins OnSync function
func (p *Registry) InitAll(c config.Configuration, sif informers.SharedInformerFactory) {
	p.register.Range(func(k string, value interface{}) bool {
		plugin := value.(Interface)
		plugin.Init(c, sif)
		return true
	})
}

// RunAll calls all registered plugins' Run function
func (p *Registry) RunAll(stopCh <-chan struct{}) {
	p.register.Range(func(k string, value interface{}) bool {
		plugin := value.(Interface)
		go plugin.Run(stopCh)
		return true
	})
}

// SyncAll calls all registered plugins OnSync function
func (p *Registry) SyncAll(lb *lbapi.LoadBalancer) {
	p.register.Range(func(k string, value interface{}) bool {
		plugin := value.(Interface)
		go plugin.OnSync(lb)
		return true
	})
}
