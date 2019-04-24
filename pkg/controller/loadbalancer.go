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
	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

// LoadBalancerController is responsible for synchronizing LoadBalancer objects stored
// in the system with actual running proxies and providers.
type LoadBalancerController struct {
	client   kubernetes.Interface
	factory  informers.SharedInformerFactory
	lbLister lblisters.LoadBalancerLister
	nodeCtl  *nodeController
	queue    *syncqueue.SyncQueue
}

// NewLoadBalancerController creates a new LoadBalancerController.
func NewLoadBalancerController(cfg config.Configuration) *LoadBalancerController {
	// TODO register metrics
	factory := informers.NewSharedInformerFactory(cfg.Client, 0)
	lbinformer := factory.Loadbalance().V1alpha2().LoadBalancers()
	lbc := &LoadBalancerController{
		client:   cfg.Client,
		factory:  factory,
		lbLister: lbinformer.Lister(),
		nodeCtl: &nodeController{
			client:     cfg.Client,
			nodeLister: factory.Core().V1().Nodes().Lister(),
		},
	}
	// setup lb controller helper
	lbc.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, lbc.syncLoadBalancer)

	// setup informer
	lbinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    lbc.addLoadBalancer,
		UpdateFunc: lbc.updateLoadBalancer,
		DeleteFunc: lbc.deleteLoadBalancer,
	})

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
	if err := lbapi.ValidateLoadBalancer(lb); err != nil {
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

	nlb := lb.DeepCopy()
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
	return lbc.nodeCtl.syncNodes(lb)
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
