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

package plugin

import (
	"sync"

	"github.com/caicloud/clientset/informers"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
)

// Interface defines a pluggable proxy interface
type Interface interface {
	Init(config.Configuration, informers.SharedInformerFactory)
	Run(stopCh <-chan struct{})
	OnSync(*lbapi.LoadBalancer)
}

// Registry ...
type Registry struct {
	data  map[string]interface{}
	mutex sync.RWMutex
}

// NewRegistry ...
func NewRegistry() *Registry {
	return &Registry{
		data: make(map[string]interface{}),
	}
}

// Register registers a Interface into registry by name.
// Register does not allow user to override an existing Interface.
// This is expected to happen during app startup
func (r *Registry) Register(name string, plugin Interface) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.data[name]; ok {
		panic("Repeated registration key" + name)
	}
	r.data[name] = plugin
}

// Get returns a registered Interface, or nil if not
func (r *Registry) Get(name string) (Interface, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if v, ok := r.data[name]; ok {
		return v.(Interface), true
	}
	return nil, false
}

// Contains checks if the plugin's name is already registered
func (r *Registry) Contains(name string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.data[name]
	return ok
}

// AllInterfaces returns all registered plugins' names
func (r *Registry) AllInterfaces() []string {
	names := []string{}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for name := range r.data {
		names = append(names, name)
	}
	return names
}

func (r *Registry) rangeItems(f func(key string, value interface{}) bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for k, v := range r.data {
		f(k, v)
	}
}

// InitAll calls all registered plugins OnSync function
func (r *Registry) InitAll(c config.Configuration, sif informers.SharedInformerFactory) {
	r.rangeItems(func(k string, value interface{}) bool {
		plugin := value.(Interface)
		plugin.Init(c, sif)
		return true
	})
}

// RunAll calls all registered plugins' Run function
func (r *Registry) RunAll(stopCh <-chan struct{}) {
	r.rangeItems(func(k string, value interface{}) bool {
		plugin := value.(Interface)
		go plugin.Run(stopCh)
		return true
	})
}

// SyncAll calls all registered plugins OnSync function
func (r *Registry) SyncAll(lb *lbapi.LoadBalancer) {
	r.rangeItems(func(k string, value interface{}) bool {
		plugin := value.(Interface)
		go plugin.OnSync(lb)
		return true
	})
}
