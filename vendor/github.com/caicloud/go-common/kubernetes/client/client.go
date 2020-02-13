package client

import (
	"github.com/caicloud/clientset/kubernetes"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// DefaultQPS is the default QPS value we used in caicloud
	DefaultQPS = 50
	// DefaultBurst is the default Burst value we used in caicloud
	DefaultBurst = 100
)

// SetQPS sets the QPS and Burst.
func SetQPS(qps float32, burst int) func(c *rest.Config) {
	return func(c *rest.Config) {
		c.QPS = qps
		c.Burst = burst
	}
}

// NewFromConfig creates a new kubernetes client from the given config and options.
func NewFromConfig(c *rest.Config, options ...func(c *rest.Config)) (kubernetes.Interface, error) {
	for _, opt := range options {
		opt(c)
	}

	if c.QPS == 0 {
		c.QPS = DefaultQPS
	}
	if c.Burst == 0 {
		c.Burst = DefaultBurst
	}
	return kubernetes.NewForConfig(c)
}

// NewFromFlags generates a kubernetes client by master URL and kube config.
// Write your own custom function that modifies the rest.Config if default config can not satisfy your needs.
func NewFromFlags(masterURL, kubeconfigPath string, options ...func(c *rest.Config)) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		return nil, err
	}
	return NewFromConfig(config, options...)
}
