/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"github.com/caicloud/clientset/informers"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/zoumo/register"
)

var (
	plugins = register.NewRegister(nil)
)

// Plugin defines a pluggable proxy interface
type Plugin interface {
	Init(config.Configuration, informers.SharedInformerFactory)
	Run(stopCh <-chan struct{})
	OnSync(*lbapi.LoadBalancer)
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

// Init calls all registered proxy plugins Init func
func Init(c config.Configuration, sif informers.SharedInformerFactory) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		f.Init(c, sif)
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
func OnSync(lb *lbapi.LoadBalancer) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		f.OnSync(lb)
	}
}
