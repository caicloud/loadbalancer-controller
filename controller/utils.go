package controller

import (
	"fmt"

	lbapi "github.com/caicloud/loadbalancer-controller/api"
	"github.com/caicloud/loadbalancer-controller/pkg/util/validation"
	log "github.com/zoumo/logdog"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

// VerifiedNodes ...
type VerifiedNodes struct {
	Nodes          []*apiv1.Node
	TaintsToAdd    []apiv1.Taint
	TaintsToDelete []apiv1.Taint
	Labels         map[string]string
}

func (lbc *LoadBalancerController) getVerifiedNodes(lb *lbapi.LoadBalancer) (*VerifiedNodes, error) {

	err := validation.ValidateLoadBalancer(lb)
	if err != nil {
		return nil, err
	}

	ran := &VerifiedNodes{
		TaintsToAdd:    []apiv1.Taint{},
		TaintsToDelete: []apiv1.Taint{},
		Nodes:          []*apiv1.Node{},
		Labels:         map[string]string{},
	}

	ran.Labels = map[string]string{
		fmt.Sprintf(lbapi.UniqueLabelKeyFormat, lb.Namespace, lb.Name): "true",
	}

	if lb.Nodes == nil {
		// if Nodes is not fill in, we should delete taint by key
		// no matter what effect it is
		ran.TaintsToDelete = append(ran.TaintsToDelete, apiv1.Taint{
			// loadbalancer.alpha.caicloud.io/dedicated=namespace-name:effect
			Key: lbapi.TaintKey,
		})
	} else {
		if lb.Nodes.Dedicated != nil {
			// generate taints to add
			ran.TaintsToAdd = append(ran.TaintsToAdd, apiv1.Taint{
				// loadbalancer.alpha.caicloud.io/dedicated=namespace-name:effect
				Key:    lbapi.TaintKey,
				Value:  fmt.Sprintf(lbapi.TaintValueFormat, lb.Namespace, lb.Name),
				Effect: *lb.Nodes.Dedicated,
			})
		} else {
			// if dedicated is not fill in, we should delete taint by key
			// no matter what effect it is
			ran.TaintsToDelete = append(ran.TaintsToDelete, apiv1.Taint{
				// loadbalancer.alpha.caicloud.io/dedicated=namespace-name:effect
				Key: lbapi.TaintKey,
			})
		}

		// get valid nodes
		for _, name := range lb.Nodes.Names {
			// get node
			node, err := lbc.nodeLister.Get(name)
			if err != nil {
				log.Error("Error when get node info, ignore it", log.Fields{"name": name})
				continue
			}

			// BUG
			// validate taint
			// err = taints.ValidateNoTaintOverwrites(node, taintsToAdd)
			// if err != nil {
			// 	// node already has a taint with key, can not use it
			// 	log.Warn("validate node taints error, igore it", log.Fields{"name": name, "err": err})
			// 	continue
			// }

			ran.Nodes = append(ran.Nodes, node)
		}

	}

	return ran, nil
}
