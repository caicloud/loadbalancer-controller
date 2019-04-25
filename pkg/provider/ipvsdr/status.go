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

package ipvsdr

import (
	"sort"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	stringsutil "github.com/caicloud/loadbalancer-controller/pkg/util/strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog"
)

func (f *ipvsdr) syncStatus(lb *lbapi.LoadBalancer, activeDeploy *appsv1.Deployment) error {
	if lb.Spec.Providers.Ipvsdr == nil {
		return f.deleteStatus(lb)
	}
	// caculate proxy status
	providerStatus := lbapi.IpvsdrProviderStatus{
		PodStatuses: lbapi.PodStatuses{
			Replicas:      *activeDeploy.Spec.Replicas,
			ReadyReplicas: 0,
			TotalReplicas: 0,
			Statuses:      make([]lbapi.PodStatus, 0),
		},
		VIP:        lb.Spec.Providers.Ipvsdr.VIP,
		Deployment: activeDeploy.Name,
	}

	// the following loadbalancer need to get a valid vrid
	// 1. a new lb need to get a valid vrid
	// 2. a old lb didn't have a valid vrid
	ipvsdrstatus := lb.Status.ProvidersStatuses.Ipvsdr
	if ipvsdrstatus == nil || ipvsdrstatus.Vrid == nil || *ipvsdrstatus.Vrid == -1 {
		// keepalived use unicast now. so vrid will not be conflict
		vrid := f.getValidVRID()
		providerStatus.Vrid = &vrid
	} else {
		providerStatus.Vrid = ipvsdrstatus.Vrid
	}

	podList, err := f.podLister.List(f.selector(lb).AsSelector())
	if err != nil {
		log.Errorf("get pod list error: %v", err)
		return err
	}

	for _, pod := range podList {
		f.evictPod(lb, pod)

		status := lbutil.ComputePodStatus(pod)
		providerStatus.TotalReplicas++
		if status.Ready {
			providerStatus.ReadyReplicas++
		}
		providerStatus.Statuses = append(providerStatus.Statuses, status)
	}

	sort.Sort(lbutil.SortPodStatusByName(providerStatus.Statuses))

	// check whether the statuses are equal
	if ipvsdrstatus == nil || !lbutil.IpvsdrProviderStatusEqual(*ipvsdrstatus, providerStatus) {
		// js, _ := json.Marshal(providerStatus)
		// replacePatch := fmt.Sprintf(`{"status":{"providersStatuses":{"ipvsdr": %s}}}`, string(js))
		_, err := lbutil.UpdateLBWithRetries(
			f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
			f.lbLister,
			lb.Namespace,
			lb.Name,
			func(lb *lbapi.LoadBalancer) error {
				lb.Status.ProvidersStatuses.Ipvsdr = &providerStatus
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

func (f *ipvsdr) deleteStatus(lb *lbapi.LoadBalancer) error {
	if lb.Status.ProvidersStatuses.Ipvsdr == nil {
		return nil
	}

	log.Infof("delete ipvsdr status for %v/%v", lb.Namespace, lb.Name)
	_, err := lbutil.UpdateLBWithRetries(
		f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
		f.lbLister,
		lb.Namespace,
		lb.Name,
		func(lb *lbapi.LoadBalancer) error {
			lb.Status.ProvidersStatuses.Ipvsdr = nil
			return nil
		},
	)

	if err != nil {
		log.Errorf("Update loadbalancer status error: %v", err)
		return err
	}
	return nil
}

func (f *ipvsdr) evictPod(lb *lbapi.LoadBalancer, pod *v1.Pod) {

	if len(lb.Spec.Nodes.Names) == 0 {
		return
	}

	// fix: avoid evict pending pod
	if pod.Spec.NodeName == "" {
		return
	}

	evict := func() {
		f.client.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
	}

	// FIXME: when RequiredDuringSchedulingRequiredDuringExecution finished
	// This is a special issue.
	// There is bug when the nodes.Names change。
	// According to nodeAffinity RequiredDuringSchedulingIgnoredDuringExecution,
	// the system may or may not try to eventually evict the pod from its node.
	// the pod may still running on the wrong node, so we evict it manually
	if !stringsutil.StringInSlice(pod.Spec.NodeName, lb.Spec.Nodes.Names) &&
		pod.DeletionTimestamp == nil {
		evict()
		return
	}

	// evict pod MatchNodeSelector Failed
	if lbutil.IsPodMatchNodeSelectorFailed(pod) && pod.DeletionTimestamp == nil {
		evict()
		return
	}

}
