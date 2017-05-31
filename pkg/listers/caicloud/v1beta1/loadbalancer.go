package v1beta1

import (
	"github.com/caicloud/loadbalancer-controller/api"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// LoadBalancerLister helps list LoadBalancer.
type LoadBalancerLister interface {
	// List lists all LoadBalancer in the indexer.
	List(selector labels.Selector) (ret []*api.LoadBalancer, err error)
	// LoadBalancers returns an object that can list and get LoadBalancer.
	LoadBalancers(namespace string) LoadBalancerNamespaceLister
	LoadBalancerListerExpansion
}

// loadBalancerList implements the LoadBalancerLister interface.
type loadBalancerList struct {
	indexer cache.Indexer
}

// NewLoadBalancerLister returns a new LoadBalancerLister.
func NewLoadBalancerLister(indexer cache.Indexer) LoadBalancerLister {
	return &loadBalancerList{indexer: indexer}
}

// List lists all LoadBalancer in the indexer.
func (s *loadBalancerList) List(selector labels.Selector) (ret []*api.LoadBalancer, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*api.LoadBalancer))
	})

	return ret, err
}

// LoadBalancers returns an object that can list and get LoadBalancer.
func (s *loadBalancerList) LoadBalancers(namespace string) LoadBalancerNamespaceLister {
	return loadBalancerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// LoadBalancerNamespaceLister helps list and get LoadBalancers.
type LoadBalancerNamespaceLister interface {
	// List lists all LoadBalancers in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*api.LoadBalancer, err error)
	// Get retrieves the LoadBalancer from the indexer for a given namespace and name.
	Get(name string) (*api.LoadBalancer, error)
	LoadBalancerNamespaceListerExpansion
}

// loadBalancerNamespaceLister implements the loadBalancerNamespaceLister
// interface.
type loadBalancerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all loadBalancers in the indexer for a given namespace.
func (s loadBalancerNamespaceLister) List(selector labels.Selector) (ret []*api.LoadBalancer, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*api.LoadBalancer))
	})
	return ret, err
}

// Get retrieves the loadBalancer from the indexer for a given namespace and name.
func (s loadBalancerNamespaceLister) Get(name string) (*api.LoadBalancer, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(api.Resource(api.LoadBalancerName), name)
	}
	return obj.(*api.LoadBalancer), nil
}
