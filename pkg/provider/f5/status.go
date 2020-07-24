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

package f5

import (
	"fmt"
	"sort"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	log "k8s.io/klog"
)

func (f *f5) deleteStatus(lb *lbapi.LoadBalancer) error {
	if lb.Status.ProvidersStatuses.F5 == nil {
		return nil
	}

	log.Infof("delete F5 status for loadbalancer %v/%v", lb.Namespace, lb.Name)
	_, err := lbutil.UpdateLBWithRetries(
		f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
		f.lbLister,
		lb.Namespace,
		lb.Name,
		func(lb *lbapi.LoadBalancer) error {
			lb.Status.ProvidersStatuses.F5 = nil
			return nil
		},
	)

	if err != nil {
		log.Errorf("Update loadbalancer status error: %v", err)
		return err
	}
	return nil
}

func (f *f5) syncStatus(lb *lbapi.LoadBalancer) error {
	if lb.Spec.Providers.F5 == nil {
		return f.deleteStatus(lb)
	}

	var pStatus string
	var pMessage string
	if lb.Status.ProvidersStatuses.F5 != nil {
		pStatus = lb.Status.ProvidersStatuses.F5.Status
		pMessage = lb.Status.ProvidersStatuses.F5.Message
	}

	providerStatus := lbapi.F5ProviderStatus{
		PodStatuses: lbapi.PodStatuses{
			Replicas:      1,
			ReadyReplicas: 0,
			TotalReplicas: 0,
			Statuses:      make([]lbapi.PodStatus, 0),
		},
		Status:  pStatus,
		Message: pMessage,
	}

	podList, err := f.podLister.List(f.selector(lb).AsSelector())
	if err != nil {
		log.Errorf("get pod list error: %v", err)
		return err
	}

	if len(podList) == 0 {
		return fmt.Errorf("loadbalancer %v provider pod not found", lb.Name)
	}

	for _, pod := range podList {
		status := lbutil.ComputePodStatus(pod)
		providerStatus.TotalReplicas++
		if status.Ready {
			providerStatus.ReadyReplicas++
		}
		providerStatus.Statuses = append(providerStatus.Statuses, status)
	}

	sort.Sort(lbutil.SortPodStatusByName(providerStatus.Statuses))

	// check whether the statuses are equal
	if lb.Status.ProvidersStatuses.F5 == nil || !lbutil.F5ProviderStatusEqual(*lb.Status.ProvidersStatuses.F5, providerStatus) {
		_, err := lbutil.UpdateLBWithRetries(
			f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
			f.lbLister,
			lb.Namespace,
			lb.Name,
			func(lb *lbapi.LoadBalancer) error {
				lb.Status.ProvidersStatuses.F5 = &providerStatus
				return nil
			},
		)

		if err != nil {
			log.Errorf("Update loadbalancer status error, %v", err)
			return err
		}

	}

	return nil
}
