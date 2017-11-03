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

package nginx

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/caicloud/clientset/util/syncqueue"

	"github.com/caicloud/clientset/informers"
	"github.com/caicloud/clientset/kubernetes"
	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	controllerutil "github.com/caicloud/clientset/util/controller"
	"github.com/caicloud/loadbalancer-controller/pkg/api"
	"github.com/caicloud/loadbalancer-controller/pkg/config"
	"github.com/caicloud/loadbalancer-controller/pkg/proxy"
	lbutil "github.com/caicloud/loadbalancer-controller/pkg/util/lb"
	"github.com/caicloud/loadbalancer-controller/pkg/util/validation"
	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	corelisters "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
)

const (
	proxyNameSuffix = "-proxy-nginx"
	proxyName       = "nginx"
)

func init() {
	proxy.RegisterPlugin(proxyName, NewNginx())
}

var _ proxy.Plugin = &nginx{}

type nginx struct {
	initialized           bool
	image                 string
	sidecar               string
	defaultHTTPbackend    string
	defaultSSLCertificate string

	client kubernetes.Interface
	queue  *syncqueue.SyncQueue

	lbLister  lblisters.LoadBalancerLister
	dLister   extensionslisters.DeploymentLister
	podLister corelisters.PodLister
}

// NewNginx creates a new nginx proxy plugin
func NewNginx() proxy.Plugin {
	return &nginx{}
}

func (f *nginx) Init(cfg config.Configuration, sif informers.SharedInformerFactory) {
	if f.initialized {
		return
	}
	f.initialized = true

	log.Info("Initialize the nginx proxy")
	// set config
	f.defaultHTTPbackend = cfg.Proxies.DefaultHTTPBackend
	f.defaultSSLCertificate = cfg.Proxies.DefaultSSLCertificate
	f.sidecar = cfg.Proxies.Sidecar.Image
	f.image = cfg.Proxies.Nginx.Image
	f.client = cfg.Client

	// initialize controller
	lbInformer := sif.Loadbalance().V1alpha2().LoadBalancers()
	dInformer := sif.Extensions().V1beta1().Deployments()
	podInfomer := sif.Core().V1().Pods()

	f.lbLister = lbInformer.Lister()
	f.dLister = dInformer.Lister()
	f.podLister = podInfomer.Lister()

	f.queue = syncqueue.NewPassthroughSyncQueue(&lbapi.LoadBalancer{}, f.syncLoadBalancer)

	dInformer.Informer().AddEventHandler(lbutil.NewEventHandlerForDeployment(f.lbLister, f.dLister, f.queue, f.deploymentFiltered))
	podInfomer.Informer().AddEventHandler(lbutil.NewEventHandlerForSyncStatusWithPod(f.lbLister, f.podLister, f.queue, f.podFiltered))
}

func (f *nginx) Run(stopCh <-chan struct{}) {
	workers := 1
	if !f.initialized {
		log.Panic("Please initialize proxy before you run it")
		return
	}

	defer utilruntime.HandleCrash()

	log.Info("Starting nginx proxy", log.Fields{
		"workers":              workers,
		"image":                f.image,
		"default-http-backend": f.defaultHTTPbackend,
		"sidecar":              f.sidecar,
	})
	defer log.Info("Shutting down nginx proxy")

	if err := f.ensureDefaultHTTPBackend(); err != nil {
		log.Error("ensure default http backend service error", log.Fields{"err": err})
		return
	}

	// lb controller has waited all the informer synced
	// there is no need to wait again here

	defer func() {
		log.Info("Shutting down nginx proxy")
		f.queue.ShutDown()
	}()

	f.queue.Run(workers)

	<-stopCh

}

func (f *nginx) selector(lb *lbapi.LoadBalancer) labels.Set {
	return labels.Set{
		lbapi.LabelKeyCreatedBy: fmt.Sprintf(lbapi.LabelValueFormatCreateby, lb.Namespace, lb.Name),
		lbapi.LabelKeyProxy:     proxyName,
	}
}

// filter Deployment that controller does not care
func (f *nginx) deploymentFiltered(obj *extensions.Deployment) bool {
	return f.filteredByLabel(obj)
}

func (f *nginx) podFiltered(obj *v1.Pod) bool {
	return f.filteredByLabel(obj)
}

