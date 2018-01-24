/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

// ReleaseRollbackConfig describes the rollback config of a release
type ReleaseRollbackConfig struct {
	// The version to rollback to. If set to 0, rollbck to the last version.
	Version int32 `json:"version,omitempty"`
}

// ReleaseSpec describes the basic info of a release
type ReleaseSpec struct {
	// Description is the description of current release
	Description string `json:"description,omitempty"`
	// Template is an archived template data
	Template []byte `json:"template,omitempty"`
	// Config is the config for parsing template
	Config string `json:"config,omitempty"`
	// The config this release is rolling back to. Will be cleared after rollback is done.
	RollbackTo *ReleaseRollbackConfig `json:"rollbackTo,omitempty"`
}

type ReleaseConditionType string

const (
	// ReleaseAvailable means the resources of release are available and can render service.
	ReleaseAvailable ReleaseConditionType = "Available"
	// ReleaseProgressing means release is playing a mutation. It occurs when create/update/rollback
	// release. If some bad thing was trigger, release transfers to ReleaseFailure.
	ReleaseProgressing ReleaseConditionType = "Progressing"
	// ReleaseFailure means some parts of release falled into wrong field. Some parts may work
	// as usual, but the release can't provide complete service.
	ReleaseFailure ReleaseConditionType = "Failure"
)

// ReleaseCondition describes the conditions of a release
type ReleaseCondition struct {
	// Type of release condition.
	Type ReleaseConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status apiv1.ConditionStatus `json:"status"`
	// Last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// ResourceCounter is a status counter
type ResourceCounter struct {
	// Available is the count of running target
	Available int32 `json:"available"`
	// Progressing is the count of mutating target
	Progressing int32 `json:"progressing"`
	// Failure is the count of wrong target
	Failure int32 `json:"failure"`
}

// ReleaseDetailStatus describes the status of a part of a release.
type ReleaseDetailStatus struct {
	// Path is the path which resources from
	Path string `json:"path,omitempty"`
	// Resources contains a kind-counter map.
	// A kind should be a unique name of a group resources.
	Resources map[string]ResourceCounter `json:"resources,omitempty"`
}

// ReleaseStatus describes the status of a release
type ReleaseStatus struct {
	// Version is the version of current release
	Version int32 `json:"version,omitempty"`
	// Manifest is the generated kubernetes resources from template
	Manifest string `json:"manifest,omitempty"`
	// LastUpdateTime is the last update time of current release
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Details contains all resources status of current release. The key
	// should be a unique path.
	Details map[string]ReleaseDetailStatus `json:"details,omitempty"`
	// Conditions is an array of current observed release conditions.
	Conditions []ReleaseCondition `json:"conditions,omitempty"`
}

// +genclient=true
// +genclientstatus=false

// Release describes a release wich chart and values
type Release struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Release
	// +optional
	Spec ReleaseSpec `json:"spec,omitempty"`

	// Most recently observed status of the Release
	// +optional
	Status ReleaseStatus `json:"status,omitempty"`
}

// ReleaseList describes an array of Release instances
type ReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of releases
	Items []Release `json:"items"`
}

// ReleaseHistorySpec describes the history info of a release
type ReleaseHistorySpec struct {
	// Description is the description of current history
	Description string `json:"description,omitempty"`
	// Version is the version of a history
	Version int32 `json:"version,omitempty"`
	// Template is an archived template data
	Template []byte `json:"template,omitempty"`
	// Config is the config for parsing template
	Config string `json:"config,omitempty"`
	// Manifest is the generated kubernetes resources from template
	Manifest string `json:"manifest,omitempty"`
}

// +genclient=true
// +genclientstatus=false

// ReleaseHistory describes a history of a release version
type ReleaseHistory struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the ReleaseHistory
	// +optional
	Spec ReleaseHistorySpec `json:"spec,omitempty"`
}

// ReleaseHistoryList describes an array of ReleaseHistory instances
type ReleaseHistoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of release histories
	Items []ReleaseHistory `json:"items"`
}

// -----------------------------------------------------------------

// +genclient=true
// +genclientstatus=false

// CanaryRelease describes a cannary release
// which providers cannary release for applications.
type CanaryRelease struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the CanaryReleaseSpec
	Spec CanaryReleaseSpec `json:"spec,omitempty"`
	// Most recently observed status of the CanaryReleasepec
	Status CanaryReleaseStatus `json:"status,omitempty"`
}

// CanaryReleaseSpec describes the basic info of a canary release
type CanaryReleaseSpec struct {
	// Release is the name of release TPR associated with this CanaryRelease
	Release string `json:"release"`
	// Version is the version  of release TPR associated with this CanaryRelease
	Version int32 `json:"version"`
	// Path is the path of sub app which needs Canary release
	Path string `json:"path"`
	// Config is the sub config for parsing template, aka Value
	Config string `json:"config"`
	// Service is an array of services in current release node
	Service []CanaryService `json:"services,omitempty"`
	// Resources specify cpu/memory usage of current canary release
	Resources apiv1.ResourceRequirements `json:"resources,omitempty"`
	// Transition is the next phase this CanaryRelease needs to transformed into
	Transition CanaryTrasition `json:"transition,omitempty"`
}

