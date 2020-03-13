package informers

import (
	"fmt"
	"time"

	"k8s.io/client-go/informers"

	"github.com/caicloud/clientset/custominformers"
	"github.com/caicloud/clientset/kubernetes"
)

// SharedInformerFactory provides shared informers for resources in all known API group versions.
type SharedInformerFactory interface {
	Native() informers.SharedInformerFactory
	Custom() custominformers.SharedInformerFactory

	Start(stopCh <-chan struct{})
	WaitForCacheSync(stopCh <-chan struct{}) error
}

type sharedInformerFactory struct {
	nativeInformer informers.SharedInformerFactory
	customInformer custominformers.SharedInformerFactory
}

// NewSharedInformerFactory constructs a new instance of sharedInformerFactory for all namespaces.
func NewSharedInformerFactory(client kubernetes.Interface, defaultResync time.Duration) SharedInformerFactory {
	return &sharedInformerFactory{
		nativeInformer: informers.NewSharedInformerFactory(client.Native(), defaultResync),
		customInformer: custominformers.NewSharedInformerFactory(client.Custom(), defaultResync),
	}
}

// Native returns the standard kubernetes sharedInformerFactory.
func (f *sharedInformerFactory) Native() informers.SharedInformerFactory {
	return f.nativeInformer
}

// Custom returns the sharedInformerFactory for our own API group.
func (f *sharedInformerFactory) Custom() custominformers.SharedInformerFactory {
	return f.customInformer
}

// Start initializes all requested informers.
func (f *sharedInformerFactory) Start(stopCh <-chan struct{}) {
	f.nativeInformer.Start(stopCh)
	f.customInformer.Start(stopCh)
}

// WaitForCacheSync waits for all started informers' cache were synced.
func (f *sharedInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) error {
	coreResources := f.nativeInformer.WaitForCacheSync(stopCh)
	for resType, syncd := range coreResources {
		if !syncd {
			return fmt.Errorf("native resource %s WaitForCacheSync failed", resType.Name())
		}
	}
	customResources := f.customInformer.WaitForCacheSync(stopCh)
	for resType, syncd := range customResources {
		if !syncd {
			return fmt.Errorf("custom resource %s WaitForCacheSync failed", resType.Name())
		}
	}
	return nil
}
