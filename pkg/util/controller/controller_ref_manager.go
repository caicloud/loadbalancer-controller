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
	"sync"

	log "github.com/zoumo/logdog"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	extensions "k8s.io/client-go/pkg/apis/extensions/v1beta1"

	"k8s.io/kubernetes/pkg/controller"
)

type baseControllerRefManager struct {
	controller metav1.Object
	selector   labels.Selector

	canAdoptErr  error
	canAdoptOnce sync.Once
	canAdoptFunc func() error
}

func (m *baseControllerRefManager) canAdopt() error {
	m.canAdoptOnce.Do(func() {
		if m.canAdoptFunc != nil {
			m.canAdoptErr = m.canAdoptFunc()
		}
	})
	return m.canAdoptErr
}

// claimObject tries to take ownership of an object for this controller.
//
// It will reconcile the following:
//   * Adopt orphans if the match function returns true.
//   * Release owned objects if the match function returns false.
//
// A non-nil error is returned if some form of reconciliation was attemped and
// failed. Usually, controllers should try again later in case reconciliation
// is still needed.
//
// If the error is nil, either the reconciliation succeeded, or no
// reconciliation was necessary. The returned boolean indicates whether you now
// own the object.
//
// No reconciliation will be attempted if the controller is being deleted.
func (m *baseControllerRefManager) claimObject(obj metav1.Object, match func(metav1.Object) bool, adopt, release func(metav1.Object) error) (bool, error) {
	controllerRef := controller.GetControllerOf(obj)
	if controllerRef != nil {
		if controllerRef.UID != m.controller.GetUID() {
			// Owned by someone else. Ignore.
			return false, nil
		}
		if match(obj) {
			// We already own it and the selector matches.
			// Return true (successfully claimed) before checking deletion timestamp.
			// We're still allowed to claim things we already own while being deleted
			// because doing so requires taking no actions.
			return true, nil
		}
		// Owned by us but selector doesn't match.
		// Try to release, unless we're being deleted.
		if m.controller.GetDeletionTimestamp() != nil {
			return false, nil
		}
		if err := release(obj); err != nil {
			// If the pod no longer exists, ignore the error.
			if errors.IsNotFound(err) {
				return false, nil
			}
			// Either someone else released it, or there was a transient error.
			// The controller should requeue and try again if it's still stale.
			return false, err
		}
		// Successfully released.
		return false, nil
	}

	// It's an orphan.
	if m.controller.GetDeletionTimestamp() != nil || !match(obj) {
		// Ignore if we're being deleted or selector doesn't match.
		return false, nil
	}
	if obj.GetDeletionTimestamp() != nil {
		// Ignore if the object is being deleted
		return false, nil
	}
	// Selector matches. Try to adopt.
	if err := adopt(obj); err != nil {
		// If the pod no longer exists, ignore the error.
		if errors.IsNotFound(err) {
			return false, nil
		}
		// Either someone else claimed it first, or there was a transient error.
		// The controller should requeue and try again if it's still orphaned.
		return false, err
	}
	// Successfully adopted.
	return true, nil
}

// DaemonSetControllerRefManager is used to manage controllerRef of DaemontSet.
// Three methods are defined on this object 1: Classify 2: AdoptDaemonSet and
// 3: ReleaseDaemonSet which are used to classify the DaemonSet into appropriate
// categories and accordingly adopt or release them. See comments on these functions
// for more details.
type DaemonSetControllerRefManager struct {
	baseControllerRefManager
	controllerKind schema.GroupVersionKind
	client         kubernetes.Interface
}

func NewDaemonSetControllerRefManager(
	client kubernetes.Interface,
	controller metav1.Object,
	selector labels.Selector,
	controllerKind schema.GroupVersionKind,
	canAdopt func() error,
) *DaemonSetControllerRefManager {
	return &DaemonSetControllerRefManager{
		baseControllerRefManager: baseControllerRefManager{
			controller:   controller,
			selector:     selector,
			canAdoptFunc: canAdopt,
		},
		controllerKind: controllerKind,
		client:         client,
	}
}

