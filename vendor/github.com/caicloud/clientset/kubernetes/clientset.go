/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package kubernetes

import (
	apiextensionsv1beta1 "github.com/caicloud/clientset/kubernetes/typed/apiextensions/v1beta1"
	configv1alpha1 "github.com/caicloud/clientset/kubernetes/typed/config/v1alpha1"
	loadbalancev1alpha2 "github.com/caicloud/clientset/kubernetes/typed/loadbalance/v1alpha2"
	releasev1alpha1 "github.com/caicloud/clientset/kubernetes/typed/release/v1alpha1"
	resourcev1alpha1 "github.com/caicloud/clientset/kubernetes/typed/resource/v1alpha1"
	glog "github.com/golang/glog"
	kubernetes "k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	kubernetes.Interface
	ApiextensionsV1beta1() apiextensionsv1beta1.ApiextensionsV1beta1Interface
	// Deprecated: please explicitly pick a version if possible.
	Apiextensions() apiextensionsv1beta1.ApiextensionsV1beta1Interface
	ConfigV1alpha1() configv1alpha1.ConfigV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Config() configv1alpha1.ConfigV1alpha1Interface
	LoadbalanceV1alpha2() loadbalancev1alpha2.LoadbalanceV1alpha2Interface
	// Deprecated: please explicitly pick a version if possible.
	Loadbalance() loadbalancev1alpha2.LoadbalanceV1alpha2Interface
	ReleaseV1alpha1() releasev1alpha1.ReleaseV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Release() releasev1alpha1.ReleaseV1alpha1Interface
	ResourceV1alpha1() resourcev1alpha1.ResourceV1alpha1Interface
	// Deprecated: please explicitly pick a version if possible.
	Resource() resourcev1alpha1.ResourceV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*kubernetes.Clientset
	*apiextensionsv1beta1.ApiextensionsV1beta1Client
	*configv1alpha1.ConfigV1alpha1Client
	*loadbalancev1alpha2.LoadbalanceV1alpha2Client
	*releasev1alpha1.ReleaseV1alpha1Client
	*resourcev1alpha1.ResourceV1alpha1Client
}

// ApiextensionsV1beta1 retrieves the ApiextensionsV1beta1Client
func (c *Clientset) ApiextensionsV1beta1() apiextensionsv1beta1.ApiextensionsV1beta1Interface {
	if c == nil {
		return nil
	}
	return c.ApiextensionsV1beta1Client
}

// Deprecated: Apiextensions retrieves the default version of ApiextensionsClient.
// Please explicitly pick a version.
func (c *Clientset) Apiextensions() apiextensionsv1beta1.ApiextensionsV1beta1Interface {
	if c == nil {
		return nil
	}
	return c.ApiextensionsV1beta1Client
}

// ConfigV1alpha1 retrieves the ConfigV1alpha1Client
func (c *Clientset) ConfigV1alpha1() configv1alpha1.ConfigV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.ConfigV1alpha1Client
}

// Deprecated: Config retrieves the default version of ConfigClient.
// Please explicitly pick a version.
func (c *Clientset) Config() configv1alpha1.ConfigV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.ConfigV1alpha1Client
}

// LoadbalanceV1alpha2 retrieves the LoadbalanceV1alpha2Client
func (c *Clientset) LoadbalanceV1alpha2() loadbalancev1alpha2.LoadbalanceV1alpha2Interface {
	if c == nil {
		return nil
	}
	return c.LoadbalanceV1alpha2Client
}

// Deprecated: Loadbalance retrieves the default version of LoadbalanceClient.
// Please explicitly pick a version.
func (c *Clientset) Loadbalance() loadbalancev1alpha2.LoadbalanceV1alpha2Interface {
	if c == nil {
		return nil
	}
	return c.LoadbalanceV1alpha2Client
}

// ReleaseV1alpha1 retrieves the ReleaseV1alpha1Client
func (c *Clientset) ReleaseV1alpha1() releasev1alpha1.ReleaseV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.ReleaseV1alpha1Client
}

// Deprecated: Release retrieves the default version of ReleaseClient.
// Please explicitly pick a version.
func (c *Clientset) Release() releasev1alpha1.ReleaseV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.ReleaseV1alpha1Client
}

// ResourceV1alpha1 retrieves the ResourceV1alpha1Client
func (c *Clientset) ResourceV1alpha1() resourcev1alpha1.ResourceV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.ResourceV1alpha1Client
}

// Deprecated: Resource retrieves the default version of ResourceClient.
// Please explicitly pick a version.
func (c *Clientset) Resource() resourcev1alpha1.ResourceV1alpha1Interface {
	if c == nil {
		return nil
	}
	return c.ResourceV1alpha1Client
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.ApiextensionsV1beta1Client, err = apiextensionsv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.ConfigV1alpha1Client, err = configv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.LoadbalanceV1alpha2Client, err = loadbalancev1alpha2.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.ReleaseV1alpha1Client, err = releasev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.ResourceV1alpha1Client, err = resourcev1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.Clientset, err = kubernetes.NewForConfig(&configShallowCopy)
	if err != nil {
		glog.Errorf("failed to create the client-go Clientset: %v", err)
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.ApiextensionsV1beta1Client = apiextensionsv1beta1.NewForConfigOrDie(c)
	cs.ConfigV1alpha1Client = configv1alpha1.NewForConfigOrDie(c)
	cs.LoadbalanceV1alpha2Client = loadbalancev1alpha2.NewForConfigOrDie(c)
	cs.ReleaseV1alpha1Client = releasev1alpha1.NewForConfigOrDie(c)
	cs.ResourceV1alpha1Client = resourcev1alpha1.NewForConfigOrDie(c)

	cs.Clientset = kubernetes.NewForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.ApiextensionsV1beta1Client = apiextensionsv1beta1.New(c)
	cs.ConfigV1alpha1Client = configv1alpha1.New(c)
	cs.LoadbalanceV1alpha2Client = loadbalancev1alpha2.New(c)
	cs.ReleaseV1alpha1Client = releasev1alpha1.New(c)
	cs.ResourceV1alpha1Client = resourcev1alpha1.New(c)

	cs.Clientset = kubernetes.New(c)
	return &cs
}
