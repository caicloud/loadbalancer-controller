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

package v1alpha1

import (
	"fmt"

	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
)

// LoadBalancerListerExpansion allows custom methods to be added to
// LoadBalancerLister.
type LoadBalancerListerExpansion interface {
	GetLoadBalancersForControllee(obj interface{}) ([]*netv1alpha1.LoadBalancer, error)
}

// LoadBalancerNamespaceListerExpansion allows custom methods to be added to
// LoadBalancerNamespaeLister.
type LoadBalancerNamespaceListerExpansion interface{}

// GetLoadBalancersForControllee
func (s *loadBalancerList) GetLoadBalancersForControllee(obj interface{}) ([]*netv1alpha1.LoadBalancer, error) {
	meta, err := apimeta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("object has no meta: %v", err)
	}

	tpy, err := apimeta.TypeAccessor(obj)
	if err != nil {
		return nil, fmt.Errorf("object has no type: %v", err)
	}

	if len(meta.GetLabels()) == 0 {
		return nil, fmt.Errorf("no loadbalancers found for daemonset %v because it has no labels", meta.GetName())
	}

	lbList, err := s.LoadBalancers(meta.GetNamespace()).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var lbs []*netv1alpha1.LoadBalancer
	for _, lb := range lbList {
		// use loadbalancer namespace and name construct unique key
		selector := labels.Set{
			netv1alpha1.LabelKeyCreatedBy: fmt.Sprintf(netv1alpha1.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		}.AsSelector()

		if !selector.Matches(labels.Set(meta.GetLabels())) {
			// d is not creatby this lb
			continue
		}
		lbs = append(lbs, lb)
	}

	if len(lbs) == 0 {
		return nil, fmt.Errorf("could not find loadbalancer for %v %s in namespace %s with labels: %v", tpy.GetKind(), meta.GetName(), meta.GetNamespace(), meta.GetLabels())
	}

	return lbs, nil
}
