/*
Copyright 2019 caicloud authors. All rights reserved.
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Workload describes a collection of manifest
type Workload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Workload
	// +optional
	Spec WorkloadSpec `json:"spec"`
	// Most recently observed status of the Workload
	Status WorkloadStatus `json:"status,omitempty"`
}

// WorkloadSpec Specification of the desired behavior of the Workload
type WorkloadSpec struct{}

// WorkloadStatus is the most recently observed status of the Workload
type WorkloadStatus struct {
	// LastUpdateTime is the last update time of current workload
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Revision indicates the revision of the workload represented by WorkloadRevision.
	Revision int64 `json:"revision"`
	// ResourceStatuses contains all resources status of current workload.
	ResourceStatuses []ResourceStatus `json:"resourceStatuses,omitempty"`
	// PodStatistics old pod and update pod statistics
	// summary the count of pod in the status which contain updating running failed...
	PodStatistics PodStatistics `json:"podStatistics,omitempty"`
	// Conditions is an array of current observed workload conditions.
	Conditions []WorkloadCondition `json:"conditions,omitempty"`
}
type ResourceStatus struct {
	// gvk info
	schema.GroupVersionKind `json:",inline,omitempty"`
	// Name name of sub resource
	Name string `json:"name,omitempty"`
	// Namespace namespace of sub resource
	Namespace string `json:"namespace,omitempty"`
	// Phase of resource
	Phase ResourcePhase `json:"phase,omitempty"`
	// Reason for record the failed reason.
	Reason string `json:"reason,omitempty"`
	// Message human readable message indicating details about the failure.
	Message string `json:"message,omitempty"`
	// IsPrimary the major resource of the workload, e.g.: deployment statefulSet
	IsPrimary bool `json:"isPrimary,omitempty"`
}

// ResourcePhase is a label for the condition of a resource at the current time.
type ResourcePhase string

const (
	// ResourceSuspended means that:
	// - For a long running resource: it desire 0 replicas and there is really
	//   no replica belongs to it now.
	// - For CronJob: it means that cronjob suspend subsequent executions
	ResourceSuspended ResourcePhase = "Suspended"
	// ResourcePending only for CronJob, means the CronJob has no Job histories
	ResourcePending ResourcePhase = "Pending"
	// ResourceProgressing means that:
	// - Deployment, StatefulSet, DaemonSet: all pods are updated, and the replicas
	//   are in sacling
	// - Job: the succeeded pod number doesn't meet the desired completion number
	// - CronJob: there are unfinished Jobs beloings to the CronJob
	// - PVC: the pvc is not bound
	ResourceProgressing ResourcePhase = "Progressing"
	// ResourceUpdating means:
	// - Deployment, StatefulSet, DaemonSet: the system is working to deal with the
	//   resource's updating request there are some old pods mixed with the
	//   updated pods, and no pods are in Abnormal
	ResourceUpdating ResourcePhase = "Updating"
	// ResourceRunning means:
	// - Deployment, StatefulSet, DaemonSet: all pods are updated, and thay are running
	// - PVC: bound
	// - Service, ConfigMap .e.g
	ResourceRunning ResourcePhase = "Running"
	// ResourceSucceeded means:
	// - Job: the succeeded pod number meets the desired completion number
	// - CronJob: the latest Job in history is Succeeded
	ResourceSucceeded ResourcePhase = "Succeeded"
	// ResourceFailed means:
	// - Deployment, StatefulSet, DaemonSet: one of the pods is in Abnormal
	// - Job: the job doesn't finished in active deadline
	// - CronJob: the latest Job in history is Succeeded
	// - PVC: Lost
	ResourceFailed ResourcePhase = "Failed"
)

// WorkloadConditionType workload type
type WorkloadConditionType string

const (
	// WorkloadAvailable means the resources of workload are available.
	WorkloadAvailable WorkloadConditionType = "Available"
	// WorkloadProgressing means workload is playing a mutation. It occurs when create/update/rollback workload.
	// If some bad thing was trigger, workload transfers to WorkloadFailure.
	WorkloadProgressing WorkloadConditionType = "Progressing"
	// WorkloadFailure means some parts of workload have fallen into wrong field. Some parts may work
	// as usual, but the workload can't provide complete service.
	WorkloadFailure WorkloadConditionType = "Failure"
)

// WorkloadCondition describes the conditions of a workload
type WorkloadCondition struct {
	// Type of workload condition.
	Type WorkloadConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Message Human readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// PodStatistics counts all the pod in all phase, and divided them into old and updated
type PodStatistics struct {
	Old     PodStatusCounter `json:"old,omitempty"`
	Updated PodStatusCounter `json:"updated,omitempty"`
}

// PodStatusCounter is the pod status counter
type PodStatusCounter map[corev1.PodPhase]int32

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkloadList describes an array of Workload instances
type WorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of workload histories
	Items []Workload `json:"items"`
}
