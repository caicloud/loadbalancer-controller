package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IstioList describes an array of istio
type IstioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Istio `json:"items,omitempty"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Istio describes configuration and running status of istio components.
type Istio struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the istio.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec IstioSpec `json:"spec,omitempty"`
	// Most recently observed status of the istio.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Status IstioStatus `json:"status,omitempty"`
}

// IstioSpec describes configuration of istio components
type IstioSpec struct {
	Cluster string `json:"cluster,omitempty"`
	// Version is expected version of the istio components.
	Version string `json:"version,omitempty"`
	// Pilot describes configuration of pilot.
	Pilot *Pilot `json:"pilot,omitempty"`
	// Paused is to pause the control of the operator for istio.
	Paused bool `json:"paused,omitempty"`
}

// Pilot describes configuration for pilot
type Pilot struct {
	// TraceSampling is a percentage between 0 to 100
	TraceSampling float64 `json:"traceSampling,omitempty"`
}

// IstioStatus describes status of istio components
type IstioStatus struct {
	Phase IstioPhase `json:"phase,omitempty"`
}

// IstioPhase including Installing, Running and Terminating
type IstioPhase string

const (
	IstioPhasePending     IstioPhase = "Pending"
	IstioPhaseCreating    IstioPhase = "Creating"
	IstioPhaseRunning     IstioPhase = "Running"
	IstioPhaseTerminating IstioPhase = "Terminating"
)
