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

package tprclient

import (
	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/tprclient/networking/v1alpha1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

var _ Interface = &Clientset{}

// Interface ...
type Interface interface {
	NetworkingV1alpha1() netv1alpha1.NetworkingV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*netv1alpha1.NetworkingV1alpha1Client
}

// NetworkingV1alpha1 retrieves the NetworkingV1alpha1Client
func (c *Clientset) NetworkingV1alpha1() netv1alpha1.NetworkingV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.NetworkingV1alpha1Client
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	var cs Clientset
	var err error
	cs.NetworkingV1alpha1Client, err = netv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.NetworkingV1alpha1Client = netv1alpha1.NewForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.NetworkingV1alpha1Client = netv1alpha1.New(c)
	return &cs
}
