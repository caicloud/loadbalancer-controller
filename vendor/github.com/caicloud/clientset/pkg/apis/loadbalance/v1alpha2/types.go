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
	// HTTPPort is the port that LoadBalancer listen http protocol
	// default is 80
	HTTPPort int `json:"httpPort,omitempty"`
	// HTTPSPort is the port that LoadBalancer listen https protocol
	// default is 443
	HTTPSPort int `json:"httpsPort,omitempty"`
	// PortRanges define a list of port-ranges the proxy can use
	// default is [{20000,29999}]
	PortRanges []PortRange `json:"portRanges,omitempty"`
}

// PortRange describe a port range in {start, end}
type PortRange struct {
	Start int32 `json:"start"`
	End   int32 `json:"end"`
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
	// ProxyTypeKong for kong
	ProxyTypeKong ProxyType = "kong"
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
	// Name azure loadbalancer name
	Name string `json:"name,omitempty"`
	// ResourceGroupName Azure resource group name
	ResourceGroupName string `json:"resourceGroupName"`
	// Location - Resource location of china.
	Location string `json:"location"`
	// SKU The load balancer SKU.
	// explanation https://docs.microsoft.com/en-us/azure/load-balancer/load-balancer-overview
	SKU AzureSKUKind `json:"sku"`
	// ClusterID cluster id
	ClusterID string `json:"clusterID"`
	// ReserveAzure This flag tells the controller to reserve azure loadbalancer when
	// deleting compass loadbalancer
	ReserveAzure *bool `json:"reserveAzure,omitempty"`
	// IPAddress azure loadbalancer IP address properties
	IPAddressProperties AzureIPAddressProperties `json:"ipAddressProperties"`
}

// AzureIPAddressProperties azure loadbalancer IP address properties
type AzureIPAddressProperties struct {
	// Private private IP address properties
	Private *AzurePrivateIPAddressProperties `json:"private,omitempty"`
	// Public public IP address properties
	Public *AzurePublicIPAddressProperties `json:"public,omitempty"`
}

// AzurePrivateIPAddressProperties  azure loadbalancer private IP address properties
type AzurePrivateIPAddressProperties struct {
	// IPAllocationMethod - The Private IP allocation method.
	IPAllocationMethod AzureIPAllocationMethodKind `json:"ipAllocationMethod"`
	// VPC virtual private cloud id
	VPCID string `json:"vpcID"`
	// SubnetID - The reference of the subnet resource id.
	SubnetID string `json:"subnetID"`
	// PrivateIPAddress - The private IP address of the IP configuration.
	PrivateIPAddress *string `json:"privateIPAddress,omitempty"`
}

// AzurePublicIPAddressProperties azure loadbalancer public IP address properties
type AzurePublicIPAddressProperties struct {
	// IPAllocationMethod  the public IP allocation method.
	IPAllocationMethod AzureIPAllocationMethodKind `json:"ipAllocationMethod"`
	// PublicIPAddressID - The reference of the Public IP resource ID.
	PublicIPAddressID *string `json:"publicIPAddressID,omitempty"`
}

// AzureIPAddressType loadbalancer based on network type
type AzureIPAddressType string

const (
	// AzurePrivateIPAddressType private network
	AzurePrivateIPAddressType AzureIPAddressType = "private"
	// AzurePublicIPAddressType public network
	AzurePublicIPAddressType AzureIPAddressType = "public"
)

// AzureSKUKind The load balancer SKU.
type AzureSKUKind string

const (
	// AzureStandardSKU Standard sku
	AzureStandardSKU AzureSKUKind = "Standard"
	// AzureBasicSKU Basic sku
	AzureBasicSKU AzureSKUKind = "Basic"
)

// AzureIPAllocationMethodKind enumerates the values for ip allocation method.
type AzureIPAllocationMethodKind string

const (
	// AzureStaticIPAllocationMethod static ip allocation method
	AzureStaticIPAllocationMethod AzureIPAllocationMethodKind = "Static"
	// AzureDynamicIPAllocationMethod Dynamic ip allocation method
	AzureDynamicIPAllocationMethod AzureIPAllocationMethodKind = "Dynamic"
)

// LoadBalancerStatus represents the current status of a LoadBalancer
type LoadBalancerStatus struct {
	// Accessible specify if the loadbalancer is ready for access
	Accessible bool `json:"accessible,omitempty"`
	// AccessIPs specify the entrance ip of loadbalancer
	AccessIPs []string `json:"accessIPs,omitempty"`
	// NodeIPs specify the entrance node ip of loadbalancer
	NodeIPs []string `json:"nodeIPs,omitempty"`
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

// AzureProviderStatus represents the current status of the azure lb provider
type AzureProviderStatus struct {
	// Phase azure loadbalancer phase
	Phase AzureProviderPhase `json:"phase"`
	// Reason azure loadbalancer error reason
	Reason string `json:"reason,omitempty"`
	// Message azure lb create or update failed message
	Message string `json:"message,omitempty"`
	// ProvisioningState azure lb state
	ProvisioningState string `json:"provisioningState,omitempty"`
	// PublicIPAddress - The reference of the Public IP address.
	PublicIPAddress *string `json:"publicIPAddress,omitempty"`
}

// AzureProviderPhase azure loadbalancer phase
type AzureProviderPhase string

const (
	// AzureProgressingPhase progressing phase
	AzureProgressingPhase AzureProviderPhase = "Progressing"
	// AzureRunningPhase running phase
	AzureRunningPhase AzureProviderPhase = "Running"
	// AzureErrorPhase error phase
	AzureErrorPhase AzureProviderPhase = "Error"
	// AzureUpdatingPhase update phase
	AzureUpdatingPhase AzureProviderPhase = "Updating"
)

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
