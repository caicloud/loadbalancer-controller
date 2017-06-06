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
	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
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

// LoadBalancerInterface ...
type LoadBalancerInterface interface {
	Create(*netv1alpha1.LoadBalancer) (*netv1alpha1.LoadBalancer, error)
	Update(*netv1alpha1.LoadBalancer) (*netv1alpha1.LoadBalancer, error)
	// UpdateStatus(*netv1alpha1.LoadBalancer) (*netv1alpha1.LoadBalancer, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*netv1alpha1.LoadBalancer, error)
	List(opts v1.ListOptions) (*netv1alpha1.LoadBalancerList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *netv1alpha1.LoadBalancer, err error)
	LoadBalancerExpansion
}

var _ LoadBalancerInterface = &loadbalancers{}

type loadbalancers struct {
	client rest.Interface
	ns     string
}

func newLoadBalancers(c *NetworkingV1alpha1Client, namespace string) *loadbalancers {
	return &loadbalancers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Create takes the representation of a loadbalancer and creates it.  Returns the server's representation of the loadbalancer, and an error, if there is any.
func (c *loadbalancers) Create(lb *netv1alpha1.LoadBalancer) (result *netv1alpha1.LoadBalancer, err error) {
	result = &netv1alpha1.LoadBalancer{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource(netv1alpha1.LoadBalancerPlural).
		Body(lb).
		Do().
		Into(result)
	return
}

// Update takes the representation of a loadbalancer and updates it. Returns the server's representation of the loadbalancer, and an error, if there is any.
func (c *loadbalancers) Update(lb *netv1alpha1.LoadBalancer) (result *netv1alpha1.LoadBalancer, err error) {
	result = &netv1alpha1.LoadBalancer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource(netv1alpha1.LoadBalancerPlural).
		Name(lb.Name).
		Body(lb).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclientstatus=false comment above the type to avoid generating UpdateStatus().

// third party resource do not support subresource now
//
// func (c *loadbalancers) UpdateStatus(lb *netv1alpha1.LoadBalancer) (result *netv1alpha1.LoadBalancer, err error) {
// 	result = &netv1alpha1.LoadBalancer{}
// 	err = c.client.Put().
// 		Namespace(c.ns).
// 		Resource(netv1alpha1.LoadBalancerPlural).
// 		Name(lb.Name).
// 		SubResource("status").
// 		Body(lb).
// 		Do().
// 		Into(result)
// 	return
// }

// Delete takes name of the loadbalancer and deletes it. Returns an error if one occurs.
func (c *loadbalancers) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(netv1alpha1.LoadBalancerPlural).
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *loadbalancers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource(netv1alpha1.LoadBalancerPlural).
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Get takes name of the loadbalancer, and returns the corresponding loadbalancer object, and an error if there is any.
func (c *loadbalancers) Get(name string, options v1.GetOptions) (result *netv1alpha1.LoadBalancer, err error) {
	result = &netv1alpha1.LoadBalancer{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(netv1alpha1.LoadBalancerPlural).
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of LoadBalancers that match those selectors.
func (c *loadbalancers) List(opts v1.ListOptions) (result *netv1alpha1.LoadBalancerList, err error) {
	result = &netv1alpha1.LoadBalancerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource(netv1alpha1.LoadBalancerPlural).
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
		Resource(netv1alpha1.LoadBalancerPlural).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Patch applies the patch and returns the patched loadbalancer.
func (c *loadbalancers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *netv1alpha1.LoadBalancer, err error) {
	result = &netv1alpha1.LoadBalancer{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource(netv1alpha1.LoadBalancerPlural).
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
