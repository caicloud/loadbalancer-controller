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

package lb

import (
	lbclient "github.com/caicloud/clientset/kubernetes/typed/loadbalance/v1alpha2"
	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

type updateLBFunc func(lb *lbapi.LoadBalancer) error

// UpdateLBWithRetries update loadbalancer with max retries
func UpdateLBWithRetries(lbClient lbclient.LoadBalancerInterface, lblister lblisters.LoadBalancerLister, namespace, name string, applyUpdate updateLBFunc) (*lbapi.LoadBalancer, error) {
	var lb *lbapi.LoadBalancer

	retryErr := wait.ExponentialBackoff(DefaultRetry, func() (bool, error) {
		var err error
		// get from lister
		lb, err = lblister.LoadBalancers(namespace).Get(name)
		if err != nil {
			return false, err
		}
		// deep copy
		nlb := lb.DeepCopy()
		lb = nlb

		// apply change
		if applyErr := applyUpdate(lb); applyErr != nil {
			return false, applyErr
		}

		// update to apiserver
		lb, err = lbClient.Update(lb)
		if err == nil {
			return true, nil
		}
		if errors.IsConflict(err) {
			// retry
			return false, nil
		}
		return false, err
	})

	return lb, retryErr

}
