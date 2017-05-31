package v1beta1

import "github.com/caicloud/loadbalancer-controller/pkg/informers/internalinterfaces"

// Interface provides access to all the informers in this group version.
type Interface interface {
	// PodDisruptionBudgets returns a PodDisruptionBudgetInformer.
	LoadBalancer() LoadBalancerInformer
}

type version struct {
	internalinterfaces.SharedInformerFactory
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory) Interface {
	return &version{f}
}

// LoadBalancer returns a LoadBalancerInformer.
func (v *version) LoadBalancer() LoadBalancerInformer {
	return &loadBalancerInformer{factory: v.SharedInformerFactory}
}