// Claim tries to take ownership of a list of DaemonSets.
//
// It will reconcile the following:
//   * Adopt orphans if the selector matches.
//   * Release owned objects if the selector no longer matches.
//
// A non-nil error is returned if some form of reconciliation was attemped and
// failed. Usually, controllers should try again later in case reconciliation
// is still needed.
//
// If the error is nil, either the reconciliation succeeded, or no
// reconciliation was necessary. The list of DaemonSets that you now own is
// returned.
func (m *DaemonSetControllerRefManager) Claim(sets []*extensions.DaemonSet) ([]*extensions.DaemonSet, error) {
	var claimed []*extensions.DaemonSet
	var errlist []error

	match := func(obj metav1.Object) bool {
		return m.selector.Matches(labels.Set(obj.GetLabels()))
	}
	adopt := func(obj metav1.Object) error {
		return m.Adopt(obj.(*extensions.DaemonSet))
	}
	release := func(obj metav1.Object) error {
		return m.Release(obj.(*extensions.DaemonSet))
	}

	for _, rs := range sets {
		ok, err := m.claimObject(rs, match, adopt, release)
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		if ok {
			claimed = append(claimed, rs)
		}
	}
	return claimed, utilerrors.NewAggregate(errlist)
}

// Adopt sends a patch to take control of the DaemonSet. It returns the error if
// the patching fails.
func (m *DaemonSetControllerRefManager) Adopt(ds *extensions.DaemonSet) error {
	if err := m.canAdopt(); err != nil {
		return fmt.Errorf("can't adopt DaemontSet %v/%v (%v): %v", ds.Namespace, ds.Name, ds.UID, err)
	}
	// Note that ValidateOwnerReferences() will reject this patch if another
	// OwnerReference exists with controller=true.
	addControllerPatch := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"apiVersion":"%s","kind":"%s","name":"%s","uid":"%s","controller":true,"blockOwnerDeletion":true}],"uid":"%s"}}`,
		m.controllerKind.GroupVersion(), m.controllerKind.Kind,
		m.controller.GetName(), m.controller.GetUID(), ds.UID)

	_, err := m.client.ExtensionsV1beta1().DaemonSets(ds.Namespace).Patch(ds.Name, types.StrategicMergePatchType, []byte(addControllerPatch))
	return err
}

// Release sends a patch to free the DaemonSet from the control of the LoadBalancer controller.
// It returns the error if the patching fails. 404 and 422 errors are ignored.
func (m *DaemonSetControllerRefManager) Release(ds *extensions.DaemonSet) error {
	log.Info("patching DaemonSet to remove its controllerRef", log.Fields{
		"ds.name":  ds.Name,
		"ds.ns":    ds.Namespace,
		"ctl.gv":   m.controllerKind.GroupVersion(),
		"ctl.kind": m.controllerKind.Kind,
		"ctl.name": m.controller.GetName(),
	})

	deleteOwnerRefPatch := fmt.Sprintf(`{"metadata":{"ownerReferences":[{"$patch":"delete","uid":"%s"}],"uid":"%s"}}`, m.controller.GetUID(), ds.UID)
	_, err := m.client.ExtensionsV1beta1().DaemonSets(ds.Namespace).Patch(ds.Name, types.StrategicMergePatchType, []byte(deleteOwnerRefPatch))
	if err != nil {
		if errors.IsNotFound(err) {
			// if DaemonSet no longer exists, ignore it
			return nil
		}
		if errors.IsInvalid(err) {
			// Invalid error will be returned in two cases: 1. the DaemonSet
			// has no owner reference, 2. the uid of the DaemonSet doesn't
			// match, which means the DaemonSet is deleted and then recreated.
			// In both cases, the error can be ignored.
			return nil
		}
	}
	return err
}

// DeploymentControllerRefManager is used to manage controllerRef of Deployment.
// Three methods are defined on this object 1: Classify 2: AdoptDeployment and
// 3: ReleaseDeployment which are used to classify the Deployment into appropriate
// categories and accordingly adopt or release them. See comments on these functions
// for more details.
type DeploymentControllerRefManager struct {
	baseControllerRefManager
	controllerKind schema.GroupVersionKind
	client         kubernetes.Interface
}

func NewDeploymentControllerRefManager(
	client kubernetes.Interface,
	controller metav1.Object,
	selector labels.Selector,
	controllerKind schema.GroupVersionKind,
	canAdopt func() error,
) *DeploymentControllerRefManager {
	return &DeploymentControllerRefManager{
		baseControllerRefManager: baseControllerRefManager{
			controller:   controller,
			selector:     selector,
			canAdoptFunc: canAdopt,
		},
		controllerKind: controllerKind,
		client:         client,
	}
}

// Claim tries to take ownership of a list of Deployments.
//
// It will reconcile the following:
//   * Adopt orphans if the selector matches.
//   * Release owned objects if the selector no longer matches.
//
// A non-nil error is returned if some form of reconciliation was attemped and
// failed. Usually, controllers should try again later in case reconciliation
// is still needed.
//
// If the error is nil, either the reconciliation succeeded, or no
// reconciliation was necessary. The list of Deployments that you now own is
// returned.
func (m *DeploymentControllerRefManager) Claim(sets []*extensions.Deployment) ([]*extensions.Deployment, error) {
	var claimed []*extensions.Deployment
	var errlist []error

	match := func(obj metav1.Object) bool {
		return m.selector.Matches(labels.Set(obj.GetLabels()))
	}
	adopt := func(obj metav1.Object) error {
		return m.Adopt(obj.(*extensions.Deployment))
	}
	release := func(obj metav1.Object) error {
		return m.Release(obj.(*extensions.Deployment))
	}

	for _, rs := range sets {
		ok, err := m.claimObject(rs, match, adopt, release)
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		if ok {
			claimed = append(claimed, rs)
		}
	}
	return claimed, utilerrors.NewAggregate(errlist)
}

// Adopt sends a patch to take control of the Deployment. It returns the error if
// the patching fails.
func (m *DeploymentControllerRefManager) Adopt(d *extensions.Deployment) error {
	if err := m.canAdopt(); err != nil {
		return fmt.Errorf("can't adopt Deployment %v/%v (%v): %v", d.Namespace, d.Name, d.UID, err)
	}
	// Note that ValidateOwnerReferences() will reject this patch if another
	// OwnerReference exists with controller=true.
	addControllerPath := fmt.Sprintf(
		`{"metadata":{"ownerReferences":[{"apiVersion":"%s","kind":"%s","name":"%s","uid":"%s","controller":true,"blockOwnerDeletion":true}],"uid":"%s"}}`,
		m.controllerKind.GroupVersion(), m.controllerKind.Kind,
		m.controller.GetName(), m.controller.GetUID(), d.UID)

	_, err := m.client.ExtensionsV1beta1().Deployments(d.Namespace).Patch(d.Name, types.StrategicMergePatchType, []byte(addControllerPath))
	return err
}

// Release sends a patch to free the Deployment from the control of the LoadBalancer controller.
// It returns the error if the patching fails. 404 and 422 errors are ignored.
func (m *DeploymentControllerRefManager) Release(d *extensions.Deployment) error {
	log.Info("patching Deployment to remove its controllerRef", log.Fields{
		"ds.name":  d.Name,
		"ds.ns":    d.Namespace,
		"ctl.gv":   m.controllerKind.GroupVersion(),
		"ctl.kind": m.controllerKind.Kind,
		"ctl.name": m.controller.GetName(),
	})

	deleteOwnerRefPatch := fmt.Sprintf(`{"metadata":{"ownerReferences":[{"$patch":"delete","uid":"%s"}],"uid":"%s"}}`, m.controller.GetUID(), d.UID)
	_, err := m.client.ExtensionsV1beta1().Deployments(d.Namespace).Patch(d.Name, types.StrategicMergePatchType, []byte(deleteOwnerRefPatch))
	if err != nil {
		if errors.IsNotFound(err) {
			// if DaemonSet no longer exists, ignore it
			return nil
		}
		if errors.IsInvalid(err) {
			// Invalid error will be returned in two cases: 1. the DaemonSet
			// has no owner reference, 2. the uid of the DaemonSet doesn't
			// match, which means the DaemonSet is deleted and then recreated.
			// In both cases, the error can be ignored.
			return nil
		}
	}
	return err
}
