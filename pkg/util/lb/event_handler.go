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

	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	controllerutil "github.com/caicloud/clientset/util/controller"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/api"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	log "k8s.io/klog"
)

var _ cache.ResourceEventHandler = &EventHandlerForDeployment{}
var _ cache.ResourceEventHandler = &EventHandlerForSyncStatusWithPod{}

type filterDeploymentFunc func(obj *appsv1.Deployment) bool
type filterPodFunc func(obj *v1.Pod) bool

// EventHandlerForDeployment helps you create a event handler to handle with
// deployments event quickly, makes you focus on you own code
type EventHandlerForDeployment struct {
	queue *syncqueue.SyncQueue

	lbLister lblisters.LoadBalancerLister
	dLister  appslisters.DeploymentLister

	filtered filterDeploymentFunc
}

// NewEventHandlerForDeployment ...
func NewEventHandlerForDeployment(
	lbLister lblisters.LoadBalancerLister,
	dLister appslisters.DeploymentLister,
	queue *syncqueue.SyncQueue,
	filterFunc filterDeploymentFunc,
) *EventHandlerForDeployment {
	return &EventHandlerForDeployment{
		queue:    queue,
		lbLister: lbLister,
		dLister:  dLister,
		filtered: filterFunc,
	}
}

// OnAdd ...
func (eh *EventHandlerForDeployment) OnAdd(obj interface{}) {
	d := obj.(*appsv1.Deployment)

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
	if controllerRef := controllerutil.GetControllerOf(d); controllerRef != nil {
		lb := eh.resolveControllerRef(d.Namespace, controllerRef)
		if lb == nil {
			return
		}
		log.Infof("Deployment %v added, belongs to loadbalancer %v", d.Name, lb.Name)
		eh.queue.Enqueue(lb)
		return
	}

	// Otherwise, it's an orphan. Get a matching LoadBalancer for deployment
	lb, err := eh.lbLister.GetLoadBalancerForControllee(d)
	if err != nil {
		log.Errorf("Can not get loadbalancer for orpha Deployment %v, ignore it, deployment's labels %v", d.Name, d.Labels)
		return
	}
	log.Infof("Orphan Deployment %v added", d.Name)
	eh.queue.Enqueue(lb)

}

// OnUpdate ...
func (eh *EventHandlerForDeployment) OnUpdate(oldObj, curObj interface{}) {

	old := oldObj.(*appsv1.Deployment)
	cur := curObj.(*appsv1.Deployment)

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

	curControllerRef := controllerutil.GetControllerOf(cur)
	oldControllerRef := controllerutil.GetControllerOf(old)
	controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)

	// do not sync deletion update
	if !oldfiltered && old.DeletionTimestamp != nil {
		// if controller changed and this proxy is interested in the old d, sync it
		if controllerRefChanged && oldControllerRef != nil {
			if lb := eh.resolveControllerRef(old.Namespace, oldControllerRef); lb != nil {
				// The ControllerRef was changed. Sync the old controller, if any.
				log.Infof("Deployment updated, ControllerRef changed, sync for old controller %v/%v", old.Namespace, old.Name)
				eh.queue.Enqueue(lb)
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
			log.Infof("Deployment %v updated", cur.Name)
			eh.queue.Enqueue(lb)
			return
		}

		// Otherwise, it's an orphan. Get a list of all matching deployment and sync
		// them to see if anyone wants to adopt it.
		labelChanged := !reflect.DeepEqual(cur.Labels, old.Labels)

		if labelChanged || controllerRefChanged {
			lb, err := eh.lbLister.GetLoadBalancerForControllee(cur)
			if err != nil {
				log.Errorf("Can not get loadbalancer for orpha Deployment %v/%v, ignore it, deployment's labels %v", cur.Namespace, cur.Name, cur.Labels)
				return
			}
			log.Infof("Orphan Deployment %v updated", cur.Name)
			eh.queue.Enqueue(lb)
		}
	}

	// nothing happened
}

// OnDelete ...
func (eh *EventHandlerForDeployment) OnDelete(obj interface{}) {
	d, ok := obj.(*appsv1.Deployment)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*appsv1.Deployment)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a LoadBalancer %#v", obj))
			return
		}
	}

	// filter Deployment that controller does not care
	if eh.filtered(d) {
		return
	}

	controllerRef := controllerutil.GetControllerOf(d)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	lb := eh.resolveControllerRef(d.Namespace, controllerRef)
	if lb == nil {
		return
	}

	log.Infof("Deployment %v deleted, belongs to loadbalancer %v", d.Name, lb.Name)
	eh.queue.Enqueue(lb)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the corrrect Kind.
func (eh *EventHandlerForDeployment) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *lbapi.LoadBalancer {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != api.ControllerKind.Kind {
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

// GetLoadBalancerForDeployments get a list of all matching LoadBalancer for deployment
func (eh *EventHandlerForDeployment) GetLoadBalancerForDeployments(d *appsv1.Deployment) *lbapi.LoadBalancer {
	lb, err := eh.lbLister.GetLoadBalancerForControllee(d)
	if err != nil || lb == nil {
		log.Error("Error get loadbalancers for deployments")
		return nil
	}
	return lb
}

// EventHandlerForSyncStatusWithPod helps you create a event handler to sync status
// with pod event quickly, makes you focus on you own code
type EventHandlerForSyncStatusWithPod struct {
	queue     *syncqueue.SyncQueue
	lbLister  lblisters.LoadBalancerLister
	podLister corelisters.PodLister
	filtered  filterPodFunc
}

// NewEventHandlerForSyncStatusWithPod ...
func NewEventHandlerForSyncStatusWithPod(
	lbLister lblisters.LoadBalancerLister,
	podLister corelisters.PodLister,
	queue *syncqueue.SyncQueue,
	filterFunc filterPodFunc,
) *EventHandlerForSyncStatusWithPod {
	return &EventHandlerForSyncStatusWithPod{
		queue:     queue,
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

	eh.queue.Enqueue(lb)

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
			eh.queue.EnqueueAfter(oldlb, time.Second)
		}
	}

	if !curfiltered && curlb != nil {
		eh.queue.EnqueueAfter(curlb, time.Second)
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

	eh.queue.Enqueue(lb)
}

func (eh *EventHandlerForSyncStatusWithPod) getLoadbalancerForPod(pod *v1.Pod) *lbapi.LoadBalancer {
	v, ok := pod.Labels[lbapi.LabelKeyCreatedBy]
	if !ok {
		return nil
	}

	namespace, name, err := SplitNamespaceAndNameByDot(v)
	if err != nil {
		log.Errorf("error get namespace and name: %v", err)
		return nil
	}

	lb, err := eh.lbLister.LoadBalancers(namespace).Get(name)
	if errors.IsNotFound(err) {
		// deleted
		return nil
	}
	if err != nil {
		log.Errorf("can not find loadbalancer %v for pod %v: %v", name, pod.Name, err)
		return nil
	}
	return lb
}
