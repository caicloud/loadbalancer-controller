package v1beta1

import (
	"github.com/caicloud/loadbalancer-controller/api"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// LoadBalacnersGetter has a method to return a LoadBalacnerInterface.
// A group's client should implement this interface.
type LoadBalacnersGetter interface {
	LoadBalancers(namespace string) LoadBalancerInterface
}

type LoadBalancerInterface interface {
	Create(*api.LoadBalancer) (*api.LoadBalancer, error)
	Update(*api.LoadBalancer) (*api.LoadBalancer, error)
	UpdateStatus(*api.LoadBalancer) (*api.LoadBalancer, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*api.LoadBalancer, error)
	List(opts v1.ListOptions) (*api.LoadBalancerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *api.LoadBalancer, err error)
	LoadBalancerExpansion
}

var _ LoadBalancerInterface = &loadbalancers{}

type loadbalancers struct {
	client rest.Interface
	ns     string
}

func newLoadBalancers(c *CaicloudV1beta1Client, namespace string) *loadbalancers {
	return &loadbalancers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a loadbalancer and creates it.  Returns the server's representation of the loadbalancer, and an error, if there is any.
func (c *loadbalancers) Create(lb *api.LoadBalancer) (result *api.LoadBalancer, err error) {
	result = &api.LoadBalancer{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		Body(lb).
		Do().
		Into(result)
	return
}

// Update takes the representation of a loadbalancer and updates it. Returns the server's representation of the loadbalancer, and an error, if there is any.
func (c *loadbalancers) Update(lb *api.LoadBalancer) (result *api.LoadBalancer, err error) {
	result = &api.LoadBalancer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		Name(lb.Name).
		Body(lb).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

func (c *loadbalancers) UpdateStatus(lb *api.LoadBalancer) (result *api.LoadBalancer, err error) {
	result = &api.LoadBalancer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		Name(lb.Name).
		SubResource("status").
		Body(lb).
		Do().
		Into(result)
	return
}

// Delete takes name of the loadbalancer and deletes it. Returns an error if one occurs.
func (c *loadbalancers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *loadbalancers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the loadbalancer, and returns the corresponding loadbalancer object, and an error if there is any.
func (c *loadbalancers) Get(name string, options v1.GetOptions) (result *api.LoadBalancer, err error) {
	result = &api.LoadBalancer{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of LoadBalancers that match those selectors.
func (c *loadbalancers) List(opts v1.ListOptions) (result *api.LoadBalancerList, err error) {
	result = &api.LoadBalancerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested loadbalancers.
func (c *loadbalancers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched loadbalancer.
func (c *loadbalancers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *api.LoadBalancer, err error) {
	result = &api.LoadBalancer{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(api.LoadBalancerPlural).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
