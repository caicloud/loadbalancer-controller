/*
Copyright 2020 Caicloud authors. All rights reserved.

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

package kong

import (
	"fmt"
	"time"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/caicloud/loadbalancer-controller/pkg/proxy"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	log "github.com/zoumo/logdog"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	proxyNameSuffix = "-proxy-kong"
	proxyName       = "kong"
)

func init() {
	proxy.RegisterPlugin(proxyName, NewKong())
}

var _ proxy.Plugin = &kong{}

type kong struct {
	initialized bool

	client kubernetes.Interface
	queue  *syncqueue.SyncQueue

	lbLister  lblisters.LoadBalancerLister
	dLister   appslisters.DeploymentLister
	podLister corelisters.PodLister
}

// NewKong creates a new kong proxy plugin
func NewKong() proxy.Plugin {
	return &kong{}
}

func (f *kong) Init(cfg config.Configuration, sif informers.SharedInformerFactory) {
	if f.initialized {
		return
	}
	f.initialized = true

	log.Info("Initialize the kong proxy")
	// set config
	f.client = cfg.Client

	// initialize controller
	lbInformer := sif.Loadbalance().V1alpha2().LoadBalancers()
	dInformer := sif.Apps().V1().Deployments()
	podInfomer := sif.Core().V1().Pods()

	f.lbLister = lbInformer.Lister()
	f.dLister = dInformer.Lister()
	f.podLister = podInfomer.Lister()

	f.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, f.syncLoadBalancer)

	dInformer.Informer().AddEventHandler(lbutil.NewEventHandlerForDeployment(f.lbLister, f.dLister, f.queue, f.deploymentFiltered))
	podInfomer.Informer().AddEventHandler(lbutil.NewEventHandlerForSyncStatusWithPod(f.lbLister, f.podLister, f.queue, f.podFiltered))
}

func (f *kong) Run(stopCh <-chan struct{}) {
	workers := 1
	if !f.initialized {
		log.Panic("Please initialize proxy before you run it")
		return
	}

	defer utilruntime.HandleCrash()

	log.Info("Starting kong proxy", log.Fields{
		"workers": workers,
	})

	// lb controller has waited all the informer synced
	// there is no need to wait again here

	defer func() {
		log.Info("Shutting down kong proxy")
		f.queue.ShutDown()
	}()

	f.queue.Run(workers)

	<-stopCh

}

func (f *kong) selector(lb *lbapi.LoadBalancer) labels.Set {
	return labels.Set{
		// TODO: get ingress class from loadbalancer spec
		lbapi.LabelKeyCreatedBy: fmt.Sprintf(lbapi.LabelValueFormatCreateby, lb.Namespace, "kong"),
		lbapi.LabelKeyProxy:     proxyName,
	}
}

// filter Deployment that controller does not care
func (f *kong) deploymentFiltered(obj *appsv1.Deployment) bool {
	return f.filteredByLabel(obj)
}

func (f *kong) podFiltered(obj *v1.Pod) bool {
	return f.filteredByLabel(obj)
}

func (f *kong) filteredByLabel(obj metav1.ObjectMetaAccessor) bool {
	// obj.Labels
	selector := labels.Set{lbapi.LabelKeyProxy: proxyName}.AsSelector()
	match := selector.Matches(labels.Set(obj.GetObjectMeta().GetLabels()))

	return !match
}

func (f *kong) OnSync(lb *lbapi.LoadBalancer) {
	log.Info("Syncing proxy, triggered by lb controller", log.Fields{"lb": lb.Name, "namespace": lb.Namespace})
	f.queue.Enqueue(lb)
}

// TODO use event
// sync deployment with loadbalancer
// the obj will be *lbapi.LoadBalancer
func (f *kong) syncLoadBalancer(obj interface{}) error {
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
		log.Debug("Finished syncing kong proxy", log.Fields{"lb": key, "usedTime": time.Since(startTime)})
	}()

	nlb, err := f.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		log.Warn("LoadBalancer has been deleted, clean up proxy", log.Fields{"lb": key})

		// return f.cleanup(lb)
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

	lb = nlb.DeepCopy()

	if lb.Spec.Proxy.Type != lbapi.ProxyTypeKong {
		// It is not my responsible, clean up legacies
		return fmt.Errorf("Type is invalid: %v", lb.Spec.Proxy.Type)
	}

	ds, err := f.getDeploymentsForLoadBalancer(lb)
	if err != nil {
		return err
	}

	if lb.DeletionTimestamp != nil {
		// TODO sync status only
		return nil
	}

	return f.sync(lb, ds)

}

func (f *kong) getDeploymentsForLoadBalancer(lb *lbapi.LoadBalancer) ([]*appsv1.Deployment, error) {
	// construct selector
	selector := f.selector(lb).AsSelector()

	// list all
	dList, err := f.dLister.Deployments(lb.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	return dList, nil
}

// sync generate desired deployment from lb and compare it with existing deployment
func (f *kong) sync(lb *lbapi.LoadBalancer, dps []*appsv1.Deployment) error {
	return f.syncStatus(lb, dps[0])
}
