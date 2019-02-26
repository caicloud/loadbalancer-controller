package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cargo describes an instance of Cargo registry.
type Cargo struct {
	// Metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// Metadata for the particular object, including name, namespace, labels, etc
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Custom resource specification
	Spec CargoSpec `json:"spec"`
	// Status
	Status CargoStatus `json:"status,omitempty"`
}

// CargoSpec describes specification of a cargo registry.
type CargoSpec struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// CargoStatus describes status of a cargo registry
type CargoStatus struct {
	Healthy bool `json:"healthy"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CargoList describes an array of Cargo instances.
type CargoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Cargo `json:"items""`
}
