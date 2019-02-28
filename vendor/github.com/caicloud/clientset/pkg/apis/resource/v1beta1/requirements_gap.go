package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RequirementGap is the requirement gap object of the specified Deployment/DaemonSet/StatefulSet.
type RequirementGap struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RequirementGapSpec   `json:"spec,omitempty"`
	Status RequirementGapStatus `json:"status,omitempty"`
}

// RequirementGapSpec defines the spec information of RequirementGap
type RequirementGapSpec struct {
	// Reference refer the alerting Deployment/DaemonSet/StatefulSet
	Reference RequirementGapReference `json:"reference"`
}

// RequirementGapReference refer to Deployment/DaemonSet/StatefulSet
type RequirementGapReference struct {
	// Deployment/DaemonSet/StatefulSet
	Kind string `json:"kind"`
	// Namespace of object
	Namespace string `json:"namespace"`
	// Name of object
	Name string `json:"name"`
	// Tenant of the partition
	Tenant string `json:"tenant"`
}

// RequirementGapStatus represents the current information about requirement gap.
type RequirementGapStatus struct {
	// Summary is requirement info summary of referred object
	Summary RequirementGapSummary `json:"summary"`
	// LastUpdateTime record last time that this object update
	LastUpdateTime metav1.Time `json:"lastUpdateTime"`
	// Reason records the gap reason, partition quota is short or no node to schedule
	Reason string `json:"reason"`
	// Message records other message
	Message string `json:"message"`
}

type RequirementGapSummary struct {
	// Single save the whole pod resource requirement, summary all normal containers
	// for initContainers, if single initContainer is larger than total sum, total sum will add to max
	Single corev1.ResourceRequirements `json:"single"`
	// Replica is the num of needed pods
	Replica int `json:"replica"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RequirementGapList is a collection of requirement gaps.
type RequirementGapList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of RequirementGaps
	Items []RequirementGap `json:"items"`
}
