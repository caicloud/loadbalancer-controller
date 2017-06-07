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

package provider

import (
	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	"github.com/caicloud/loadbalancer-controller/pkg/informers"
	"github.com/zoumo/register"
	"gopkg.in/urfave/cli.v1"
)

var (
	plugins = register.NewRegister(nil)
)

// Plugin defines a pluggable provider interface
type Plugin interface {
	AddFlags(app *cli.App)
	Init(informers.SharedInformerFactory)
	Run(stopCh <-chan struct{})
	OnSync(*netv1alpha1.LoadBalancer)
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

// Plugins returns the name of all registered provider plugin in a string slice
func Plugins() []string {
	return plugins.Keys()
}

// AddFlags calls all registered provider plugins AddFlags func
func AddFlags(app *cli.App) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		f.AddFlags(app)
	}
}

// Init calls all registered provider plugins Init func
func Init(sif informers.SharedInformerFactory) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		f.Init(sif)
	}
}

// Run starts all registered provider plugins
// This is expected to happen after Init.
func Run(stopCh <-chan struct{}) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		go f.Run(stopCh)
	}
}

// OnSync calls all registered provider plugins OnSync func
func OnSync(lb *netv1alpha1.LoadBalancer) {
	for _, v := range plugins.Iter() {
		f := v.(Plugin)
		f.OnSync(lb)
	}
}
