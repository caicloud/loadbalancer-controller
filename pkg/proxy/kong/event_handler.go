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

package kong

import (
	"fmt"

	lblisters "github.com/caicloud/clientset/listers/loadbalance/v1alpha2"
	releaselisters "github.com/caicloud/clientset/listers/release/v1alpha1"
	lbapi "github.com/caicloud/clientset/pkg/apis/loadbalance/v1alpha2"
	releaseapi "github.com/caicloud/clientset/pkg/apis/release/v1alpha1"
	controllerutil "github.com/caicloud/clientset/util/controller"
	"github.com/caicloud/clientset/util/syncqueue"
	"github.com/caicloud/loadbalancer-controller/pkg/api"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	log "k8s.io/klog"
)

var _ cache.ResourceEventHandler = &EventHandlerForRelease{}

type filterReleaseFunc func(obj *releaseapi.Release) bool

// EventHandlerForRelease helps you create a event handler to handle with
// Releases event quickly, makes you focus on you own code
type EventHandlerForRelease struct {
	queue *syncqueue.SyncQueue

	lbLister lblisters.LoadBalancerLister
	releaseLister  releaselisters.ReleaseLister

	filtered filterReleaseFunc
}

// newEventHandlerForRelease ...
func newEventHandlerForRelease(
	lbLister lblisters.LoadBalancerLister,
	releaseLister releaselisters.ReleaseLister,
	queue *syncqueue.SyncQueue,
	filterFunc filterReleaseFunc,
) *EventHandlerForRelease {
	return &EventHandlerForRelease{
		queue:    queue,
		lbLister: lbLister,
		releaseLister:  releaseLister,
		filtered: filterFunc,
	}
}

// OnAdd ...
func (eh *EventHandlerForRelease) OnAdd(obj interface{}) {
	d := obj.(*releaseapi.Release)

	if d.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		eh.OnDelete(d)
		return
	}

	// filter Release that controller does not care
	// this Release maybe controlled by other proxy
	if eh.filtered(d) {
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := controllerutil.GetControllerOf(d); controllerRef != nil {
		lb := eh.resolveControllerRef(d.Namespace, controllerRef)
		if lb == nil {
			return
		}
		log.Infof("Release %v added, belongs to loadbalancer %v", d.Name, lb.Name)
		eh.queue.Enqueue(lb)
		return
	}

	// Otherwise, it's an orphan. Get a matching LoadBalancer for Release
	lb, err := eh.lbLister.GetLoadBalancerForControllee(d)
	if err != nil {
		log.Errorf("Can not get loadbalancer for orpha Release %v, ignore it, Release's labels %v", d.Name, d.Labels)
		return
	}
	log.Infof("Orphan Release %v added", d.Name)
	eh.queue.Enqueue(lb)

}

// OnUpdate ...
func (eh *EventHandlerForRelease) OnUpdate(oldObj, curObj interface{}) {

	old := oldObj.(*releaseapi.Release)
	cur := curObj.(*releaseapi.Release)

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
	if curControllerRef != nil {
		if lb := eh.resolveControllerRef(cur.Namespace, curControllerRef); lb != nil {
			// The ControllerRef was changed. Sync the old controller, if any.
			log.Infof("Release updated, ControllerRef changed, sync for old controller %v/%v", old.Namespace, old.Name)
			eh.queue.Enqueue(lb)
		}
	}
}

// OnDelete ...
func (eh *EventHandlerForRelease) OnDelete(obj interface{}) {
	d, ok := obj.(*releaseapi.Release)

	if !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Couldn't get object from tombstone %#v", obj))
			return
		}
		d, ok = tombstone.Obj.(*releaseapi.Release)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("Tombstone contained object that is not a LoadBalancer %#v", obj))
			return
		}
	}

	// filter Release that controller does not care
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

	log.Infof("Release %v deleted, belongs to loadbalancer %v", d.Name, lb.Name)
	eh.queue.Enqueue(lb)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the corrrect Kind.
func (eh *EventHandlerForRelease) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *lbapi.LoadBalancer {
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

// GetLoadBalancerForReleases get a list of all matching LoadBalancer for Release
func (eh *EventHandlerForRelease) GetLoadBalancerForReleases(d *releaseapi.Release) *lbapi.LoadBalancer {
	lb, err := eh.lbLister.GetLoadBalancerForControllee(d)
	if err != nil || lb == nil {
		log.Error("Error get loadbalancers for Releases")
		return nil
	}
	return lb
}