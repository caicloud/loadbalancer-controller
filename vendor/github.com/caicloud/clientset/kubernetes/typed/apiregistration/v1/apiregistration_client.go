/*
Copyright 2019 caicloud authors. All rights reserved.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"github.com/caicloud/clientset/kubernetes/scheme"
	v1 "github.com/caicloud/clientset/pkg/apis/apiregistration/v1"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type ApiregistrationV1Interface interface {
	RESTClient() rest.Interface
	APIServicesGetter
}

// ApiregistrationV1Client is used to interact with features provided by the apiregistration.k8s.io group.
type ApiregistrationV1Client struct {
	restClient rest.Interface
}

func (c *ApiregistrationV1Client) APIServices() APIServiceInterface {
	return newAPIServices(c)
}

// NewForConfig creates a new ApiregistrationV1Client for the given config.
func NewForConfig(c *rest.Config) (*ApiregistrationV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ApiregistrationV1Client{client}, nil
}

// NewForConfigOrDie creates a new ApiregistrationV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *ApiregistrationV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new ApiregistrationV1Client for the given RESTClient.
func New(c rest.Interface) *ApiregistrationV1Client {
	return &ApiregistrationV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ApiregistrationV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}