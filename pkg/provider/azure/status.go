package azure

import (
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"

	"k8s.io/api/core/v1"
	log "k8s.io/klog"
)

func (f *azure) deleteStatus(lb *lbapi.LoadBalancer) error {
	if lb.Status.ProvidersStatuses.Azure == nil {
		return nil
	}

	log.Infof("delete azure status for loadbalancer %v/%v", lb.Namespace, lb.Name)
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
		log.Errorf("Update loadbalancer status error: %v", err)
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
		log.Errorf("get pod list error: %v", err)
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
		log.Infof("update azure status %s message %s", status.Phase, status.Message)
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
			log.Errorf("Update loadbalancer status error: %v", err)
			return err
		}
	}
	return nil
}
