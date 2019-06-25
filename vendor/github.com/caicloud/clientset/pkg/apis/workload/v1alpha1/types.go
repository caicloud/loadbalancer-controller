/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// +optional
	Status WorkloadStatus `json:"status,omitempty"`
}

// WorkloadSpec Specification of the desired behavior of the Workload
type WorkloadSpec struct {
	// RollbackRevision indicates the revision of the workload rollback to.
	// Will be cleared after rollback is done.
	RollbackRevision *int64 `json:"rollbackRevision,omitempty"`
	// This flag tells the controller to suspend deployment, statefulset and cronjob.
	// This flag can not work on job or daemonset.
	Suspend *bool `json:"suspend,omitempty"`
}

// WorkloadStatus is the most recently observed status of the Workload
type WorkloadStatus struct {
	// LastUpdateTime is the last update time of current workload
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Revision indicates the revision of the workload represented by WorkloadRevision.
	Revision int64 `json:"revision"`
	// Summaries contains all resources status of current workload. The key
	// should be a unique path.
	Summaries map[string]Summary `json:"summaries,omitempty"`
	// PodStatistics old pod and update pod statistics
	// summary the count of pod in the status which contain updating running failed...
	PodStatistics PodStatistics `json:"podStatistics,omitempty"`
	// Conditions is an array of current observed workload conditions.
	Conditions []WorkloadCondition `json:"conditions,omitempty"`
	// The number of old history to retain to allow rollback.
	// This is a pointer to distinguish between explicit zero and not specified.
	// Be consistent with the revisionHistoryLimit of watched object.
	// eg: daemonset statefulset deployment
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
	// Manifest is the serialized representation of the main resource.
	// eg: Daemonset Job CronJob Deployment Statefulset
	Manifest string `json:"manifest"`
	// SubManifest is the serialized representation of the sub resource, eg:Service.
	SubManifests []string `json:"subManifests,omitempty"`
}

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
	// WorkloadFailure means some parts of workload falled into wrong field. Some parts may work
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
	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// Summary describes the status of a part of a workload.
type Summary struct {
	// Path is the path which resources from
	Path string `json:"path,omitempty"`
	// Resources contains a kind-counter map.
	// A kind should be a unique name of a group resources.
	Resources map[string]ResourceCounter `json:"resources,omitempty"`
	// Reason for record the failed reason.
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about the failure.
	Message string `json:"message,omitempty"`
}

// PodStatistics counts all the pod in all phase, and divided them into old and updated
type PodStatistics struct {
	OldPods     PodStatusCounter `json:"oldPods,omitempty"`
	UpdatedPods PodStatusCounter `json:"updatedPods,omitempty"`
}

// PodStatusCounter is the pod status counter
type PodStatusCounter map[corev1.PodPhase]int32

// ResourceCounter is a status counter
type ResourceCounter map[ResourcePhase]int32

// ResourcePhase is a label for the condition of a resource at the current time.
type ResourcePhase string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkloadList describes an array of Workload instances
type WorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of workload histories
	Items []Workload `json:"items"`
}

//-------------------------------------------------------------------------------

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkloadRevision implements an immutable snapshot of state data. Clients
// are responsible for serializing and deserializing the objects that contain
// their internal state.
// Once a WorkloadRevision has been successfully created, it can not be updated.
// The API Server will fail validation of all requests that attempt to mutate
// the Data field. WorkloadRevision may, however, be deleted.
type WorkloadRevision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Revision indicates the revision of the state represented by Data.
	Revision int64 `json:"revision"`
	// Manifest is the serialized representation of the main resource.
	// eg: Daemonset Job CronJob Deployment Statefulset
	Manifest string `json:"manifest"`
	// SubManifest is the serialized representation of the sub resource, eg:Service.
	SubManifests []string `json:"subManifests,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkloadRevisionList describes an array of WorkloadRevision instances
type WorkloadRevisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of workload histories
	Items []WorkloadRevision `json:"items"`
}
