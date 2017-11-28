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
	"time"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	apiextensions "github.com/caicloud/clientset/pkg/apis/apiextensions/v1beta1"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/caicloud/loadbalancer-controller/pkg/provider"
	"github.com/caicloud/loadbalancer-controller/pkg/proxy"
	"github.com/caicloud/loadbalancer-controller/pkg/util/taints"
	"github.com/caicloud/loadbalancer-controller/pkg/util/validation"
	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	corelisters "k8s.io/client-go/listers/core/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

// LoadBalancerController is responsible for synchronizing LoadBalancer objects stored
// in the system with actual running proxies and providers.
type LoadBalancerController struct {
	client kubernetes.Interface

	factory    informers.SharedInformerFactory
	lbLister   lblisters.LoadBalancerLister
	nodeLister corelisters.NodeLister
	queue      *syncqueue.SyncQueue
}

// NewLoadBalancerController creates a new LoadBalancerController.
func NewLoadBalancerController(cfg config.Configuration) *LoadBalancerController {
	// TODO register metrics

	lbc := &LoadBalancerController{
		client:  cfg.Client,
		factory: informers.NewSharedInformerFactory(cfg.Client, 0),
	}
	// setup lb controller helper
	lbc.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, lbc.syncLoadBalancer)

	// setup informer
	lbinformer := lbc.factory.Loadbalance().V1alpha2().LoadBalancers()
	lbinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    lbc.addLoadBalancer,
		UpdateFunc: lbc.updateLoadBalancer,
		DeleteFunc: lbc.deleteLoadBalancer,
	})

	lbc.lbLister = lbinformer.Lister()
	lbc.nodeLister = lbc.factory.Core().V1().Nodes().Lister()

	// setup proxies
	proxy.Init(cfg, lbc.factory)
	// setup providers
	provider.Init(cfg, lbc.factory)

	return lbc
}

// Run begins watching and syncing.
func (lbc *LoadBalancerController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	log.Info("Startting loadbalancer controller")
	defer log.Info("Shutting down loadbalancer controller")

	// ensure loadbalancer tpr initialized
	if err := lbc.ensureResource(); err != nil {
		log.Error("Ensure loadbalancer resource error", log.Fields{"err": err})
		return
	}

	// start shared informer
	log.Info("Startting informer factory")
	lbc.factory.Start(stopCh)

	// wait cache synced
	log.Info("Wait for all caches synced")
	synced := lbc.factory.WaitForCacheSync(stopCh)
	for tpy, sync := range synced {
		if !sync {
			log.Error("Wait for cache sync timeout", log.Fields{"type": tpy})
			return
		}
	}
	log.Info("All caches have synced, Running LoadBalancer Controller ...", log.Fields{"worker": workers})

	defer func() {
		log.Info("Shuttingdown controller queue")
		lbc.queue.ShutDown()
	}()

	// start loadbalancer worker
	lbc.queue.Run(workers)

	// run proxy
	proxy.Run(stopCh)
	// run providers
	provider.Run(stopCh)

	<-stopCh
}

// ensure loadbalancer tpr initialized
func (lbc *LoadBalancerController) ensureResource() error {
	crd := &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loadbalancers." + lbapi.GroupName,
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   lbapi.GroupName,
			Version: lbapi.SchemeGroupVersion.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   "loadbalancers",
				Singular: "loadbalancer",
				Kind:     "LoadBalancer",
				ListKind: "LoadBalancerList",
			},
		},
	}
	_, err := lbc.client.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)

	if errors.IsAlreadyExists(err) {
		log.Info("Skip the creation for CustomResourceDefinition LoadBalancer because it has already been created")
		return nil
	}

	if err != nil {
		return err
	}

	log.Info("Create CustomResourceDefinition LoadBalancer successfully")

	return nil
}

// syncLoadBalancer will sync the loadbalancer with the given key.
// This function is not meant to be invoked concurrently with the same key.
func (lbc *LoadBalancerController) syncLoadBalancer(obj interface{}) error {
	lb, ok := obj.(*lbapi.LoadBalancer)
	if !ok {
		return fmt.Errorf("expect loadbalancer, got %v", obj)
	}

	// Validate loadbalancer scheme
	if err := validation.ValidateLoadBalancer(lb); err != nil {
		log.Debug("invalid loadbalancer scheme", log.Fields{"err": err})
		return err
	}

	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(lb)

	startTime := time.Now()
	defer func() {
		log.Debug("Finished syncing loadbalancer", log.Fields{"key": key, "usedTime": time.Since(startTime)})
	}()

	nlb, err := lbc.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		log.Warn("LoadBalancer has been deleted", log.Fields{"lb": key})
		// deleted
		return lbc.sync(lb, true)
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Unable to retrieve LoadBalancer %v from store: %v", key, err))
		return err
	}

	// fresh lb
	if lb.UID != nlb.UID {
		//  original loadbalancer is gone
		return nil
	}
	lb = nlb

	return lbc.sync(lb, false)
}

