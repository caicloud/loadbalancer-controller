package v1beta1

import (
	"time"

	"github.com/caicloud/loadbalancer-controller/api"
	"github.com/caicloud/loadbalancer-controller/pkg/informers/internalinterfaces"
	"github.com/caicloud/loadbalancer-controller/pkg/listers/caicloud/v1beta1"
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
	Lister() v1beta1.LoadBalancerLister
}

type loadBalancerInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

func newLoadBalancerInformer(client tprclient.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {

	sharedIndexInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.CaicloudV1beta1().LoadBalancers(v1.NamespaceAll).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.CaicloudV1beta1().LoadBalancers(v1.NamespaceAll).Watch(options)
			},
		},
		&api.LoadBalancer{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
	return sharedIndexInformer
}

func (f *loadBalancerInformer) Informer() cache.SharedIndexInformer {
	return f.factory.TPRInformerFor(&api.LoadBalancer{}, newLoadBalancerInformer)
}

func (f *loadBalancerInformer) Lister() v1beta1.LoadBalancerLister {
	return v1beta1.NewLoadBalancerLister(f.Informer().GetIndexer())
}
