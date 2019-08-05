package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceClass defines how to choose the ExtendedResource.
type ResourceClass struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec ResourceClassSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// ResourceClassSpec is the specification of the desired behavior of the ResourceClass.
type ResourceClassSpec struct {
	// raw resource name. E.g.: nvidia.com/gpu
	RawResourceName string `json:"rawResourceName" protobuf:"bytes,1,opt,name=rawResourceName"`
	// defines general resource property matching constraints.
	// e.g.: zone in { us-west1-b, us-west1-c }; type: k80
	Requirements metav1.LabelSelector `json:"requirements" protobuf:"bytes,2,opt,name=requirements"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceClassList is a collection of ResourceClass.
type ResourceClassList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of ResourceClass.
	Items []ResourceClass `json:"items" protobuf:"bytes,2,rep,name=items"`
}