func (f *nginx) filteredByLabel(obj metav1.ObjectMetaAccessor) bool {
	// obj.Labels
	selector := labels.Set{lbapi.LabelKeyProxy: proxyName}.AsSelector()
	match := selector.Matches(labels.Set(obj.GetObjectMeta().GetLabels()))

	return !match
}

func (f *nginx) OnSync(lb *lbapi.LoadBalancer) {
	if lb.Spec.Proxy.Type != lbapi.ProxyTypeNginx {
		// It is not my responsible
		return
	}
	log.Info("Syncing proxy, triggered by lb controller", log.Fields{"lb": lb.Name, "namespace": lb.Namespace})
	f.queue.Enqueue(lb)
}

// TODO use event
// sync deployment with loadbalancer
// the obj will be *lbapi.LoadBalancer
func (f *nginx) syncLoadBalancer(obj interface{}) error {
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
		log.Debug("Finished syncing nginx proxy", log.Fields{"lb": key, "usedTime": time.Since(startTime)})
	}()

	nlb, err := f.lbLister.LoadBalancers(lb.Namespace).Get(lb.Name)
	if errors.IsNotFound(err) {
		log.Warn("LoadBalancer has been deleted, clean up proxy", log.Fields{"lb": key})

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

	lb, err = api.LoadBalancerDeepCopy(nlb)
	if err != nil {
		return err
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

func (f *nginx) getDeploymentsForLoadBalancer(lb *lbapi.LoadBalancer) ([]*extensions.Deployment, error) {

	// construct selector
	selector := f.selector(lb).AsSelector()

	// list all
	dList, err := f.dLister.Deployments(lb.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached quorum read sometime after listing deployment (see kubernetes#42639).
	canAdoptFunc := controllerutil.RecheckDeletionTimestamp(func() (metav1.Object, error) {
		// fresh lb
		fresh, err := f.client.LoadbalanceV1alpha2().LoadBalancers(lb.Namespace).Get(lb.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if fresh.UID != lb.UID {
			return nil, fmt.Errorf("original LoadBalancer %v/%v is gone: got uid %v, wanted %v", lb.Namespace, lb.Name, fresh.UID, lb.UID)
		}
		return fresh, nil
	})

	cm := controllerutil.NewDeploymentControllerRefManager(f.client, lb, selector, api.ControllerKind, canAdoptFunc)
	return cm.Claim(dList)
}

// sync generate desired deployment from lb and compare it with existing deployment
func (f *nginx) sync(lb *lbapi.LoadBalancer, dps []*extensions.Deployment) error {
	desiredDeploy := f.generateDeployment(lb)

	// update
	var err error
	updated := false
	activeDeploy := desiredDeploy

	for _, dp := range dps {

		// two conditions will trigger controller to scale down deployment
		// 1. deployment does not have auto-generated prefix
		// 2. if there are more than one active controllers, there may be many valid deployments.
		//    But we only need one.
		if !strings.HasPrefix(dp.Name, lb.Name+proxyNameSuffix) || updated {
			if *dp.Spec.Replicas == 0 {
				continue
			}
			// scale unexpected deployment replicas to zero
			log.Info("Scale unexpected proxy replicas to zero", log.Fields{"d.name": dp.Name, "lb.name": lb.Name})
			copy, _ := api.DeploymentDeepCopy(dp)
			replica := int32(0)
			copy.Spec.Replicas = &replica
			f.client.ExtensionsV1beta1().Deployments(lb.Namespace).Update(copy)
			continue
		}

		updated = true
		copyDp, changed, newErr := f.ensureDeployment(desiredDeploy, dp)
		if newErr != nil {
			err = newErr
			continue
		}
		if changed {
			log.Info("Sync nginx for lb", log.Fields{"d.name": dp.Name, "lb.name": lb.Name})
			_, err = f.client.ExtensionsV1beta1().Deployments(lb.Namespace).Update(copyDp)
			if err != nil {
				return err
			}
		}
		activeDeploy = copyDp
	}

	// len(dps) == 0 or no deployment's name match desired deployment
	if !updated {
		// create deployment
		log.Info("Create nginx for lb", log.Fields{"d.name": desiredDeploy.Name, "lb.name": lb.Name})
		_, err = f.client.ExtensionsV1beta1().Deployments(lb.Namespace).Create(desiredDeploy)
		if err != nil {
			return err
		}
	}

	err = f.ensureConfigMaps(lb)
	if err != nil {
		return err
	}

	// update status
	return f.syncStatus(lb, activeDeploy)
}

func (f *nginx) ensureDeployment(desiredDeploy, oldDeploy *extensions.Deployment) (*extensions.Deployment, bool, error) {
	copyDp, err := api.DeploymentDeepCopy(oldDeploy)
	if err != nil {
		return nil, false, err
	}

	// ensure labels
	for k, v := range desiredDeploy.Labels {
		copyDp.Labels[k] = v
	}
	// ensure replicas
	copyDp.Spec.Replicas = desiredDeploy.Spec.Replicas
	// ensure containers
	var containersChanged = false
	copyContainers := copyDp.Spec.Template.Spec.Containers
	desiredContainers := desiredDeploy.Spec.Template.Spec.Containers
	if len(copyContainers) != len(desiredContainers) {
		containersChanged = true
	} else {
		for _, c1 := range desiredContainers {
			found := false
			for _, c2 := range copyContainers {
				if c1.Name == c2.Name {
					found = true
					if c1.Image != c2.Image {
						containersChanged = true
					}
					break
				}
			}
			if !found {
				containersChanged = true
			}
		}
	}

	if containersChanged {
		copyDp.Spec.Template.Spec.Containers = desiredContainers
	}

	// ensure nodeaffinity
	copyDp.Spec.Template.Spec.Affinity.NodeAffinity = desiredDeploy.Spec.Template.Spec.Affinity.NodeAffinity

	// check if changed
	nodeAffinityChanged := !reflect.DeepEqual(copyDp.Spec.Template.Spec.Affinity.NodeAffinity, oldDeploy.Spec.Template.Spec.Affinity.NodeAffinity)
	labelChanged := !reflect.DeepEqual(copyDp.Labels, oldDeploy.Labels)
	replicasChanged := *(copyDp.Spec.Replicas) != *(oldDeploy.Spec.Replicas)

	changed := labelChanged || replicasChanged || nodeAffinityChanged || containersChanged
	if changed {
		log.Info("Abount to correct nginx proxy", log.Fields{
			"dp.name":             copyDp.Name,
			"labelChanged":        labelChanged,
			"replicasChanged":     replicasChanged,
			"nodeAffinityChanged": nodeAffinityChanged,
			"containersChanged":   containersChanged,
		})
	}

	return copyDp, changed, nil
}

// cleanup deployment and other resource controlled by lb proxy
func (f *nginx) cleanup(lb *lbapi.LoadBalancer) error {

	selector := f.selector(lb)

	ds, err := f.getDeploymentsForLoadBalancer(lb)
	if err != nil {
		return err
	}

	policy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(30)

	for _, d := range ds {
		err = f.client.ExtensionsV1beta1().Deployments(d.Namespace).Delete(d.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
			PropagationPolicy:  &policy,
		})
		if err != nil {
			log.Warn("Cleanup proxy error", log.Fields{"ns": d.Namespace, "d.name": d.Name, "err": err})
			return err
		}
	}

	// clean up config map
	err = f.client.CoreV1().ConfigMaps(lb.Namespace).DeleteCollection(nil, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		log.Warn("Cleanup ConfigMap error", log.Fields{"err": err})
		return err
	}

	// clean up ingress
	selector = labels.Set{
		// createdby ingressClass
		lbapi.LabelKeyCreatedBy: fmt.Sprintf(lbapi.LabelValueFormatCreateby, lb.Namespace, lb.Name),
	}
	ingresses, err := f.client.ExtensionsV1beta1().Ingresses(metav1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: selector.String(),
	})

	if err != nil {
		log.Warn("Cleanup Ingress error", log.Fields{"err": err})
		return err
	}

	for _, ingress := range ingresses.Items {
		err = f.client.ExtensionsV1beta1().Ingresses(ingress.Namespace).Delete(ingress.Name, &metav1.DeleteOptions{})
		if err != nil {
			log.Warn("Cleanup Ingress error", log.Fields{"err": err})
			return err
		}
	}

	return nil
}
