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

package taints

import (
	"fmt"

	"k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

const (
	// MODIFIED ...
	MODIFIED = "modified"
	// TAINTED ...
	TAINTED = "tainted"
	// UNTAINTED ...
	UNTAINTED = "untainted"
)

// ValidateNoTaintOverwrites validates that when overwrite is false, to-be-updated taints don't exist in the node taint list (yet)
func ValidateNoTaintOverwrites(node *v1.Node, taints []v1.Taint) error {
	allErrs := []error{}
	oldTaints := node.Spec.Taints
	for _, taint := range taints {
		for _, oldTaint := range oldTaints {
			if taint.Key == oldTaint.Key {
				allErrs = append(allErrs, fmt.Errorf("Node '%s' already has a taint with key (%s) ", node.Name, taint.Key))
				break
			}
		}
	}
	return utilerrors.NewAggregate(allErrs)
}

// ReorganizeTaints returns the updated set of taints, taking into account old taints that were not updated,
// old taints that were updated, old taints that were deleted, and new taints.
func ReorganizeTaints(node *v1.Node, overwrite bool, taintsToAdd []v1.Taint, taintsToRemove []v1.Taint) (string, []v1.Taint, error) {
	newTaints := append([]v1.Taint{}, taintsToAdd...)
	oldTaints := node.Spec.Taints
	// add taints that already existing but not updated to newTaints
	added := AddTaints(oldTaints, &newTaints)
	allErrs, deleted := DeleteTaints(taintsToRemove, &newTaints)
	if (added && deleted) || overwrite {
		return MODIFIED, newTaints, utilerrors.NewAggregate(allErrs)
	} else if added {
		return TAINTED, newTaints, utilerrors.NewAggregate(allErrs)
	}
	return UNTAINTED, newTaints, utilerrors.NewAggregate(allErrs)
}

// DeleteTaints deletes the given taints from the node's taintlist.
func DeleteTaints(taintsToRemove []v1.Taint, newTaints *[]v1.Taint) ([]error, bool) {
	allErrs := []error{}
	var removed bool
	for _, taintToRemove := range taintsToRemove {
		removed = false
		if len(taintToRemove.Effect) > 0 {
			*newTaints, removed = DeleteTaint(*newTaints, &taintToRemove)
		} else {
			*newTaints, removed = DeleteTaintsByKey(*newTaints, taintToRemove.Key)
		}
		if !removed {
			allErrs = append(allErrs, fmt.Errorf("taint %q not found", taintToRemove.ToString()))
		}
	}
	return allErrs, removed
}

// AddTaints adds the newTaints list to existing ones and updates the newTaints List.
// TODO: This needs a rewrite to take only the new values instead of appended newTaints list to be consistent.
func AddTaints(oldTaints []v1.Taint, newTaints *[]v1.Taint) bool {
	for _, oldTaint := range oldTaints {
		existsInNew := false
		for _, taint := range *newTaints {
			if taint.MatchTaint(&oldTaint) {
				existsInNew = true
				break
			}
		}
		if !existsInNew {
			*newTaints = append(*newTaints, oldTaint)
		}
	}
	return len(oldTaints) != len(*newTaints)
}

// DeleteTaintsByKey removes all the taints that have the same key to given taintKey
func DeleteTaintsByKey(taints []v1.Taint, taintKey string) ([]v1.Taint, bool) {
	newTaints := []v1.Taint{}
	deleted := false
	for i := range taints {
		if taintKey == taints[i].Key {
			deleted = true
			continue
		}
		newTaints = append(newTaints, taints[i])
	}
	return newTaints, deleted
}

// DeleteTaint removes all the the taints that have the same key and effect to given taintToDelete.
func DeleteTaint(taints []v1.Taint, taintToDelete *v1.Taint) ([]v1.Taint, bool) {
	newTaints := []v1.Taint{}
	deleted := false
	for i := range taints {
		if taintToDelete.MatchTaint(&taints[i]) {
			deleted = true
			continue
		}
		newTaints = append(newTaints, taints[i])
	}
	return newTaints, deleted
}
