package azure

import (
	log "github.com/zoumo/logdog"
	"k8s.io/api/core/v1"

	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
)

func (f *azure) deleteStatus(lb *lbapi.LoadBalancer) error {
	if lb.Status.ProvidersStatuses.Azure == nil {
		return nil
	}

	log.Notice("delete azure status", log.Fields{"lb.name": lb.Name, "lb.ns": lb.Namespace})
	_, err := lbutil.UpdateLBWithRetries(
		f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
		f.lbLister,
		lb.Namespace,
		lb.Name,
		func(lb *lbapi.LoadBalancer) error {
			lb.Status.ProvidersStatuses.Azure = nil
			return nil
		},
	)

	if err != nil {
		log.Error("Update loadbalancer status error", log.Fields{"err": err})
		return err
	}
	return nil
}

func (f *azure) syncStatus(lb *lbapi.LoadBalancer) error {
	if lb.Spec.Providers.Azure == nil {
		return f.deleteStatus(lb)
	}

	podList, err := f.podLister.List(f.selector(lb).AsSelector())
	if err != nil {
		log.Error("get pod list error", log.Fields{"lb.ns": lb.Namespace, "lb.name": lb.Name, "err": err})
		return err
	}

	var status *lbapi.PodStatus
	for _, pod := range podList {
		s := lbutil.ComputePodStatus(pod)
		if s.Phase != string(v1.PodRunning) && s.Phase != string(v1.PodSucceeded) {
			status = &s
		}
	}

	if status != nil && (lb.Status.ProvidersStatuses.Azure.Phase != lbapi.AzureErrorPhase ||
		lb.Status.ProvidersStatuses.Azure.Message != status.Message) {
		log.Notice("update azure status %s message %s",
			log.Fields{"lb.name": lb.Name, "lb.ns": lb.Namespace, "lb.status": status.Phase, "lb.message": status.Message})
		_, err := lbutil.UpdateLBWithRetries(
			f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
			f.lbLister,
			lb.Namespace,
			lb.Name,
			func(lb *lbapi.LoadBalancer) error {
				lb.Status.ProvidersStatuses.Azure.Phase = lbapi.AzureErrorPhase
				lb.Status.ProvidersStatuses.Azure.Message = status.Message
				return nil
			},
		)

		if err != nil {
			log.Error("Update loadbalancer status error", log.Fields{"err": err})
			return err
		}
	}
	return nil
}
