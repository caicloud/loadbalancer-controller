package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

// LoadBalancerList is a collection of LoadBalancer
type LoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []LoadBalancer `json:"items"`
}

// +genclient=true

// LoadBalancer describe a LoadBalancer witch providers Load balancing for services
type LoadBalancer struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the LoadBalancer
	LoadBalancerSpec `json:"spec,omitempty"`
}

// LoadBalancerSpec is a description of a LoadBalancer
type LoadBalancerSpec struct {
	// Type determines the type of LoadBalancer
	// if type is internal, the providers MUST and ONLY can be service
	Type LoadBalancerType `json:"type"`
	// Replica only used when Provider's type is service
	// you can not use replica and nodes at the same time
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Specification of the desired behavior of the nodes
	// If you set nodes, replicas will be disabled
	// +optional
	Nodes *NodesSpec `json:"nodes,omitempty"`
	// Specification of the desired behavior of the proxy
	Proxy ProxySpec `json:"proxy"`
	// Specification of the desired behavior of the providers
	Providers ProvidersSpec `json:"providers"`
}

// LoadBalancerType ...
type LoadBalancerType string

const (
	// LoadBalancerTypeInternal is internal type of LoadBalancer
	LoadBalancerTypeInternal = "internal"
	// LoadBalancerTypeExternal is internal type of LoadBalancer
	LoadBalancerTypeExternal = "external"
)

// NodesSpec is a description of nodes
type NodesSpec struct {
	Names []string `json:"names"`
	// +optional
	Dedicated *apiv1.TaintEffect `json:"dedicated,omitempty"`
}

// ProxySpec is a description of a proxy
type ProxySpec struct {
	Type ProxyType `json:"type"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// +optional
	Resources apiv1.ResourceRequirements `json:"resources,omitempty"`
}

// ProxyType ...
type ProxyType string

const (
	// ProxyTypeNginx for nginx
	ProxyTypeNginx = "nginx"
)

// ProvidersSpec is a description of prividers
type ProvidersSpec struct {
	// k8s service
	Service *ServiceProvider `json:"service"`
	// ipvs dr
	Ipvsdr *IpvsdrProvider `json:"ipvsdr"`
	// aliyun slb
	Aliyun *AliyunProvider `json:"aliyun"`
	// azure
	Azure *AzureProvider `json:"azure"`
}

// ServiceProvider is a k8s service provider
type ServiceProvider struct {
	Name string `json:"name"`
}

// IpvsdrProvider is a ipvs dr provider
type IpvsdrProvider struct {
	// Virtual IP Address
	Vip string `json:"vip"`
	// ipvs shceduler algorithm type
	Scheduler IpvsScheduler `json:"scheduler"`
}

// IpvsScheduler is ipvs shceduler algorithm type
type IpvsScheduler string

const (
	// IpvsSchedulerRR - Round Robin
	IpvsSchedulerRR = "rr"
	// IpvsSchedulerWRR - Weighted Round Robin
	IpvsSchedulerWRR = "wrr"
	// IpvsSchedulerLC - Round Robin
	IpvsSchedulerLC = "lc"
	// IpvsSchedulerWLC - Weighted Least Connections
	IpvsSchedulerWLC = "wlc"
	// IpvsSchedulerLBLC - Locality-Based Least Connections
	IpvsSchedulerLBLC = "lblc"
	// IpvsSchedulerDH - Destination Hashing
	IpvsSchedulerDH = "dh"
	// IpvsSchedulerSH - Source Hashing
	IpvsSchedulerSH = "sh"
)

// AliyunProvider ...
type AliyunProvider struct {
	Name string `json:"name"`
}

// AzureProvider ...
type AzureProvider struct {
	Name string `json:"name"`
}

// LoadBalancerStatus represents the current status of a LoadBalancer
type LoadBalancerStatus struct {
	// +optional
	// Conditions        []LoadBalancerCondition `json:"conditions,omitempty"`
	ProxyStatus       string
	ProvidersStatuses string
}
