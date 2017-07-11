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

package informers

import (
	"reflect"
	"sync"
	"time"

	informerinternal "github.com/caicloud/loadbalancer-controller/pkg/informers/internalinterfaces"
	"github.com/caicloud/loadbalancer-controller/pkg/informers/networking"
	"github.com/caicloud/loadbalancer-controller/pkg/tprclient"
	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/apps"
	"k8s.io/client-go/informers/autoscaling"
	"k8s.io/client-go/informers/batch"
	"k8s.io/client-go/informers/certificates"
	"k8s.io/client-go/informers/core"
	"k8s.io/client-go/informers/extensions"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/informers/policy"
	"k8s.io/client-go/informers/rbac"
	"k8s.io/client-go/informers/settings"
	"k8s.io/client-go/informers/storage"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

var _ SharedInformerFactory = &sharedInformerFactory{}

type sharedInformerFactory struct {
	client    kubernetes.Interface
	tprclient tprclient.Interface

	lock          sync.Mutex
	defaultResync time.Duration

	informers map[reflect.Type]cache.SharedIndexInformer
	// startedInformers is used for tracking which informers have been started.
	// This allows Start() to be called multiple times safely.
	startedInformers map[reflect.Type]bool
}

// NewSharedInformerFactory constructs a new instance of sharedInformerFactory
func NewSharedInformerFactory(client kubernetes.Interface, tprclient tprclient.Interface, defaultResync time.Duration) SharedInformerFactory {
	return &sharedInformerFactory{
		client:           client,
		tprclient:        tprclient,
		defaultResync:    defaultResync,
		informers:        make(map[reflect.Type]cache.SharedIndexInformer),
		startedInformers: make(map[reflect.Type]bool),
	}
}

// Start initializes all requested informers.
func (f *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for informerType, informer := range f.informers {
		if !f.startedInformers[informerType] {
			log.Debug("Startting informer", log.Fields{"type": informerType})
			go informer.Run(stopCh)
			f.startedInformers[informerType] = true
		}
	}
}

// WaitForCacheSync waits for all started informers' cache were synced.
func (f *sharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool {
	informers := func() map[reflect.Type]cache.SharedIndexInformer {
		f.lock.Lock()
		defer f.lock.Unlock()

		informers := map[reflect.Type]cache.SharedIndexInformer{}
		for informerType, informer := range f.informers {
			if f.startedInformers[informerType] {
				informers[informerType] = informer
			}
		}
		return informers
	}()

	res := map[reflect.Type]bool{}
	for informType, informer := range informers {
		res[informType] = cache.WaitForCacheSync(stopCh, informer.HasSynced)
		log.Debug("Cache has synced", log.Fields{"type": informType, "result": res[informType]})

	}
	return res
}

// InternalInformerFor returns the SharedIndexInformer for obj using an internal
// client.
func (f *sharedInformerFactory) InformerFor(obj runtime.Object, newFunc internalinterfaces.NewInformerFunc) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informerType := reflect.TypeOf(obj)
	informer, exists := f.informers[informerType]
	if exists {
		return informer
	}
	informer = newFunc(f.client, f.defaultResync)
	f.informers[informerType] = informer

	return informer
}

// InternalInformerFor returns the SharedIndexInformer for obj using an tpr rest
// client.
func (f *sharedInformerFactory) TPRInformerFor(obj runtime.Object, newFunc informerinternal.NewTPRInformerFunc) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informerType := reflect.TypeOf(obj)
	informer, exists := f.informers[informerType]
	if exists {
		return informer
	}

	informer = newFunc(f.tprclient, f.defaultResync)
	f.informers[informerType] = informer

	return informer
}

// Client returns kubernetes clientset
func (f *sharedInformerFactory) Client() kubernetes.Interface {
	return f.client
}

// Client returns kubernetes clientset
func (f *sharedInformerFactory) TPRClient() tprclient.Interface {
	return f.tprclient
}

// SharedInformerFactory provides shared informers for resources in all known
// API group versions.
type SharedInformerFactory interface {
	informerinternal.SharedInformerFactory
	Client() kubernetes.Interface
	TPRClient() tprclient.Interface

	ForResource(resource schema.GroupVersionResource) (informers.GenericInformer, error)
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool

	Apps() apps.Interface
	Autoscaling() autoscaling.Interface
	Batch() batch.Interface
	Certificates() certificates.Interface
	Core() core.Interface
	Extensions() extensions.Interface
	Policy() policy.Interface
	Rbac() rbac.Interface
	Settings() settings.Interface
	Storage() storage.Interface

	// TPR
	Networking() networking.Interface
}

func (f *sharedInformerFactory) Apps() apps.Interface {
	return apps.New(f)
}

func (f *sharedInformerFactory) Autoscaling() autoscaling.Interface {
	return autoscaling.New(f)
}

func (f *sharedInformerFactory) Batch() batch.Interface {
	return batch.New(f)
}

func (f *sharedInformerFactory) Certificates() certificates.Interface {
	return certificates.New(f)
}

func (f *sharedInformerFactory) Core() core.Interface {
	return core.New(f)
}

func (f *sharedInformerFactory) Extensions() extensions.Interface {
	return extensions.New(f)
}

func (f *sharedInformerFactory) Policy() policy.Interface {
	return policy.New(f)
}

func (f *sharedInformerFactory) Rbac() rbac.Interface {
	return rbac.New(f)
}

func (f *sharedInformerFactory) Settings() settings.Interface {
	return settings.New(f)
}

func (f *sharedInformerFactory) Storage() storage.Interface {
	return storage.New(f)
}

// Third Party Resource
func (f *sharedInformerFactory) Networking() networking.Interface {
	return networking.New(f)
}
