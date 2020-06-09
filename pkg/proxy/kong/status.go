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

package kong

import (
	"fmt"
	"sort"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"

	"k8s.io/apimachinery/pkg/labels"
	log "k8s.io/klog"
)

func (f *kong) syncStatus(lb *lbapi.LoadBalancer) error {
    // get ingress class from loadbalancer
    annotations := lb.GetAnnotations()
    ingressClass := annotations[ingressClassKey]
    if ingressClass == "" {
        ingressClass = defaultIngressClass
        annotations[ingressClass] = defaultIngressClass
        lb.Annotations = annotations
	}
	// get release name from loadbalancer annotation
	releaseName := annotations[releaseKey]
	if releaseName == "" {
		return fmt.Errorf("No release name annotation on loadbalancer %v", lb.Name)
	}
	
	replicas, _ := lbutil.CalculateReplicas(lb)
	// caculate proxy status
	proxyStatus := lbapi.ProxyStatus{
		PodStatuses: lbapi.PodStatuses{
			Replicas:      replicas,
			ReadyReplicas: 0,
			TotalReplicas: 0,
			Statuses:      make([]lbapi.PodStatus, 0),
		},
		Deployment: releaseName,
		IngressClass: ingressClass,
	}

	podList, err := f.podLister.List(labels.Set{releaseKey: releaseName}.AsSelector())
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
			f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
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