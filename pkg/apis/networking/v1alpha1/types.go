/*
Copyright 2017 Caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

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
	// Type determines the type of LoadBalancer, valid options are: external, internal
	// For internal type, the specifications for providers can ONLY be service.
	Type LoadBalancerType `json:"type"`
	// Specification of the desired behavior of the nodes
	Nodes NodesSpec `json:"nodes"`
	// Specification of the desired behavior of the proxy
	Proxy ProxySpec `json:"proxy"`
	// Specification of the desired behavior of the providers
	Providers ProvidersSpec `json:"providers"`
}

// LoadBalancerType ...
type LoadBalancerType string

const (
	// LoadBalancerTypeInternal is internal type of LoadBalancer
	LoadBalancerTypeInternal LoadBalancerType = "internal"
	// LoadBalancerTypeExternal is internal type of LoadBalancer
	LoadBalancerTypeExternal LoadBalancerType = "external"
)

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
	Effect *apiv1.TaintEffect `json:"dedicated,omitempty"`
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
	ProxyTypeNginx ProxyType = "nginx"
	// ProxyTypeHaproxy for haproxy
	ProxyTypeHaproxy ProxyType = "haproxy"
	// ProxyTypeTraefik for traefik
	ProxyTypeTraefik ProxyType = "traefik"
)

// ProvidersSpec is a description of prividers
type ProvidersSpec struct {
	// k8s service
	Service *ServiceProvider `json:"service,omitempty"`
	// ipvs dr
	Ipvsdr *IpvsdrProvider `json:"ipvsdr,omitempty"`
	// aliyun slb
	Aliyun *AliyunProvider `json:"aliyun,omitempty"`
	// azure
	Azure *AzureProvider `json:"azure,omitempty"`
}

// ServiceProvider is a k8s service provider
// It provides a entrance for in-cluster applications
// to access the proxy (ingress controller)
// For internal type, the specifications for providers can ONLY be service.
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
	Replicas     int32    `json:"replicas,omitempty"`
	Deployments  []string `json:"deployments,omitempty"`
	IngressClass string   `json:"ingressClass,omitempty"`
	ConfigMap    string   `json:"configMap,omitempty"`
	TCPConfigMap string   `json:"tcpConfigMap,omitempty"`
	UDPConfigMap string   `json:"udpConfigMap,omitempty"`
}

// ProvidersStatuses represents the current status of Providers
type ProvidersStatuses struct {
	// k8s service
	Service *ServiceProviderStatus `json:"service,omitempty"`
	// ipvs dr
	Ipvsdr *IpvsdrProviderStatus `json:"ipvsdr,omitempty"`
	// aliyun slb
	Aliyun *AliyunProviderStatus `json:"aliyun,omitempty"`
	// azure
	Azure *AzureProviderStatus `json:"azure,omitempty"`
}

// ServiceProviderStatus represents the current status of the service provider
type ServiceProviderStatus struct {
}

// IpvsdrProviderStatus represents the current status of the ipvsdr provider
type IpvsdrProviderStatus struct {
	Vip  string `json:"vip"`
	Vrid *int   `json:"vrid"`
}

// AliyunProviderStatus represents the current status of the aliyun provider
type AliyunProviderStatus struct {
}

// AzureProviderStatus represents the current status of the azure provider
type AzureProviderStatus struct {
}
