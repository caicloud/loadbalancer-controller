package lb

import (
	"fmt"
	"reflect"

	"github.com/caicloud/loadbalancer-controller/api"
	caicloudinformers "github.com/caicloud/loadbalancer-controller/pkg/informers/caicloud/v1beta1"
	listers "github.com/caicloud/loadbalancer-controller/pkg/listers/caicloud/v1beta1"
	controllerutil "github.com/caicloud/loadbalancer-controller/pkg/util/controller"
	log "github.com/zoumo/logdog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	extensionsinformers "k8s.io/client-go/informers/extensions/v1beta1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/controller"
)

// controllerKind contains the schema.GroupVersionKind for this controller type.
var controllerKind = api.SchemeGroupVersion.WithKind(api.LoadBalancerKind)

type filterFunc func(obj *extensions.Deployment) bool

// EventHandlerForDeployment helps you create a event handler to handle with
// deployments event quickly, makes you focus on you own code
type EventHandlerForDeployment struct {
	helper *controllerutil.Helper

	lbLister       listers.LoadBalancerLister
	dLister        extensionslisters.DeploymentLister
	lbListerSynced cache.InformerSynced
	dListerSynced  cache.InformerSynced

	filtered filterFunc
}

// NewEventHandlerForDeployment ...
func NewEventHandlerForDeployment(lbInformer caicloudinformers.LoadBalancerInformer, dInformer extensionsinformers.DeploymentInformer, helper *controllerutil.Helper, filterFunc filterFunc) *EventHandlerForDeployment {
	c := &EventHandlerForDeployment{
		helper:         helper,
		lbLister:       lbInformer.Lister(),
		dLister:        dInformer.Lister(),
		lbListerSynced: lbInformer.Informer().HasSynced,
		dListerSynced:  dInformer.Informer().HasSynced,
		filtered:       filterFunc,
	}

	return c
}

// OnAdd ...
func (c *EventHandlerForDeployment) OnAdd(obj interface{}) {
	d := obj.(*extensions.Deployment)

	if d.DeletionTimestamp != nil {
		// On a restart of the controller manager, it's possible for an object to
		// show up in a state that is already pending deletion.
		c.OnDelete(d)
		return
	}

	// filter Deployment that controller does not care
	// this Deployment maybe controlled by other proxy
	if c.filtered(d) {
		return
	}

	// If it has a ControllerRef, that's all that matters.
	if controllerRef := controller.GetControllerOf(d); controllerRef != nil {
		lb := c.resolveControllerRef(d.Namespace, controllerRef)
		if lb == nil {
			return
		}
		log.Info("Deployment added", log.Fields{"d.name": d.Name, "lb.name": lb.Name, "ns": lb.Namespace})
		c.helper.Enqueue(lb)
		return
	}

	// Otherwise, it's an orphan. Get a list of all matching LoadBalancer for deployment and sync
	// them to see if anyone wants to adopt it.
	lbs := c.GetLoadBalancersForDeplyments(d)
	if len(lbs) == 0 {
		log.Debug("Can not get loadbalancer for orpha Deployment, ignore it", log.Fields{"d.name": d.Name, "ns": d.Namespace, "labels": d.Labels})
		return
	}
	log.Info("Orphan Deployment added", log.Fields{"d.name": d.Name})
	for _, lb := range lbs {
		c.helper.Enqueue(lb)
	}

}

// OnUpdate ...
func (c *EventHandlerForDeployment) OnUpdate(oldObj, curObj interface{}) {

	old := oldObj.(*extensions.Deployment)
	cur := curObj.(*extensions.Deployment)

	if old.ResourceVersion == cur.ResourceVersion {
		// Periodic resync will send update events for all known LoadBalancer.
		// Two different versions of the same LoadBalancer will always have different RVs.
		return
	}

	oldfiltered := c.filtered(old)
	curfiltered := c.filtered(cur)

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
			if lb := c.resolveControllerRef(old.Namespace, oldControllerRef); lb != nil {
				// The ControllerRef was changed. Sync the old controller, if any.
				log.Info("Deployment updated, ControllerRef changed, sync for old controller", log.Fields{"name": old.Name, "ns": old.Namespace})
				c.helper.Enqueue(lb)
			}
		}
	}

	// do not sync deletion update
	if !curfiltered && cur.DeletionTimestamp != nil {
		// If it has a ControllerRef and this proxy is interested in it, that's all that matters.
		if curControllerRef != nil {
			lb := c.resolveControllerRef(cur.Namespace, curControllerRef)
			if lb == nil {
				return
			}
			log.Info("Deployment updated", log.Fields{"name": cur.Name})
			c.helper.Enqueue(lb)
			return
		}

		// Otherwise, it's an orphan. Get a list of all matching deployment and sync
		// them to see if anyone wants to adopt it.
		labelChanged := !reflect.DeepEqual(cur.Labels, old.Labels)

		if labelChanged || controllerRefChanged {
			lbs := c.GetLoadBalancersForDeplyments(cur)
			if len(lbs) == 0 {
				log.Debug("Can not get loadbalancer for orpha Deployment, ignore it", log.Fields{"d.name": cur.Name, "ns": cur.Namespace, "labels": cur.Labels})
				return
			}
			log.Info("Orphan Deployment updated", log.Fields{"d.name": cur.Name})
			for _, lb := range lbs {
				c.helper.Enqueue(lb)
			}
		}
	}

	// nothing happend
}

// OnDelete ...
func (c *EventHandlerForDeployment) OnDelete(obj interface{}) {
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
	if c.filtered(d) {
		return
	}

	controllerRef := controller.GetControllerOf(d)
	if controllerRef == nil {
		// No controller should care about orphans being deleted.
		return
	}

	lb := c.resolveControllerRef(d.Namespace, controllerRef)
	if lb == nil {
		return
	}

	log.Info("Deployment deleted", log.Fields{"d.name": d.Name, "lb.name": lb.Name})
	c.helper.Enqueue(lb)
}

// resolveControllerRef returns the controller referenced by a ControllerRef,
// or nil if the ControllerRef could not be resolved to a matching controller
// of the corrrect Kind.
func (c *EventHandlerForDeployment) resolveControllerRef(namespace string, controllerRef *metav1.OwnerReference) *api.LoadBalancer {
	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != controllerKind.Kind {
		return nil
	}
	lb, err := c.lbLister.LoadBalancers(namespace).Get(controllerRef.Name)
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
func (c *EventHandlerForDeployment) GetLoadBalancersForDeplyments(d *extensions.Deployment) []*api.LoadBalancer {
	lbs, err := c.lbLister.GetLoadBalancersForControllee(d)
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
