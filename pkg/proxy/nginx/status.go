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

package nginx

import (
	"fmt"
	"sort"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"

	log "k8s.io/klog"
)

func (f *nginx) syncStatus(lb *lbapi.LoadBalancer) error {
	replicas, _ := lbutil.CalculateReplicas(lb)
	// caculate proxy status
	proxyStatus := lbapi.ProxyStatus{
		PodStatuses: lbapi.PodStatuses{
			Replicas:      replicas,
			ReadyReplicas: 0,
			TotalReplicas: 0,
			Statuses:      make([]lbapi.PodStatus, 0),
		},
		IngressClass: fmt.Sprintf(lbapi.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		ConfigMap:    fmt.Sprintf(configMapName, lb.Name),
		TCPConfigMap: fmt.Sprintf(tcpConfigMapName, lb.Name),
		UDPConfigMap: fmt.Sprintf(udpConfigMapName, lb.Name),
	}

	podList, err := f.podLister.List(f.selector(lb).AsSelector())
	if err != nil {
		log.Errorf("get pod list error: %v", err)
		return err
	}

	for _, pod := range podList {
		lbutil.EvictPod(f.client, lb, pod)

		status := lbutil.ComputePodStatus(pod)
		proxyStatus.TotalReplicas++
		if status.Ready {
			proxyStatus.ReadyReplicas++
		}
		proxyStatus.Statuses = append(proxyStatus.Statuses, status)
	}

	sort.Sort(lbutil.SortPodStatusByName(proxyStatus.Statuses))

	// check whether the statuses are equal
	if !lbutil.ProxyStatusEqual(lb.Status.ProxyStatus, proxyStatus) {
		// js, _ := json.Marshal(proxyStatus)
		// replacePatch := fmt.Sprintf(`{"status":{"proxyStatus": %s }}`, string(js))
		// _, err := f.tprclient.NetworkingV1alpha1().LoadBalancers(lb.Namespace).Patch(lb.Name, types.MergePatchType, []byte(replacePatch))
		_, err := lbutil.UpdateLBWithRetries(
			f.client.Custom().LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
			f.lbLister,
			lb.Namespace,
			lb.Name,
			func(lb *lbapi.LoadBalancer) error {
				lb.Status.ProxyStatus = proxyStatus
				return nil
			},
		)

		if err != nil {
			log.Errorf("Update loadbalancer status error: %v", err)
			return err
		}
	}
	return nil
}
