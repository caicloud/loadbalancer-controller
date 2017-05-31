package validation

import (
	"fmt"

	"github.com/caicloud/loadbalancer-controller/api"
)

// ValidateLoadBalancer validate loadbalancer
func ValidateLoadBalancer(lb *api.LoadBalancer) error {
	lbType := lb.LoadBalancerSpec.Type

	if lb.Replicas == nil && lb.Nodes == nil {
		return fmt.Errorf("both replicas and nodes are not fill in")
	}

	switch lbType {
	case api.LoadBalancerTypeInternal:
		// internal lb must set service provider
		if lb.Providers.Service == nil {
			return fmt.Errorf("%s type loadbalancer must set servise provider spec", lbType)
		}
	case api.LoadBalancerTypeExternal:
		// external lb must set Nodes
		if lb.Nodes == nil {
			return fmt.Errorf("%s type loadbalancer must set node spec", lbType)
		}
	default:
		return fmt.Errorf("Unknown loadbalancer type %v", lbType)
	}

	return nil
}