// CanaryTrasition specify the next phase this canary release want to be
type CanaryTrasition string

const (
	// CanaryTrasitionNone is the default value of  trasition
	CanaryTrasitionNone CanaryTrasition = ""
	// CanaryTrasitionAdopted means that this canary release should be adopted
	CanaryTrasitionAdopted CanaryTrasition = "Adopted"
	// CanaryTrasitionDeprecated means that this canary release should be deprecated
	CanaryTrasitionDeprecated CanaryTrasition = "Deprecated"
)

// CanaryService describes a config of a service from release node
type CanaryService struct {
	// Service is the name of the service
	Service string `json:"service,omitempty"`
	// Ports contains an array of port configs
	Ports []CanaryPort `json:"ports,omitempty"`
}

// CanaryPort defines protocol and usable config for a serviec port
type CanaryPort struct {
	// Port is the port number
	Port int32 `json:"port,omitempty"`
	// Protocol is the protocol used by the port
	Protocol Protocol `json:"protocol,omitempty"`
	// Config is the port proxy option
	Config CanaryConfig `json:"config,omitempty"`
}

// Protocol is the network type for ports
type Protocol string

const (
	ProtocolHTTP  Protocol = "HTTP"
	ProtocolHTTPS Protocol = "HTTPS"
	ProtocolTCP   Protocol = "TCP"
	ProtocolUDP   Protocol = "UDP"
)

// CanaryConfig describes a proxy config for a service port
type CanaryConfig struct {
	// Weight is the only proxy config now. The value of weight should be [1,100].
	Weight *int32 `json:"weight,omitempty"`
}

// CanaryReleaseStatus describes the current status of a canary release
type CanaryReleaseStatus struct {
	// Phase is the current phase of canary release.
	// It will be set after the transition in spec successfully.
	Phase CanaryTrasition `json:"phase,omitempty"`
	// Manifest is the generated kubernetes resources from template
	Manifest string `json:"manifest,omitempty"`
	// LastUpdateTime is the last update time of current release
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Conditions is an array of current observed release conditions.
	Conditions []CanaryReleaseCondition `json:"conditions,omitempty"`
	// canary release proxy status
	Proxy CanaryReleaseProxyStatus `json:"proxyStatus,omitempty"`
}

// CanaryReleaseConditionType describes the type of condition
type CanaryReleaseConditionType string

const (
	// CanaryReleaseAvailable means the resources of release are available and can render service.
	CanaryReleaseAvailable CanaryReleaseConditionType = "Available"
	// CanaryReleaseProgressing means release is playing a mutation. It occurs when create/update
	// a canary release. If some bad thing occurs, canary release transfers to ReleaseFailure.
	CanaryReleaseProgressing CanaryReleaseConditionType = "Progressing"
	// CanaryReleaseFailure means some parts of cananry release falled into wrong field. Some parts may work
	// as usual, but the canary release can't provide complete service.
	CanaryReleaseFailure CanaryReleaseConditionType = "Failure"
	// CanaryReleaseArchived means this canary release has been archived
	CanaryReleaseArchived CanaryReleaseConditionType = "Archived"
)

// CanaryReleaseCondition describes a condition of the canary release status
type CanaryReleaseCondition struct {
	// Type of release condition.
	Type CanaryReleaseConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status apiv1.ConditionStatus `json:"status"`
	// Last time the condition transit from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	Message string `json:"message,omitempty"`
}

// CanaryReleaseProxyStatus describes the current status of canary release proxy replicas
type CanaryReleaseProxyStatus struct {
	Deployment    string      `json:"deployment"`
	Replicas      int32       `json:"replicas"`
	TotalReplicas int32       `json:"totalReplicas"`
	ReadyReplicas int32       `json:"readyReplicas"`
	PodStatuses   []PodStatus `json:"podStatuses"`
}

// PodStatus represents the current status of a pod
type PodStatus struct {
	Name            string         `json:"name"`
	Ready           bool           `json:"ready"`
	RestartCount    int32          `json:"restartCount"`
	ReadyContainers int32          `json:"readyContainers"`
	TotalContainers int32          `json:"totalContainers"`
	NodeName        string         `json:"nodeName"`
	Phase           apiv1.PodPhase `json:"phase"`
	Reason          string         `json:"reason"`
	Message         string         `json:"message,omitempty"`
}

// CanaryReleaseList describes an array of canary release instances
type CanaryReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []CanaryRelease `json:"items"`
}
