package internalinterfaces

import (
	"time"

	"github.com/caicloud/loadbalancer-controller/pkg/tprclient"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/tools/cache"
)

// SharedInformerFactory a small interface to allow for adding an informer without an import cycle
type SharedInformerFactory interface {
	Start(stopCh <-chan struct{})
	InformerFor(obj runtime.Object, newFunc internalinterfaces.NewInformerFunc) cache.SharedIndexInformer
	TPRInformerFor(obj runtime.Object, newFunc NewTPRInformerFunc) cache.SharedIndexInformer
}

// NewTPRInformerFunc ...
type NewTPRInformerFunc func(tprclient.Interface, time.Duration) cache.SharedIndexInformer
