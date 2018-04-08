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
	"net"

	v1alpha2 "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
)

// ValidateLoadBalancer validate loadbalancer
func ValidateLoadBalancer(lb *v1alpha2.LoadBalancer) error {

	if lb.Spec.Providers.Ipvsdr != nil {
		ipvsdr := lb.Spec.Providers.Ipvsdr
		if net.ParseIP(ipvsdr.VIP) == nil {
			return fmt.Errorf("ipvsdr: vip is invalid")
		}
		switch ipvsdr.Scheduler {
		case v1alpha2.IpvsSchedulerRR:
		case v1alpha2.IpvsSchedulerWRR:
		case v1alpha2.IpvsSchedulerLC:
		case v1alpha2.IpvsSchedulerWLC:
		case v1alpha2.IpvsSchedulerLBLC:
		case v1alpha2.IpvsSchedulerDH:
		case v1alpha2.IpvsSchedulerSH:
			break
		default:
			return fmt.Errorf("ipvsdr: scheduler %v is invalid", ipvsdr.Scheduler)
		}
	}

	return nil
}
