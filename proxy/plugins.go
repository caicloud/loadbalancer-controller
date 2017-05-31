package proxy

import (
	"github.com/caicloud/loadbalancer-controller/api"
	"github.com/caicloud/loadbalancer-controller/pkg/informers"
	"github.com/zoumo/register"
	"gopkg.in/urfave/cli.v1"
)

var (
	plugins = register.NewRegister(nil)
)

// Plugin defines a pluggable proxy interface
type Plugin interface {
	AddFlags(app *cli.App)
	Init(informers.SharedInformerFactory)
	Run(stopCh <-chan struct{})
	OnSync(*api.LoadBalancer)
}

// RegisterPlugin registers a Plugin by name.
// Register does not allow user to override an existing Plugin.
// This is expected to happen during app startup.
func RegisterPlugin(name string, plugin Plugin) {
	plugins.Register(name, plugin)
}

// GetPlugin returns a registered Plugin, or nil if not
func GetPlugin(name string) (Plugin, bool) {
	v, found := plugins.Get(name)
	if !found {
		return nil, false
	}
	return v.(Plugin), true
}

// IsPlugin returns true if name corresponds to an already registered Plugin
func IsPlugin(name string) bool {
	return plugins.Contains(name)
}

// Plugins returns the name of all registered proxy plugin in a string slice
func Plugins() []string {
	return plugins.Keys()
}

// AddFlags calls all registered proxy plugins AddFlags func
func AddFlags(app *cli.App) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		f.AddFlags(app)
	}
}

// Init calls all registered proxy plugins Init func
func Init(sif informers.SharedInformerFactory) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		f.Init(sif)
	}
}

// Run starts all registered proxy plugins
// This is expected to happen after Init.
func Run(stopCh <-chan struct{}) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		go f.Run(stopCh)
	}
}

// OnSync calls all registered proxy plugins OnSync func
func OnSync(lb *api.LoadBalancer) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		f.OnSync(lb)
	}
}
