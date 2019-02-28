package azure

import (
	log "github.com/zoumo/logdog"

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
