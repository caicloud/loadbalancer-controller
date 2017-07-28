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
	"encoding/json"
	"fmt"

	log "github.com/zoumo/logdog"

	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	stringsutil "github.com/caicloud/loadbalancer-controller/pkg/util/strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

func (f *ipvsdr) syncStatus(lb *netv1alpha1.LoadBalancer, activeDeploy *extensions.Deployment) error {

	// caculate proxy status
	providerStatus := netv1alpha1.IpvsdrProviderStatus{
		PodStatuses: netv1alpha1.PodStatuses{
			Replicas:      *activeDeploy.Spec.Replicas,
			ReadyReplicas: 0,
			TotalReplicas: 0,
			Statuses:      make([]netv1alpha1.PodStatus, 0),
		},
		Vip:        lb.Spec.Providers.Ipvsdr.Vip,
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
		log.Error("get pod list error", log.Fields{"lb.ns": lb.Namespace, "lb.name": lb.Name, "err": err})
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

	// check whether the statuses are equal
	if ipvsdrstatus == nil || !lbutil.IpvsdrProviderStatusEqual(*ipvsdrstatus, providerStatus) {
		js, _ := json.Marshal(providerStatus)
		replacePatch := fmt.Sprintf(`{"status":{"providersStatuses":{"ipvsdr": %s}}}`, string(js))
		log.Notice("update ipvsdr status", log.Fields{"lb.name": lb.Name, "lb.ns": lb.Namespace})
		_, err = f.tprclient.NetworkingV1alpha1().LoadBalancers(lb.Namespace).Patch(lb.Name, types.MergePatchType, []byte(replacePatch))
		if err != nil {
			log.Error("Update loadbalancer status error", log.Fields{"err": err})
			return err
		}

	}
	return nil
}

func (f *ipvsdr) evictPod(lb *netv1alpha1.LoadBalancer, pod *v1.Pod) {

	if len(lb.Spec.Nodes.Names) == 0 {
		return
	}
	// FIXME: when RequiredDuringSchedulingRequiredDuringExecution finished
	// This is a special issue.
	// There is bug when the nodes.Names change。
	// According to nodeAffinity RequiredDuringSchedulingIgnoredDuringExecution,
	// the system may or may not try to eventually evict the pod from its node.
	// the pod may still running on the wrong node, so we evict it manually
	if !stringsutil.StringInSlice(pod.Spec.NodeName, lb.Spec.Nodes.Names) &&
		pod.DeletionTimestamp == nil {
		f.client.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
	}
}
