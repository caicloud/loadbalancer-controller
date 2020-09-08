package v1alpha2

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DatasetPlural = "datasets"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DatasetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dataset `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Dataset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DatasetSpec   `json:"spec"`
	Status            DatasetStatus `json:"status,omitempty"`
}

type DatasetSpec struct {
	DatasetType DatasetType        `json:"datasetType"`
	Size        *resource.Quantity `json:"size"`
}

type DatasetStatus struct {
	Phase DatasetPhase `json:"phase"`
	// size of the current dataset
	Size *resource.Quantity `json:"size"`
}

type DatasetType string

const (
	Git       DatasetType = "git"
	Glusterfs DatasetType = "glusterfs"
	CephFS    DatasetType = "cephfs"
)

type DatasetPhase string

const (
	DatasetActive      DatasetPhase = "Active"
	DatasetTerminating DatasetPhase = "Terminating"
	DatasetUnknown     DatasetPhase = "Unknown"
)
