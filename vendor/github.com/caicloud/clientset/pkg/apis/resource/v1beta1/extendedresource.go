package v1beta1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:nonNamespaced

// ExtendedResource describes a bare device.
type ExtendedResource struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec ExtendedResourceSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// Most recently observed status of the ExtendedResource.
	// +optional
	Status ExtendedResourceStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ExtendedResourceSpec is the specification of the desired behavior of the ExtendedResource.
type ExtendedResourceSpec struct {
	// raw resource name. E.g.: nvidia.com/gpu
	RawResourceName string `json:"rawResourceName" protobuf:"bytes,1,opt,name=rawResourceName"`
	// device unique id
	DeviceID string `json:"deviceID" protobuf:"bytes,2,opt,name=deviceID"`
	// resource metadata received from device plugin.
	// e.g., gpuType: k80, zone: us-west1-b
	Properties map[string]string `json:"properties" protobuf:"bytes,3,opt,name=properties"`
	// NodeName constraints that limit what nodes this resource can be accessed from.
	// This field influences the scheduling of pods that use this resource.
	NodeName string `json:"nodeName" protobuf:"bytes,4,opt,name=nodeName"`
}

// ExtendedResourceStatus is the most recently observed status of the ExtendedResource.
type ExtendedResourceStatus struct {
	// Phase indicates if the compute resource is available or pending
	// +optional
	Phase ExtendedResourcePhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase"`
	// A human-readable message indicating details about why ExtendedResource is in this phase
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message"`
	// Brief string that describes any failure, used for CLI etc
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`

	// Capacity represents the total resources of a ExtendedResource.
	// +optional
	Capacity v1.ResourceList `json:"capacity,omitempty" protobuf:"bytes,4,rep,name=capacity,casttype=ResourceList,castkey=ResourceName"`
	// Allocatable represents the ExtendedResource that are available for scheduling.
	// Defaults to Capacity.
	// +optional
	Allocatable v1.ResourceList `json:"allocatable,omitempty" protobuf:"bytes,5,rep,name=allocatable,casttype=ResourceList,castkey=ResourceName"`
}

// ExtendedResourcePhase defines ExtendedResource status.
type ExtendedResourcePhase string

const (
	// ExtendedResourceAvailable used for ExtendedResource that is already on a specific node.
	ExtendedResourceAvailable ExtendedResourcePhase = "Available"

	// ExtendedResourceBound used for ExtendedResource that is already using by pod.
	ExtendedResourceBound ExtendedResourcePhase = "Bound"

	// ExtendedResourcePending used for ExtendedResource that is not available now,
	// due to device plugin send device is unhealthy.
	ExtendedResourcePending ExtendedResourcePhase = "Pending"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExtendedResourceList is a collection of ExtendedResource.
type ExtendedResourceList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of ExtendedResource.
	Items []ExtendedResource `json:"items" protobuf:"bytes,2,rep,name=items"`
}
