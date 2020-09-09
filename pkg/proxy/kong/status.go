package kong

import (
	"fmt"
	"sort"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	"k8s.io/klog"
)

func (k *kong) syncStatus(lb *lbapi.LoadBalancer) error {
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
	}

	podList, err := k.podLister.List(k.selector(lb).AsSelector())
	if err != nil {
		klog.Errorf("get pod list error: %v", err)
		return err
	}

	for _, pod := range podList {
		lbutil.EvictPod(k.client, lb, pod)

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
		if _, err := lbutil.UpdateLBWithRetries(
			k.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
			k.lbLister,
			lb.Namespace,
			lb.Name,
			func(lb *lbapi.LoadBalancer) error {
				lb.Status.ProxyStatus = proxyStatus
				return nil
			},
		); err != nil {
			klog.Errorf("Update loadbalancer status error: %v", err)
			return err
		}
	}
	return nil
}
