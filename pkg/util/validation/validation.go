/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package validation

import (
	"fmt"

	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
)

// ValidateLoadBalancer validate loadbalancer
func ValidateLoadBalancer(lb *netv1alpha1.LoadBalancer) error {
	lbType := lb.Spec.Type

	// if lb.Spec.Nodes.Replicas == nil && len(lb.Spec.Nodes.Names) == 0 {
	// 	return fmt.Errorf("both replicas and nodes are not fill in")
	// }

	switch lbType {
	case netv1alpha1.LoadBalancerTypeInternal:
		// internal lb must set service provider
		if lb.Spec.Providers.Service == nil {
			return fmt.Errorf("%s type loadbalancer must set servise provider spec", lbType)
		}
	case netv1alpha1.LoadBalancerTypeExternal:
		// external lb must set Nodes
		// if len(lb.Spec.Nodes.Names) == 0 {
		// 	return fmt.Errorf("%s type loadbalancer must set node spec", lbType)
		// }
	default:
		return fmt.Errorf("Unknown loadbalancer type %v", lbType)
	}

	return nil
}
