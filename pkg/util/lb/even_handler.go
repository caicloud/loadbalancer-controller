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

package lb

import (
	"fmt"
	"reflect"
	"time"

	netv1alpha1 "github.com/caicloud/loadbalancer-controller/pkg/apis/networking/v1alpha1"
	netlisters "github.com/caicloud/loadbalancer-controller/pkg/listers/networking/v1alpha1"
	controllerutil "github.com/caicloud/loadbalancer-controller/pkg/util/controller"
	log "github.com/zoumo/logdog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	corelisters "k8s.io/client-go/listers/core/v1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/pkg/api/v1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/controller"
)

var _ cache.ResourceEventHandler = &EventHandlerForDeployment{}
var _ cache.ResourceEventHandler = &EventHandlerForSyncStatusWithPod{}

// controllerKind contains the schema.GroupVersionKind for this controller type.
var controllerKind = netv1alpha1.SchemeGroupVersion.WithKind(netv1alpha1.LoadBalancerKind)

type filterDeploymentFunc func(obj *extensions.Deployment) bool
type filterPodFunc func(obj *v1.Pod) bool

// EventHandlerForDeployment helps you create a event handler to handle with
// deployments event quickly, makes you focus on you own code
type EventHandlerForDeployment struct {
	helper *controllerutil.Helper

	lbLister netlisters.LoadBalancerLister
	dLister  extensionslisters.DeploymentLister

	filtered filterDeploymentFunc
}

// NewEventHandlerForDeployment ...
func NewEventHandlerForDeployment(
	lbLister netlisters.LoadBalancerLister,
	dLister extensionslisters.DeploymentLister,
	helper *controllerutil.Helper,
	filterFunc filterDeploymentFunc,
) *EventHandlerForDeployment {
	return &EventHandlerForDeployment{
		helper:   helper,
		lbLister: lbLister,
		dLister:  dLister,
		filtered: filterFunc,
	}
}

// OnAdd ...
func (eh *EventHandlerForDeployment) OnAdd(obj interface{}) {
	d := obj.(*extensions.Deployment)

	if d.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		eh.OnDelete(d)
		return
	}

	// filter Deployment that controller does not care
	// this Deployment maybe controlled by other proxy
	if eh.filtered(d) {
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := controller.GetControllerOf(d); controllerRef != nil {
		lb := eh.resolveControllerRef(d.Namespace, controllerRef)
		if lb == nil {
			return
		}
		log.Info("Deployment added", log.Fields{"d.name": d.Name, "lb.name": lb.Name, "ns": lb.Namespace})
		eh.helper.Enqueue(lb)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching LoadBalancer for deployment and sync
	// them to see if anyone wants to adopt it.
	lbs := eh.GetLoadBalancersForDeplyments(d)
	if len(lbs) == 0 {
		log.Debug("Can not get loadbalancer for orpha Deployment, ignore it", log.Fields{"d.name": d.Name, "ns": d.Namespace, "labels": d.Labels})
		return
	}
	log.Info("Orphan Deployment added", log.Fields{"d.name": d.Name})
	for _, lb := range lbs {
		eh.helper.Enqueue(lb)
	}

}

// OnUpdate ...
func (eh *EventHandlerForDeployment) OnUpdate(oldObj, curObj interface{}) {

	old := oldObj.(*extensions.Deployment)
	cur := curObj.(*extensions.Deployment)

	if old.ResourceVersion == cur.ResourceVersion {
		// Periodic resync will send update events for all known LoadBalancer.
		// Two different versions of the same LoadBalancer will always have different RVs.
		return
	}

	oldfiltered := eh.filtered(old)
	curfiltered := eh.filtered(cur)

	if oldfiltered && curfiltered {
		// both old and cur don't care it
		return
	}

	curControllerRef := controller.GetControllerOf(cur)
	oldControllerRef := controller.GetControllerOf(old)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)

	// do not sync deletion update
	if !oldfiltered && old.DeletionTimestamp != nil {
		// if controller changed and this proxy is interested in the old d, sync it
		if controllerRefChanged && oldControllerRef != nil {
			if lb := eh.resolveControllerRef(old.Namespace, oldControllerRef); lb != nil {
				// The ControllerRef was changed. Sync the old controller, if any.
				log.Info("Deployment updated, ControllerRef changed, sync for old controller", log.Fields{"name": old.Name, "ns": old.Namespace})
				eh.helper.Enqueue(lb)
			}
		}
	}

	// do not sync deletion update
	if !curfiltered && cur.DeletionTimestamp != nil {
		// If it has a ControllerRef and this proxy is interested in it, that's all that matters.
		if curControllerRef != nil {
			lb := eh.resolveControllerRef(cur.Namespace, curControllerRef)
			if lb == nil {
				return
			}
			log.Info("Deployment updated", log.Fields{"name": cur.Name})
			eh.helper.Enqueue(lb)
			return
		}

		// Otherwise, it's an orphan. Get a list of all matching deployment and sync
		// them to see if anyone wants to adopt it.
		labelChanged := !reflect.DeepEqual(cur.Labels, old.Labels)

		if labelChanged || controllerRefChanged {
			lbs := eh.GetLoadBalancersForDeplyments(cur)
			if len(lbs) == 0 {
				log.Debug("Can not get loadbalancer for orpha Deployment, ignore it", log.Fields{"d.name": cur.Name, "ns": cur.Namespace, "labels": cur.Labels})
				return
			}
			log.Info("Orphan Deployment updated", log.Fields{"d.name": cur.Name})
			for _, lb := range lbs {
				eh.helper.Enqueue(lb)
			}
		}
	}

	// nothing happend
}

