/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1alpha1

import (
	releaseapi "github.com/caicloud/clientset/pkg/apis/release/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Application describes a collection of services and their startup sequence
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Application
	// +optional
	Spec ApplicationSpec `json:"spec,omitempty"`

	// Most recently observed status of the Application
	// +optional
	Status ApplicationStatus `json:"status,omitempty"`
}

// ApplicationSpec ...
type ApplicationSpec struct {
	// Description is the description of current application
	// +optional
	Description string `json:"description,omitempty"`

	// Template is an archived template data
	Template []byte `json:"template"`

	// Graph contains all workloads and their startup sequence
	Graph WorkloadGraph `json:"graph"`
}

// WorkloadGraph describes the workflows startup sequence DAG
type WorkloadGraph struct {
	// This flag tells the controller to suspend all the workloads
	// Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// List of vertexes belonging to the graph
	// +optional
	Vertexes []Vertex `json:"vertexes,omitempty"`

	// List of directed edges belonging to the graph
	// The edges describe the startup sequence. The following kinds of edge
	// is not allowed:
	// - self edge
	// - deplicated edge
	// - vertexes in edge are not present in node list
	// - cycle
	// +optional
	Edges []Edge `json:"edges,omitempty"`
}

// Vertex describes a workload in the graph
type Vertex struct {
	VertexMeta `json:"metadata,omitempty"`
	Spec       VertexSpec `json:"spec,omitempty"`
}

// VertexMeta describes the metadata of a workload
type VertexMeta struct {
	// Specification of the index of node in the graph
	Index int64 `json:"index"`

	// Specification of the name of Release
	// +optional
	Name string `json:"name,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// The Annotations will be present in workload's Annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// The Labels will be present in workload's Labels
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Memos contain extra infromation of this workload, such as coordinates.
	// The difference beween Memos and Annotations is that the informations in Memos
	// are only stored here, will not be present in workload
	// +optional
	Memos map[string]string `json:"memos,omitempty"`
}

// VertexSpec ...
type VertexSpec struct {
	// This flag tells the controller to suspend the workload
	// Defaults to false.
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// Description is the description of current workload
	// +optional
	Description string `json:"description,omitempty"`

	// Config is the config for parsing template
	Config string `json:"config,omitempty"`

	// AdoptedIfOrphan means if the vertex is orphan,
	// it could be adopted by the application.
	// Defaults to false.
	// +optional
	AdoptedIfOrphan *bool `json:"adoptedIfOrphan,omitempty"`
}

// Edge is edge in workflows startup sequence DAG
type Edge struct {
	// source vertex
	From int64 `json:"from,omitempty"`
	// destination vertex
	To int64 `json:"to,omitempty"`
}

// ApplicationStatus is the most recently observed status of the Application
type ApplicationStatus struct {
	// Conditions is an array of current observed application conditions.
	Conditions []AppicationCondition `json:"conditions,omitempty"`
	// LastUpdateTime is the last update time of current application
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// workloads status
	Workloads []WorkloadStatus `json:"workloadStatuses,omitempty"`
}

// WorkloadStatus describes the current
type WorkloadStatus struct {
	Index   int64         `json:"index,omitempty"`
	Name    string        `json:"name,omitempty"`
	Phase   ResourcePhase `json:"phase,omitempty"`
	Reason  string        `json:"reason,omitempty"`
	Message string        `json:"message,omitempty"`
}

// ApplicationConditionType ...
type ApplicationConditionType string

// AppicationCondition describes the conditions of a application
type AppicationCondition struct {
	// Type of application condition.
	Type ApplicationConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// Last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationList describes an array of application instances
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

// ResourcePhase is a label for the condition of a resource at the current time.
type ResourcePhase = releaseapi.ResourcePhase

const (
	// ResourceNotCreated means that the resources are not created
	ResourceNotCreated ResourcePhase = "NotCreated"
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
	// - CronJob: there are unfinished Jobs belongs to the CronJob
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

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationDraft describes a temporary application draft
type ApplicationDraft struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Data contains the application draft data
	Data ApplicationDraftData `json:"data,omitempty"`
}

// ApplicationDraftData contains the application draft data
type ApplicationDraftData struct {
	// Graph contains all workloads and their startup sequence
	Graph WorkloadGraph `json:"graph"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationDraftList describes an array of application instances
type ApplicationDraftList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationDraft `json:"items"`
}
