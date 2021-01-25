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
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/caicloud/loadbalancer-controller/pkg/plugin"
	"github.com/caicloud/loadbalancer-controller/pkg/provider"
	"github.com/caicloud/loadbalancer-controller/pkg/proxy"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	log "k8s.io/klog"
)

// LoadBalancerController is responsible for synchronizing LoadBalancer objects stored
// in the system with actual running proxies and providers.
type LoadBalancerController struct {
	client    kubernetes.Interface
	factory   informers.SharedInformerFactory
	lbLister  lblisters.LoadBalancerLister
	nodeCtl   *nodeController
	queue     *syncqueue.SyncQueue
	proxies   *plugin.Registry
	providers *plugin.Registry
}

// NewLoadBalancerController creates a new LoadBalancerController.
func NewLoadBalancerController(cfg config.Configuration) *LoadBalancerController {
	// TODO register metrics
	factory := informers.NewSharedInformerFactory(cfg.Client, 0)
	lbinformer := factory.Custom().Loadbalance().V1alpha2().LoadBalancers()
	lbc := &LoadBalancerController{
		client:   cfg.Client,
		factory:  factory,
		lbLister: lbinformer.Lister(),
		nodeCtl: &nodeController{
			client:     cfg.Client,
			nodeLister: factory.Native().Core().V1().Nodes().Lister(),
		},
		proxies:   plugin.NewRegistry(),
		providers: plugin.NewRegistry(),
	}
	_ = proxy.AddToRegistry(lbc.proxies)
	_ = provider.AddToRegistry(lbc.providers)

	// setup lb controller helper
	lbc.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, lbc.syncLoadBalancer)

	// setup informer
	lbinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    lbc.addLoadBalancer,
		UpdateFunc: lbc.updateLoadBalancer,
		DeleteFunc: lbc.deleteLoadBalancer,
	})

	// setup proxies
	lbc.proxies.InitAll(cfg, factory)
	// setup providers
	lbc.providers.InitAll(cfg, factory)

	return lbc
}

// Run begins watching and syncing.
func (lbc *LoadBalancerController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	log.Info("Startting loadbalancer controller")
	defer log.Info("Shutting down loadbalancer controller")

	// ensure loadbalancer tpr initialized
	if err := lbc.ensureResource(); err != nil {
		log.Errorf("Ensure loadbalancer resource error: %v", err)
		return
	}

	// start shared informer
	log.Info("Startting informer factory")
	lbc.factory.Start(stopCh)

	// wait cache synced
	log.Info("Wait for all caches synced")
	if err := lbc.factory.WaitForCacheSync(stopCh); err != nil {
		log.Errorf("Wait for cache sync error %v", err)
	}
	log.Infof("All caches have synced, Running LoadBalancer Controller, workers %v", workers)

	defer func() {
		log.Info("Shuttingdown controller queue")
		lbc.queue.ShutDown()
	}()

	// start loadbalancer worker
	lbc.queue.Run(workers)

	// run proxy
	lbc.proxies.RunAll(stopCh)
	// run providers
	lbc.providers.RunAll(stopCh)

	<-stopCh
}

// ensure loadbalancer tpr initialized
func (lbc *LoadBalancerController) ensureResource() error {
	// set x-kubernetes-preserve-unknown-fields to true, stops the API server
	// decoding step from pruning fields which are not specified
	// in the validation schema.
	xPreserveUnknownFields := true
	crd := &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loadbalancers." + lbapi.GroupName,
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group: lbapi.GroupName,
			Scope: apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:   "loadbalancers",
				Singular: "loadbalancer",
				Kind:     "LoadBalancer",
				ListKind: "LoadBalancerList",
				ShortNames: []string{
					"lb",
				},
			},
			Versions: []apiextensions.CustomResourceDefinitionVersion{
				{
					Name: lbapi.SchemeGroupVersion.Version,
					AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
						{
							Name:     "VIP",
							Type:     "string",
							JSONPath: ".spec.providers.*.vip",
						},
						{
							Name:     "VIPS",
							Type:     "string",
							JSONPath: ".spec.providers.*.vips",
						},
						{
							Name:     "NODES",
							Type:     "string",
							JSONPath: ".spec.nodes.names",
						},
					},
					Schema: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
							// x-kubernetes-preserve-unknown-fields
							XPreserveUnknownFields: &xPreserveUnknownFields,
						},
					},
					Served:  true,
					Storage: true,
				},
			},
		},
	}
	_, err := lbc.client.Apiextensions().ApiextensionsV1().CustomResourceDefinitions().Create(crd)

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
		log.Errorf("invalid loadbalancer scheme: %v", err)
		return err
	}

	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(lb)

	startTime := time.Now()
	defer func() {
		log.V(5).Infof("Finished syncing loadbalancer, key: %v, uesdTime: %v", key, time.Since(startTime))
	}()

	nlb, err := lbc.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		log.Warningf("LoadBalancer %v has been deleted", key)
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
	lbc.proxies.SyncAll(lb)
	// sync provider
	lbc.providers.SyncAll(lb)

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
	log.Infof("Adding LoadBalancer %v", lb.Name)
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

	log.Infof("Updating LoadBalancer %v", old.Name)
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

	log.Infof("Deleting LoadBalancer %v/%v", lb.Namespace, lb.Name)

	lbc.queue.Enqueue(lb)
}
