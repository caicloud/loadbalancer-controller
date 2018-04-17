/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1alpha2

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancerList is a collection of LoadBalancer
type LoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []LoadBalancer `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LoadBalancer describes a LoadBalancer which provides Load Balancing for applications
// LoadBalancer contains a proxy and multiple providers to load balance
// either internal or external traffic.
//
// A proxy is an ingress controller watching ingress resource to provide access that
// allow inbound connections to reach the cluster services
//
// A provider is the entrance of the cluster providing high availability for connections
// to proxy (ingress controller)
type LoadBalancer struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Specification of the desired behavior of the LoadBalancer
	Spec LoadBalancerSpec `json:"spec,omitempty"`
	// Most recently observed status of the loadbalancer.
	// This data may not be up to date.
	// Populated by the system.
	// Read-only.
	// +optional
	Status LoadBalancerStatus `json:"status,omitempty"`
}

// LoadBalancerSpec is a description of a LoadBalancer
type LoadBalancerSpec struct {
	// Specification of the desired behavior of the nodes
	Nodes NodesSpec `json:"nodes"`
	// Specification of the desired behavior of the proxy
	Proxy ProxySpec `json:"proxy"`
	// Specification of the desired behavior of the providers
	Providers ProvidersSpec `json:"providers"`
}

// NodesSpec is a description of nodes
type NodesSpec struct {
	// Replica is only used when Provider's type is service now
	// you can not use replica and names at the same time
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Names is a name list of nodes selected to run proxy
	// It MUST be filled in when loadbalancer's type is external
	// +optional
	Names []string `json:"names,omitempty"`
	// +optional
	Effect *v1.TaintEffect `json:"taintEffect,omitempty"`
}

// ProxySpec is a description of a proxy
type ProxySpec struct {
	Type ProxyType `json:"type"`
	// Config contains the optional config of proxy
	Config map[string]string `json:"config,omitempty"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// +optional
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

// ProxyType ...
type ProxyType string

const (
	// ProxyTypeNginx for nginx
	ProxyTypeNginx ProxyType = "nginx"
	// ProxyTypeHaproxy for haproxy
	ProxyTypeHaproxy ProxyType = "haproxy"
	// ProxyTypeTraefik for traefik
	ProxyTypeTraefik ProxyType = "traefik"
)

// ProvidersSpec is a description of prividers
type ProvidersSpec struct {
	// external provider
	External *ExternalProvider `json:"external,omitempty"`
	// ipvs dr
	Ipvsdr *IpvsdrProvider `json:"ipvsdr,omitempty"`
	// aliyun slb
	Aliyun *AliyunProvider `json:"aliyun,omitempty"`
	// azure
	Azure *AzureProvider `json:"azure,omitempty"`
}

// ExternalProvider is a provider docking for external loadbalancer
type ExternalProvider struct {
	VIP string `json:"vip"`
}

// IpvsdrProvider is a ipvs dr provider
type IpvsdrProvider struct {
	// Virtual IP Address
	VIP string `json:"vip"`
	// ipvs shceduler algorithm type
	Scheduler IpvsScheduler `json:"scheduler"`
}

// IpvsScheduler is ipvs shceduler algorithm type
type IpvsScheduler string

const (
	// IpvsSchedulerRR - Round Robin
	IpvsSchedulerRR IpvsScheduler = "rr"
	// IpvsSchedulerWRR - Weighted Round Robin
	IpvsSchedulerWRR IpvsScheduler = "wrr"
	// IpvsSchedulerLC - Round Robin
	IpvsSchedulerLC IpvsScheduler = "lc"
	// IpvsSchedulerWLC - Weighted Least Connections
	IpvsSchedulerWLC IpvsScheduler = "wlc"
	// IpvsSchedulerLBLC - Locality-Based Least Connections
	IpvsSchedulerLBLC IpvsScheduler = "lblc"
	// IpvsSchedulerDH - Destination Hashing
	IpvsSchedulerDH IpvsScheduler = "dh"
	// IpvsSchedulerSH - Source Hashing
	IpvsSchedulerSH IpvsScheduler = "sh"
)

// AliyunProvider ...
type AliyunProvider struct {
	Name string `json:"name,omitempty"`
}

// AzureProvider ...
type AzureProvider struct {
	Name string `json:"name,omitempty"`
}

// LoadBalancerStatus represents the current status of a LoadBalancer
type LoadBalancerStatus struct {
	// +optional
	ProxyStatus ProxyStatus `json:"proxyStatus"`
	// +optional
	ProvidersStatuses ProvidersStatuses `json:"providersStatuses"`
}

// ProxyStatus represents the current status of a Proxy
type ProxyStatus struct {
	PodStatuses  `json:",inline"`
	Deployment   string `json:"deployment,omitempty"`
	IngressClass string `json:"ingressClass,omitempty"`
	ConfigMap    string `json:"configMap,omitempty"`
	TCPConfigMap string `json:"tcpConfigMap,omitempty"`
	UDPConfigMap string `json:"udpConfigMap,omitempty"`
}

// ProvidersStatuses represents the current status of Providers
type ProvidersStatuses struct {
	// external loadbalancer provider
	External *ExpternalProviderStatus `json:"external,omitempty"`
	// ipvs dr
	Ipvsdr *IpvsdrProviderStatus `json:"ipvsdr,omitempty"`
	// aliyun slb
	Aliyun *AliyunProviderStatus `json:"aliyun,omitempty"`
	// azure
	Azure *AzureProviderStatus `json:"azure,omitempty"`
}

// ExpternalProviderStatus represents the current status of the external provider
type ExpternalProviderStatus struct {
	VIP string `json:"vip"`
}

// IpvsdrProviderStatus represents the current status of the ipvsdr provider
type IpvsdrProviderStatus struct {
	PodStatuses `json:",inline"`
	Deployment  string `json:"deployment,omitempty"`
	VIP         string `json:"vip"`
	Vrid        *int   `json:"vrid,omitempty"`
}

// AliyunProviderStatus represents the current status of the aliyun provider
type AliyunProviderStatus struct {
}

// AzureProviderStatus represents the current status of the azure provider
type AzureProviderStatus struct {
}

// PodStatuses represents the current statuses of a list of pods
type PodStatuses struct {
	Replicas      int32       `json:"replicas"`
	TotalReplicas int32       `json:"totalReplicas"`
	ReadyReplicas int32       `json:"readyReplicas"`
	Statuses      []PodStatus `json:"podStatuses"`
}

// PodStatus represents the current status of pods
type PodStatus struct {
	Name            string `json:"name"`
	Ready           bool   `json:"ready"`
	RestartCount    int32  `json:"restartCount"`
	ReadyContainers int32  `json:"readyContainers"`
	TotalContainers int32  `json:"totalContainers"`
	NodeName        string `json:"nodeName"`
	Phase           string `json:"phase"`
	Reason          string `json:"reason"`
	Message         string `json:"message"`
}
