package v1beta1

import (
	"github.com/caicloud/loadbalancer-controller/api"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// CaiCloudV1beta1Interface ...
type CaiCloudV1beta1Interface interface {
	RESTClient() rest.Interface
	LoadBalacnersGetter
}

var _ CaiCloudV1beta1Interface = &CaicloudV1beta1Client{}

// CaicloudV1beta1Client is is used to interact with features provided by the caicloud group.
type CaicloudV1beta1Client struct {
	restClient rest.Interface
}

// LoadBalancers returns LoadBalancerInterface
func (c *CaicloudV1beta1Client) LoadBalancers(namespace string) LoadBalancerInterface {
	return newLoadBalancers(c, namespace)
}

// NewForConfig creates a new CaicloudV1beta1Client for the given config.
func NewForConfig(c *rest.Config) (*CaicloudV1beta1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &CaicloudV1beta1Client{client}, nil
}

// NewForConfigOrDie creates a new AppsV1beta1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *CaicloudV1beta1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new CaicloudV1beta1Client for the given RESTClient.
func New(c rest.Interface) *CaicloudV1beta1Client {
	return &CaicloudV1beta1Client{c}
}

func setConfigDefaults(config *rest.Config) error {

	config.GroupVersion = &api.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *CaicloudV1beta1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
