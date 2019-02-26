package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	NetworkPolicyPlural      = "networkpolicies"
	DefaultNetworkpolicyName = "default"

	StatusAnnotationKey = "networkpolicy.tenant.caicloud.io/status"

	// if API to delete networkpolicy is called
	// annotation of status will be set
	// TODO(liubog2008): delete after finalizer is enabled
	DeletingAnnotation = "deleting"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []NetworkPolicy `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// This is a resource which is subset of NetworkPolicy of networking.caicloud.io
// see github.com/kubernetes/kubernetes/pkg/apis/networking/v1/types.go
type NetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NetworkPolicySpec `json:"spec"`
}

type NetworkPolicySpec struct {
	Ingress []NetworkPolicyIngressRule `json:"ingress,omitempty"`
}

type NetworkPolicyIngressRule struct {
	From NetworkPolicyPeer `json:"from,omitempty"`
}

type NetworkPolicyPeer struct {
	AllowAll          PeerType `json:"allowAll,omitempty"`
	PartitionSelector []string `json:"partitionSelector,omitempty"`
}

type PeerType string

const (
	PartitionPeer PeerType = "partition"
)
