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

package v1alpha1

import (
	"time"

	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	"github.com/caicloud/loadbalancer-controller/pkg/informers/internalinterfaces"
	netlisters "github.com/caicloud/loadbalancer-controller/pkg/listers/networking/v1alpha1"
	"github.com/caicloud/loadbalancer-controller/pkg/tprclient"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// LoadBalancerInformer provides access to a shared informer and lister for
// LoadBalancers.
type LoadBalancerInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() netlisters.LoadBalancerLister
}

type loadBalancerInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

func newLoadBalancerInformer(client tprclient.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {

	sharedIndexInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.NetworkingV1alpha1().LoadBalancers(v1.NamespaceAll).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.NetworkingV1alpha1().LoadBalancers(v1.NamespaceAll).Watch(options)
			},
		},
		&netv1alpha1.LoadBalancer{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	return sharedIndexInformer
}

func (f *loadBalancerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.TPRInformerFor(&netv1alpha1.LoadBalancer{}, newLoadBalancerInformer)
}

func (f *loadBalancerInformer) Lister() netlisters.LoadBalancerLister {
	return netlisters.NewLoadBalancerLister(f.Informer().GetIndexer())
}
