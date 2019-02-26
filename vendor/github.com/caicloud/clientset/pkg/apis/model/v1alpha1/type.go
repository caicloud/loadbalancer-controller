package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ModelPlural = "models"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Model `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Model struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ModelSpec   `json:"spec"`
	Status            ModelStatus `json:"status,omitempty"`
}

type FrameworkType string

const (
	Tensorflow FrameworkType = "tensorflow"
	Pytorch                  = "pytorch"
	MXNet                    = "mxnet"
	Chainer                  = "chainer"
	Caffe                    = "caffe"
	Caffe2                   = "caffe2"
	Keras                    = "keras"
)

type ModelSpec struct {
	Framework FrameworkType `json:"frameworkType"`
}

type ModelStatus struct {
	Phase ModelPhase `json:"phase"`
	// size of the current Model
	Size resource.Quantity `json:"size"`
	// the total number of commits
	CommitNums int64 `json:"commitNums"`
	// last commit time, sort by default.
	LastCommitTime *metav1.Time `json:"lastCommitTime,omitempty"`
}

type ModelPhase string

const (
	ModelActive      ModelPhase = "Active"
	ModelTerminating ModelPhase = "Terminating"
	ModelUnknown     ModelPhase = "Unknown"
)
