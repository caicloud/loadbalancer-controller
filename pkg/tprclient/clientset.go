package tprclient

import (
	caicloudv1beta1 "github.com/caicloud/loadbalancer-controller/pkg/tprclient/caicloud/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

var _ Interface = &Clientset{}

// Interface ...
type Interface interface {
	CaicloudV1beta1() caicloudv1beta1.CaiCloudV1beta1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*caicloudv1beta1.CaicloudV1beta1Client
}

// CaicloudV1beta1 retrieves the CaicloudV1beta1Client
func (c *Clientset) CaicloudV1beta1() caicloudv1beta1.CaiCloudV1beta1Interface {
	if c == nil {
		return nil
	}
	return c.CaicloudV1beta1Client
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}

	var cs Clientset
	var err error
	cs.CaicloudV1beta1Client, err = caicloudv1beta1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.CaicloudV1beta1Client = caicloudv1beta1.NewForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.CaicloudV1beta1Client = caicloudv1beta1.New(c)
	return &cs
}