// OnDelete ...
func (eh *EventHandlerForDeployment) OnDelete(obj interface{}) {
	d, ok := obj.(*extensions.Deployment)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*extensions.Deployment)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a LoadBalancer %#v", obj))
			return
		}
	}

	// filter Deployment that controller does not care
	if eh.filtered(d) {
		return
	}

	controllerRef := controller.GetControllerOf(d)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	lb := eh.resolveControllerRef(d.Namespace, controllerRef)
	if lb == nil {
		return
	}

	log.Info("Deployment deleted", log.Fields{"d.name": d.Name, "lb.name": lb.Name})
	eh.helper.Enqueue(lb)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the corrrect Kind.
func (eh *EventHandlerForDeployment) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *netv1alpha1.LoadBalancer {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	lb, err := eh.lbLister.LoadBalancers(namespace).Get(controllerRef.Name)
	if err != nil {
		return nil
	}
	if lb.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return nil
	}
	return lb
}

// GetLoadBalancersForDeplyments get a list of all matching LoadBalancer for deployment
func (eh *EventHandlerForDeployment) GetLoadBalancersForDeplyments(d *extensions.Deployment) []*netv1alpha1.LoadBalancer {
	lbs, err := eh.lbLister.GetLoadBalancersForControllee(d)
	if err != nil || len(lbs) == 0 {
		return nil
	}
	// Because all deployments's belonging to a loadbalancer should have a unique label key,
	// there should never be more than one loadbalancer returned by the above method.
	// If that happens we should probably dynamically repair the situation by ultimately
	// trying to clean up one of the controllers, for now we just return the older one
	if len(lbs) > 1 {
		log.Warn("user error! more than one loadbalancer is selecting deployment with labels", log.Fields{
			"namespace":        d.Namespace,
			"deploymentName":   d.Name,
			"labels":           d.Labels,
			"loadbalancerName": lbs[0].Name,
		})
	}

	return lbs
}

// EventHandlerForSyncStatusWithPod helps you create a event handler to sync status
// with pod event quickly, makes you focus on you own code
type EventHandlerForSyncStatusWithPod struct {
	helper *controllerutil.Helper

	lbLister  netlisters.LoadBalancerLister
	podLister corelisters.PodLister

	filtered filterPodFunc
}

// NewEventHandlerForSyncStatusWithPod ...
func NewEventHandlerForSyncStatusWithPod(
	lbLister netlisters.LoadBalancerLister,
	podLister corelisters.PodLister,
	helper *controllerutil.Helper,
	filterFunc filterPodFunc,
) *EventHandlerForSyncStatusWithPod {
	return &EventHandlerForSyncStatusWithPod{
		helper:    helper,
		lbLister:  lbLister,
		podLister: podLister,
		filtered:  filterFunc,
	}
}

// OnAdd ...
func (eh *EventHandlerForSyncStatusWithPod) OnAdd(obj interface{}) {
	pod := obj.(*v1.Pod)
	if eh.filtered(pod) {
		return
	}

	if pod.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		eh.OnDelete(pod)
		return
	}

	lb := eh.getLoadbalancerForPod(pod)
	if lb == nil {
		return
	}

	eh.helper.Enqueue(lb)

}

// OnUpdate ...
func (eh *EventHandlerForSyncStatusWithPod) OnUpdate(oldObj, curObj interface{}) {
	old := oldObj.(*v1.Pod)
	cur := curObj.(*v1.Pod)

	if old.ResourceVersion == cur.ResourceVersion {
		// Periodic resync will send update events for all known LoadBalancer.
		// Two different versions of the same LoadBalancer will always have different RVs.
		return
	}

	oldfiltered := eh.filtered(old)
	curfiltered := eh.filtered(cur)

	if oldfiltered && curfiltered {
		return
	}

	oldlb := eh.getLoadbalancerForPod(old)
	curlb := eh.getLoadbalancerForPod(cur)

	if !oldfiltered && oldlb != nil {
		if curlb == nil || oldlb.Name != curlb.Name || oldlb.Namespace != curlb.Namespace {
			// lb changed
			eh.helper.EnqueueAfter(oldlb, time.Second)
		}
	}

	if !curfiltered && curlb != nil {
		eh.helper.EnqueueAfter(curlb, time.Second)
	}

}

// OnDelete ...
func (eh *EventHandlerForSyncStatusWithPod) OnDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a LoadBalancer %#v", obj))
			return
		}
	}

	if eh.filtered(pod) {
		return
	}

	lb := eh.getLoadbalancerForPod(pod)
	if lb == nil {
		return
	}

	eh.helper.Enqueue(lb)
}

func (eh *EventHandlerForSyncStatusWithPod) getLoadbalancerForPod(pod *v1.Pod) *netv1alpha1.LoadBalancer {
	v, ok := pod.Labels[netv1alpha1.LabelKeyCreatedBy]
	if !ok {
		return nil
	}

	namespace, name, err := SplitNamespaceAndNameByDot(v)
	if err != nil {
		log.Error("error get namespace and name", log.Fields{"err": err})
		return nil
	}

	lb, err := eh.lbLister.LoadBalancers(namespace).Get(name)
	if err != nil {
		log.Error("can not find loadbalancer for pod", log.Fields{"lb.name": name, "lb.ns": namespace, "pod.name": pod.Name})
		return nil
	}
	return lb
}
