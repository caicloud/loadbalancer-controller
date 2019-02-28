package v1alpha1

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
	// current DatasetSpec is empty and will add extra fields when expand feature
}

type DatasetStatus struct {
	Phase DatasetPhase `json:"phase"`
	// size of the current dataset
	Size resource.Quantity `json:"size"`
	// the total number of commits
	CommitNums int64 `json:"commitNums"`
	// last commit time, sort by default.
	LastCommitTime *metav1.Time `json:"lastCommitTime,omitempty"`
}

type DatasetPhase string

const (
	DatasetActive      DatasetPhase = "Active"
	DatasetTerminating DatasetPhase = "Terminating"
	DatasetUnknown     DatasetPhase = "Unknown"
)
