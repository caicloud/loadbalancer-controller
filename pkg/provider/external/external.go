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

package external

import (
	"fmt"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/caicloud/loadbalancer-controller/pkg/plugin"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"

	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	log "k8s.io/klog"
)

const (
	providerName = "external"
)

type external struct {
	initialized bool

	client kubernetes.Interface
	queue  *syncqueue.SyncQueue

	lbLister lblisters.LoadBalancerLister
}

// New creates a new external provider plugin
func New() plugin.Interface {
	return &external{}
}

func (f *external) Init(cfg config.Configuration, sif informers.SharedInformerFactory) {
	if f.initialized {
		return
	}
	f.initialized = true
	log.Info("Initialize the external provider")

	// set config
	f.client = cfg.Client
	// initialize controller
	lbInformer := sif.Loadbalance().V1alpha2().LoadBalancers()
	f.lbLister = lbInformer.Lister()
	f.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, f.syncLoadBalancer)

}

func (f *external) Run(stopCh <-chan struct{}) {

	workers := 1

	if !f.initialized {
		panic("Please initialize provider before you run it")
		return
	}

	defer utilruntime.HandleCrash()

	log.Infof("Starting external provider, workers %v", workers)
	defer log.Info("Shutting down external provider")

	// lb controller has waited all the informer synced
	// there is no need to wait again here

	defer func() {
		log.Info("Shutting down external provider")
		f.queue.ShutDown()
	}()

	f.queue.Run(workers)

	<-stopCh
}

func (f *external) OnSync(lb *lbapi.LoadBalancer) {

	log.Infof("Syncing providers, triggered by loadbalancer %v/%v", lb.Namespace, lb.Name)
	f.queue.Enqueue(lb)
}

func (f *external) syncLoadBalancer(obj interface{}) error {
	lb, ok := obj.(*lbapi.LoadBalancer)
	if !ok {
		return fmt.Errorf("expect loadbalancer, got %v", obj)
	}

	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(lb)

	nlb, err := f.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Unable to retrieve LoadBalancer %v from store: %v", key, err))
		return err
	}

	// fresh lb
	if lb.UID != nlb.UID {
		return nil
	}

	lb = nlb.DeepCopy()

	if lb.Spec.Providers.External == nil {
		// It is not my responsible, clean up legacies
		return f.deleteStatus(lb)
	}

	if lb.DeletionTimestamp != nil {
		// TODO sync status only
		return nil
	}

	// sync status
	providerStatus := lbapi.ExpternalProviderStatus{
		VIP: lb.Spec.Providers.External.VIP,
	}
	externalstatus := lb.Status.ProvidersStatuses.External
	// check whether the statuses are equal
	if externalstatus == nil || !lbutil.ExternalProviderStatusEqual(*externalstatus, providerStatus) {
		_, err := lbutil.UpdateLBWithRetries(
			f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
			f.lbLister,
			lb.Namespace,
			lb.Name,
			func(lb *lbapi.LoadBalancer) error {
				lb.Status.ProvidersStatuses.External = &providerStatus
				return nil
			},
		)

		if err != nil {
			log.Errorf("Update loadbalancer status error, %v", err)
			return err
		}

	}
	return nil
}

func (f *external) deleteStatus(lb *lbapi.LoadBalancer) error {
	if lb.Status.ProvidersStatuses.External == nil {
		return nil
	}

	log.Infof("delete external status for loadbalancer %v/%v", lb.Namespace, lb.Name)
	_, err := lbutil.UpdateLBWithRetries(
		f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace),
		f.lbLister,
		lb.Namespace,
		lb.Name,
		func(lb *lbapi.LoadBalancer) error {
			lb.Status.ProvidersStatuses.External = nil
			return nil
		},
	)

	if err != nil {
		log.Errorf("Update loadbalancer status error: %v", err)
		return err
	}
	return nil
}
