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
	releaselisters "github.com/caicloud/clientset/listers/release/v1alpha1"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	releaseapi "github.com/caicloud/clientset/pkg/apis/release/v1alpha1"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/caicloud/loadbalancer-controller/pkg/api"

	"github.com/caicloud/loadbalancer-controller/pkg/plugin"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	log "k8s.io/klog"
)

const (
	proxyNameSuffix  = "-proxy-kong"
	proxyName        = "kong"
	defaultNamespace = "kube-system"
	releaseKey       = "controller.caicloud.io/release"
	ingressClassKey  = "loadbalance.caicloud.io/ingress.class"
	defaultIngressClass = "kong"
)

type kong struct {
	initialized           bool

	client kubernetes.Interface
	queue  *syncqueue.SyncQueue

	lbLister  lblisters.LoadBalancerLister
	podLister corelisters.PodLister
	releaseLister releaselisters.ReleaseLister
}

// New creates a new kong proxy plugin
func New() plugin.Interface {
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

	// init crds
	_ = installKongCrds()

	// initialize controller
	lbInformer := sif.Loadbalance().V1alpha2().LoadBalancers()
	podInfomer := sif.Core().V1().Pods()
	releaseInformer := sif.Release().V1alpha1().Releases()

	f.lbLister = lbInformer.Lister()
	f.podLister = podInfomer.Lister()
	f.releaseLister = releaseInformer.Lister()

	f.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, f.syncLoadBalancer)

	releaseInformer.Informer().AddEventHandler(newEventHandlerForRelease(f.lbLister, f.releaseLister, f.queue, f.releaseFiltered))
}

func (f *kong) Run(stopCh <-chan struct{}) {
	workers := 1
	if !f.initialized {
		panic("Please initialize proxy before you run it")
	}

	defer utilruntime.HandleCrash()

	log.Infof("Starting kong proxy, workers %v", workers)

	// lb controller has waited all the informer synced
	// there is no need to wait again here

	defer func() {
		log.Info("Shutting down kong proxy")
		f.queue.ShutDown()
	}()

	f.queue.Run(workers)

	<-stopCh

}

func (f *kong) releaseFiltered(obj *releaseapi.Release) bool {
	return f.filteredByLabel(obj)
}

func (f *kong) filteredByLabel(obj metav1.ObjectMetaAccessor) bool {
	// obj.Labels
	selector := labels.Set{lbapi.LabelKeyProxy: proxyName}.AsSelector()
	match := selector.Matches(labels.Set(obj.GetObjectMeta().GetLabels()))

	return !match
}

func (f *kong) OnSync(lb *lbapi.LoadBalancer) {
	log.Infof("Syncing proxy, triggered by loadbalancer %v/%v", lb.Namespace, lb.Name)
	f.queue.Enqueue(lb)
}

// TODO use event
// sync release with loadbalancer
// the obj will be *lbapi.LoadBalancer
func (f *kong) syncLoadBalancer(obj interface{}) error {
	lb, ok := obj.(*lbapi.LoadBalancer)
	if !ok {
		return fmt.Errorf("expect loadbalancer, got %v", obj)
	}

	key, _ := cache.DeletionHandlingMetaNamespaceKeyFunc(lb)

	startTime := time.Now()
	defer func() {
		log.V(5).Infof("Finished syncing kong proxy for %v, usedTime %v", key, time.Since(startTime))
	}()

	nlb, err := f.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		log.Warningf("LoadBalancer %v has been deleted, clean up proxy", key)

		return f.cleanup(lb)
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
		log.Infof("Loadbalancer %v is not a kong loadbalancer", lb.Name)
		return nil
	}

	release, err := f.getOrSetReleaseForLoadBalancer(lb)
	if err != nil {
 		log.Errorf("Wait for release ready, error %v", err)
		return nil
	}

	if lb.DeletionTimestamp != nil {
		// TODO sync status only
		return nil
	}

	return f.sync(lb, release)

}

func (f *kong) getOrSetReleaseForLoadBalancer(lb *lbapi.LoadBalancer) (*releaseapi.Release, error) {
	// get release name from annotations
	annotations := lb.GetAnnotations()
	if annotations == nil {
		return nil, fmt.Errorf("No release annotation found for loadbalancer %v", lb.Name)
	}
	releaseName := annotations[releaseKey]
	if releaseName == "" {
		return nil, fmt.Errorf("No release annotation found for loadbalancer %v", lb.Name)
	}

	// get release
	release, err := f.releaseLister.Releases(defaultNamespace).Get(releaseName)
	if err != nil {
		log.Errorf("Get release for loadbalancer %v error", lb.Name)
		return nil, err
	}

	// set owner reference
	t := true
	if release.OwnerReferences == nil || len(release.OwnerReferences) == 0 {
		release.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         api.ControllerKind.GroupVersion().String(),
				Kind:               api.ControllerKind.Kind,
				Name:               lb.Name,
				UID:                lb.UID,
				Controller:         &t,
				BlockOwnerDeletion: &t,
			},
		}
		if release, err = f.client.ReleaseV1alpha1().Releases(defaultNamespace).Update(release); err != nil {
			log.Errorf("Set owner references for release %v of loadbalancer %v error", release.Name, lb.Name)
			return nil, err
		}
	}

	return release, nil
}

// sync release
func (f *kong) sync(lb *lbapi.LoadBalancer, release *releaseapi.Release) error {
    // get ingress class from loadbalancer
	annotations := lb.GetAnnotations()
	if annotations == nil {
		return fmt.Errorf("No annotations on loadbalancer %v/%v", lb.Namespace, lb.Name)
	}
    ingressClass := annotations[ingressClassKey]
    if ingressClass == "" {
    	ingressClass = defaultIngressClass
    	annotations[ingressClass] = defaultIngressClass
    	lb.Annotations = annotations
    }
	// update status
	return f.syncStatus(lb, release.Name, ingressClass)
}

func (f *kong) cleanup(lb *lbapi.LoadBalancer) error {
    // clean up ingress
    selector := labels.Set{
        // createdby ingressClass
        lbapi.LabelKeyCreatedBy: lb.Status.ProxyStatus.IngressClass,
    }
    ingresses, err := f.client.ExtensionsV1beta1().Ingresses(metav1.NamespaceAll).List(metav1.ListOptions{
        LabelSelector: selector.String(),
    })

    if err != nil {
        log.Errorf("Cleanup Ingress error: %v", err)
        return err
    }

    for _, ingress := range ingresses.Items {
        err = f.client.ExtensionsV1beta1().Ingresses(ingress.Namespace).Delete(ingress.Name, &metav1.DeleteOptions{})
        if err != nil {
            log.Errorf("Cleanup Ingress error: %v", err)
            return err
        }
    }
	return nil
}