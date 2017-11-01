package api

import (
	"fmt"

	"github.com/caicloud/clientset/kubernetes/scheme"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	log "github.com/zoumo/logdog"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// LoadBalancerDeepCopy clones the given LoadBalancer and returns a new one.
func LoadBalancerDeepCopy(lb *lbapi.LoadBalancer) (*lbapi.LoadBalancer, error) {
	lbi, err := scheme.Scheme.DeepCopy(lb)
	if err != nil {
		log.Error("Unable to deepcopy loadbalancer", log.Fields{"lb.name": lb.Name, "err": err})
		return nil, err
	}

	nlb, ok := lbi.(*lbapi.LoadBalancer)
	if !ok {
		nerr := fmt.Errorf("expected loadbalancer, got %#v", lbi)
		log.Error(nerr)
		return nil, err
	}
	return nlb, nil
}

// DeploymentDeepCopy returns a deepcopy for given deployment
func DeploymentDeepCopy(deployment *extensions.Deployment) (*extensions.Deployment, error) {
	objCopy, err := scheme.Scheme.DeepCopy(deployment)
	if err != nil {
		return nil, err
	}
	copied, ok := objCopy.(*extensions.Deployment)
	if !ok {
		return nil, fmt.Errorf("expected Deployment, got %#v", objCopy)
	}
	return copied, nil
}