func (lbc *LoadBalancerController) sync(lb *lbapi.LoadBalancer, deleted bool) error {

	nlb, err := lbc.clone(lb)
	if err != nil {
		return err
	}

	lb = nlb

	// sync proxy
	proxy.OnSync(lb)
	// sync provider
	provider.OnSync(lb)

	// sync nodes
	if deleted {
		replicas := int32(0)
		lb.Spec.Nodes = lbapi.NodesSpec{
			Replicas: &replicas,
			Names:    []string{},
		}
	}

	// sync nodes
	err = lbc.syncNodes(lb)

	return err
}

func (lbc *LoadBalancerController) syncNodes(lb *lbapi.LoadBalancer) error {
	// varify desired nodes
	desiredNodes, err := lbc.getVerifiedNodes(lb)
	if err != nil {
		log.Error("varify nodes error", log.Fields{"err": err})
		return err
	}

	oldNodes, err := lbc.getNodesForLoadBalancer(lb)
	if err != nil {
		log.Error("list node error")
		return err
	}
	// compute diff
	nodesToDelete := lbc.nodesDiff(oldNodes, desiredNodes.Nodes)
	return lbc.doLabelAndTaints(nodesToDelete, desiredNodes)
}

func (lbc *LoadBalancerController) getNodesForLoadBalancer(lb *lbapi.LoadBalancer) ([]*apiv1.Node, error) {
	// list old nodes
	labelkey := fmt.Sprintf(lbapi.UniqueLabelKeyFormat, lb.Namespace, lb.Name)
	selector := labels.Set{labelkey: "true"}.AsSelector()
	return lbc.nodeLister.List(selector)
}

func (lbc *LoadBalancerController) nodesDiff(oldNodes, desiredNodes []*apiv1.Node) []*apiv1.Node {

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

func (lbc *LoadBalancerController) addLoadBalancer(obj interface{}) {
	lb := obj.(*lbapi.LoadBalancer)
	log.Info("Adding LoadBalancer", log.Fields{"name": lb.Name})
	lbc.queue.Enqueue(lb)
}

func (lbc *LoadBalancerController) updateLoadBalancer(oldObj, curObj interface{}) {
	old := oldObj.(*lbapi.LoadBalancer)
	cur := curObj.(*lbapi.LoadBalancer)

	if old.ResourceVersion == cur.ResourceVersion {
		// Periodic resync will send update events for all known LoadBalancer.
		// Two different versions of the same LoadBalancer will always have different RVs.
		return
	}

	if reflect.DeepEqual(old.Spec, cur.Spec) {
		return
	}

	log.Info("Updating LoadBalancer", log.Fields{"name": old.Name})
	lbc.queue.EnqueueAfter(cur, 1*time.Second)
}

func (lbc *LoadBalancerController) deleteLoadBalancer(obj interface{}) {
	lb, ok := obj.(*lbapi.LoadBalancer)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		lb, ok = tombstone.Obj.(*lbapi.LoadBalancer)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a LoadBalancer %#v", obj))
			return
		}
	}

	log.Info("Deleting LoadBalancer", log.Fields{"lb.name": lb.Name, "lb.ns": lb.Namespace})

	lbc.queue.Enqueue(lb)
}

func (lbc *LoadBalancerController) clone(lb *lbapi.LoadBalancer) (*lbapi.LoadBalancer, error) {
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

// doLabelAndTaints delete label and taints in nodesToDelete
// add label and taints in nodes
func (lbc *LoadBalancerController) doLabelAndTaints(nodesToDelete []*apiv1.Node, desiredNodes *VerifiedNodes) error {
	// delete labels and taints from old nodes
	for _, node := range nodesToDelete {
		copy, _ := scheme.Scheme.DeepCopy(node)
		copyNode := copy.(*apiv1.Node)

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
			_, err = lbc.client.CoreV1().Nodes().Patch(node.Name, types.StrategicMergePatchType, patch)
			if err != nil {
				log.Errorf("update node err: %v", err)
				return err
			}
			log.Notice("Delete labels and taints from old nodes", log.Fields{
				"node":  node.Name,
				"patch": string(patch),
			})
		}

	}

	// ensure labels and taints in cur nodes
	for _, node := range desiredNodes.Nodes {
		copy, _ := scheme.Scheme.DeepCopy(node)
		copyNode := copy.(*apiv1.Node)

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
			_, err = lbc.client.CoreV1().Nodes().Patch(node.Name, types.StrategicMergePatchType, patch)
			if err != nil {
				log.Errorf("update node err: %v", err)
				return err
			}
			log.Notice("Ensure labels and taints for requested nodes", log.Fields{
				"node":  node.Name,
				"patch": string(patch),
			})
		}
	}

	return nil

}
