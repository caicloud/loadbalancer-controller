/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1alpha2

import (
	"fmt"
	"strings"

	loadbalanceapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"k8s.io/apimachinery/pkg/api/meta"
)

// LoadBalancerListerExpansion allows custom methods to be added to
// LoadBalancerLister.
type LoadBalancerListerExpansion interface {
	GetLoadBalancerForControllee(obj interface{}) (*loadbalanceapi.LoadBalancer, error)
}

// LoadBalancerNamespaceListerExpansion allows custom methods to be added to
// LoadBalancerNamespaceLister.
type LoadBalancerNamespaceListerExpansion interface{}

// GetLoadBalancerForControllee
// 1. try to get loadbalancer from OwnerReferences
// 2. try to find created-by key in controllee's labels, parse the value to find namesapce and name
func (s *loadBalancerLister) GetLoadBalancerForControllee(obj interface{}) (*loadbalanceapi.LoadBalancer, error) {
	version := "v1alpha2"
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("Error access object metadata: %v", err)
	}

	var namespace, name string

	// 1. get namespace, name from owner reference
	owners := accessor.GetOwnerReferences()
	for _, owner := range owners {
		if owner.APIVersion == loadbalanceapi.GroupName+"/"+version && owner.Kind == "LoadBalancer" {
			namespace = accessor.GetNamespace()
			name = owner.Name
			break
		}
	}

	if namespace == "" && name == "" {
		// 2. try to get from labels
		labels := accessor.GetLabels()
		value, ok := labels[loadbalanceapi.LabelKeyCreatedBy]
		if !ok {
			return nil, fmt.Errorf("the object %v doesn't have label to indicate it is created by loadbalancer", accessor.GetName())
		}
		temp := strings.Split(value, ".")
		if len(temp) != 2 {
			return nil, fmt.Errorf("Error label value format")
		}
		namespace, name = temp[0], temp[1]
	}

	return s.LoadBalancers(namespace).Get(name)

}
