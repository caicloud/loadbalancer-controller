/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigClaimStatusType is sync status of config claim
type ConfigClaimStatusType string

const (
	// Unknown means that config is sync not yet
	Unknown ConfigClaimStatusType = "Unknown"
	// Success means taht config is sync success
	Success ConfigClaimStatusType = "Success"
	// Failure  means taht config is sync failuer
	Failure ConfigClaimStatusType = "Failure"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigClaim describes a config sync status
type ConfigClaim struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Most recently observed status of the Release
	// +optional
	Status ConfigClaimStatus `json:"status,omitempty"`
}

// ConfigClaimStatus describes the status of a ConfigClaim
type ConfigClaimStatus struct {
	// Status is sync status of Config
	Status ConfigClaimStatusType `json:"status"`
	// Reason describes success or Failure of status
	Reason string `json:"reason,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigClaimList describes an array of ConfigClaim instances
type ConfigClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of ConfigClaim
	Items []ConfigClaim `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigReference describes the config reference list.
type ConfigReference struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the Release
	// +optional
	Status ConfigReferenceStatus `json:"status,omitempty"`
}

// ConfigReferenceStatus describes the config reference list.
type ConfigReferenceStatus struct {
	Refs []*Reference `json:"refs,omitempty"`
}

// Reference describes the config reference.
type Reference struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Kind        string            `json:"kind"`
	APIGroup    string            `json:"apiGroup"`
	APIVersion  string            `json:"apiVersion"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Keys        []string          `json:"keys,omitempty"`
	Children    []*Reference      `json:"children,omitempty"`
	ExtraInfo   `json:",inline"`
}

// ExtraInfo describes the extra infomation of config reference.
type ExtraInfo struct {
	ReleaseExtraInfo `json:",inline"`
	IngressExtraInfo `json:",inline"`
}

// ExtraInfo describes the release extra infomation of config reference.
type ReleaseExtraInfo struct {
	Alias        string        `json:"alias,omitempty"`
	ReleaseKind  string        `json:"releaseKind,omitempty"`
	ControlledBy *ControlledBy `json:"controlledBy,omitempty"`
}

type ControlledBy struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

// IngressExtraInfo describes the ingress extra infomation of config reference.
type IngressExtraInfo struct {
	LoadBalancer string `json:"loadBalancer,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigReferenceList describes an array of ConfigReference instances.
type ConfigReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of ConfigClaim
	Items []ConfigReference `json:"items"`
}
