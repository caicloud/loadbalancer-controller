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

package controller

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/caicloud/clientset/kubernetes"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/loadbalancer-controller/pkg/util/taints"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	corelisters "k8s.io/client-go/listers/core/v1"
	log "k8s.io/klog"
)

// VerifiedNodes ...
type VerifiedNodes struct {
	NodesInUse     []*apiv1.Node
	NodesToDelete  []*apiv1.Node
	TaintsToAdd    []apiv1.Taint
	TaintsToDelete []apiv1.Taint
	Labels         map[string]string
}

type nodeController struct {
	client     kubernetes.Interface
	nodeLister corelisters.NodeLister
}

func (nc *nodeController) syncNodes(lb *lbapi.LoadBalancer) error {
	oldNodes, err := nc.getNodesForLoadBalancer(lb)
	if err != nil {
		log.Error("list node error")
		return err
	}
	// varify desired nodes
	desiredNodes, err := nc.getVerifiedNodes(lb, oldNodes)
	if err != nil {
		log.Errorf("varify nodes error: %v", err)
		return err
	}
	return nc.doLabelAndTaints(desiredNodes)
}

func (nc *nodeController) getNodesForLoadBalancer(lb *lbapi.LoadBalancer) ([]*apiv1.Node, error) {
	ingressClass := getIngressClassFromLoadbalancer(lb)
	// list old nodes
	labelkey := fmt.Sprintf(lbapi.UniqueLabelKeyFormat, lb.Namespace, ingressClass)
	selector := labels.Set{labelkey: "true"}.AsSelector()
	return nc.nodeLister.List(selector)
}

func (nc *nodeController) getVerifiedNodes(lb *lbapi.LoadBalancer, oldNodes []*apiv1.Node) (*VerifiedNodes, error) {
	ran := &VerifiedNodes{
		TaintsToAdd:    []apiv1.Taint{},
		TaintsToDelete: []apiv1.Taint{},
		NodesInUse:     []*apiv1.Node{},
		NodesToDelete:  []*apiv1.Node{},
		Labels:         map[string]string{},
	}

	ingressClass := getIngressClassFromLoadbalancer(lb)
	ran.Labels = map[string]string{
		fmt.Sprintf(lbapi.UniqueLabelKeyFormat, lb.Namespace, ingressClass): "true",
	}

	if len(lb.Spec.Nodes.Names) == 0 {
		// if Nodes is not fill in, we should delete taint by key
		// no matter what effect it is
		ran.TaintsToDelete = append(ran.TaintsToDelete, apiv1.Taint{
			// loadbalancer.alpha.caicloud.io/dedicated=namespace-name:effect
			Key: lbapi.TaintKey,
		})

		// delete all old nodes
		ran.NodesToDelete = oldNodes

		return ran, nil
	}

	if lb.Spec.Nodes.Effect != nil {
		// generate taints to add
		ran.TaintsToAdd = append(ran.TaintsToAdd, apiv1.Taint{
			// loadbalancer.alpha.caicloud.io/dedicated=namespace-name:effect
			Key:    lbapi.TaintKey,
			Value:  fmt.Sprintf(lbapi.TaintValueFormat, lb.Namespace, lb.Name),
			Effect: *lb.Spec.Nodes.Effect,
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
	for _, name := range lb.Spec.Nodes.Names {
		// get node
		node, err := nc.nodeLister.Get(name)
		if err != nil {
			log.Errorf("Error when get node %v info, ignore it", name)
			continue
		}

		// BUG
		// validate taint
		// err = taints.ValidateNoTaintOverwrites(node, taintsToAdd)
		// if err != nil {
		// 	// node already has a taint with key, can not use it
		// 	log.Warn("validate node taints error, ignore it", log.Fields{"name": name, "err": err})
		// 	continue
		// }

		ran.NodesInUse = append(ran.NodesInUse, node)
	}

	ran.NodesToDelete = nc.nodesDiff(oldNodes, ran.NodesInUse)

	return ran, nil
}

func (nc *nodeController) nodesDiff(oldNodes, desiredNodes []*apiv1.Node) []*apiv1.Node {
	if len(desiredNodes) == 0 {
		return oldNodes
	}
	nodesToDelete := make([]*apiv1.Node, 0)

NEXT:
	for _, oldNode := range oldNodes {
		for _, desiredNode := range desiredNodes {
			if oldNode.Name == desiredNode.Name {
				continue NEXT
			}
		}
		nodesToDelete = append(nodesToDelete, oldNode)
	}
	return nodesToDelete
}

// doLabelAndTaints delete label and taints in nodesToDelete
// add label and taints in nodes
func (nc *nodeController) doLabelAndTaints(desiredNodes *VerifiedNodes) error {
	// delete labels and taints from old nodes
	for _, node := range desiredNodes.NodesToDelete {
		copyNode := node.DeepCopy()

		// change labels
		for key := range desiredNodes.Labels {
			delete(copyNode.Labels, key)
		}

		// change taints
		// maybe taints are not found, reorganize will return error but it doesn't matter
		// taints will not be changed
		_, newTaints, _ := taints.ReorganizeTaints(copyNode, false, nil, []apiv1.Taint{
			{Key: lbapi.TaintKey},
		})
		copyNode.Spec.Taints = newTaints

		labelChanged := !reflect.DeepEqual(node.Labels, copyNode.Labels)
		taintChanged := !reflect.DeepEqual(node.Spec.Taints, copyNode.Spec.Taints)
		if labelChanged || taintChanged {

			orginal, _ := json.Marshal(node)
			modified, _ := json.Marshal(copyNode)
			patch, err := strategicpatch.CreateTwoWayMergePatch(orginal, modified, node)
			if err != nil {
				return err
			}
			_, err = nc.client.CoreV1().Nodes().Patch(node.Name, types.StrategicMergePatchType, patch)
			if err != nil {
				log.Errorf("update node err: %v", err)
				return err
			}
			log.V(2).Infof("Delete labels and taints from node %v, patch %v", node.Name, string(patch))
		}

	}

	// ensure labels and taints in cur nodes
	for _, node := range desiredNodes.NodesInUse {
		copyNode := node.DeepCopy()

		// change labels
		for k, v := range desiredNodes.Labels {
			copyNode.Labels[k] = v
		}

		// override taint, add or delete
		_, newTaints, _ := taints.ReorganizeTaints(copyNode, true, desiredNodes.TaintsToAdd, desiredNodes.TaintsToDelete)
		// If you don't judge， it maybe change from nil to []Taint{}
		// do not change taints when length of original and new taints are both equal to 0
		if !(len(copyNode.Spec.Taints) == 0 && len(newTaints) == 0) {
			copyNode.Spec.Taints = newTaints
		}

		labelChanged := !reflect.DeepEqual(node.Labels, copyNode.Labels)
		taintChanged := !reflect.DeepEqual(node.Spec.Taints, copyNode.Spec.Taints)
		if labelChanged || taintChanged {

			orginal, _ := json.Marshal(node)
			modified, _ := json.Marshal(copyNode)
			patch, err := strategicpatch.CreateTwoWayMergePatch(orginal, modified, node)
			if err != nil {
				return err
			}
			_, err = nc.client.CoreV1().Nodes().Patch(node.Name, types.StrategicMergePatchType, patch)
			if err != nil {
				log.Errorf("update node err: %v", err)
				return err
			}
			log.V(2).Infof("Ensure labels and taints for requested nodes %v, patch %v", node.Name, string(patch))
		}
	}

	return nil
}
