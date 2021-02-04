package kubernetes

import (
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/caicloud/clientset/customclient"
)

// Interface contains the clientsets for all API groups.
type Interface interface {
	Native() kubernetes.Interface
	Custom() customclient.Interface

	// third-part clients

	Apiextensions() apiextensionsclientset.Interface
}

type clientset struct {
	kubeClient       *kubernetes.Clientset
	customClient     *customclient.Clientset
	extensionsClient *apiextensionsclientset.Clientset
}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (Interface, error) {
	kubeClient, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	customClient, err := customclient.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	return &clientset{
		kubeClient:       kubeClient,
		customClient:     customClient,
		extensionsClient: apiextensionsClient,
	}, nil
}

// Native returns the standard kubernetes clientset.
func (c *clientset) Native() kubernetes.Interface {
	return c.kubeClient
}

// Custom returns the clientset for our own API group.
func (c *clientset) Custom() customclient.Interface {
	return c.customClient
}

// Apiextensions returns the apiextensions clientset.
func (c *clientset) Apiextensions() apiextensionsclientset.Interface {
	return c.extensionsClient
}
