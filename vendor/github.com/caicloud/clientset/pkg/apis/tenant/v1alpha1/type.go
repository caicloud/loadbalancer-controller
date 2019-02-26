package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TenantPlural is used by tenant CRD
	// It defines collection name of tenant
	TenantPlural = "tenants"
	// PartitionPlural is used by partition CRD
	// It defines collection name of partition
	PartitionPlural = "partitions"
	// ClusterQuotaPlural is used by cluster quota CRD
	// It defines collection name of clusterquota
	// TODO(liubog2008): clusterquota should be a single resource
	ClusterQuotaPlural = "clusterquotas"

	// SystemClusterQuota defines cluster quota name
	SystemClusterQuota = "system"

	// SystemTenant defines name of system tenant
	SystemTenant = "system-tenant"

	// ResourceNvidiaGPU defines nvidia gpu
	ResourceNvidiaGPU v1.ResourceName = "nvidia.com/gpu"

	// ResourceRequestsNvidiaGPU  defines nvidia gpu requests
	ResourceRequestsNvidiaGPU v1.ResourceName = "requests.nvidia.com/gpu"
)

var (
	// LimitedResourceNames defines resource which will be initialized into 0
	// if it is not be set explicitly
	// It is used to let our system resource allocation strictly
	LimitedResourceNames = []v1.ResourceName{
		"limits.cpu",
		"limits.memory",
		"requests.cpu",
		"requests.memory",
		"requests.nvidia.com/gpu",
	}
)

const (
	// Add tenant info to resource label
	// so Lister can filter by label
	// e.g. partition
	TenantLabelKey    = "tenant.tenant.caicloud.io"
	PartitionLabelKey = "partition.tenant.caicloud.io"
	DynamicLabelKey   = "dynamic.tenant.caicloud.io"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TenantList
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Tenant `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Tenant
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TenantSpec   `json:"spec"`
	Status            TenantStatus `json:"status,omitempty"`
}

type TenantSpec struct {
	Quota v1.ResourceList `json:"quota"`
	// Deprecated if GC were enabled by apiextensions apiserver
	Finalizers []FinalizerName `json:"finalizers,omitempty"`
}

type TenantStatus struct {
	Phase      TenantPhase       `json:"phase"`
	Conditions []TenantCondition `json:"conditions,omitempty"`
	ActualUsed v1.ResourceList   `json:"actualUsed"`
	Used       v1.ResourceList   `json:"used"`
	Hard       v1.ResourceList   `json:"hard"`
}

type TenantCondition struct {
	Type    ConditionType      `json:"type"`
	Status  v1.ConditionStatus `json:"status"`
	Reason  string             `json:"reason,omitempty"`
	Message string             `json:"message,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterQuotaList
type ClusterQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterQuota `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterQuota
type ClusterQuota struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterQuotaSpec   `json:"spec"`
	Status            ClusterQuotaStatus `json:"status,omitempty"`
}

type ClusterQuotaSpec struct {
	Ratio map[v1.ResourceName]int64 `json:"ratio"`
}

type ClusterQuotaStatus struct {
	Logical  `json:"logical"`
	Physical `json:"physical"`
}

type Logical struct {
	Total      v1.ResourceList `json:"total"`
	Allocated  v1.ResourceList `json:"allocated"`
	SystemUsed v1.ResourceList `json:"systemUsed"`
	Used       v1.ResourceList `json:"used"`
}

type Physical struct {
	Capacity    v1.ResourceList `json:"capacity"`
	Allocatable v1.ResourceList `json:"allocatable"`
	Unavailable v1.ResourceList `json:"unavailable"`
}

const (
	ResourcePartitions    v1.ResourceName = "tenant.caicloud.io/partitions"
	ResourceLoadbalancers v1.ResourceName = "tenant.caicloud.io/loadbalancers"
)

type TenantPhase string

const (
	TenantActive      TenantPhase = "Active"
	TenantTerminating TenantPhase = "Terminating"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PartitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Partition `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Partition
type Partition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PartitionSpec   `json:"spec"`
	Status            PartitionStatus `json:"status,omitempty"`
}

type PartitionSpec struct {
	Tenant     string          `json:"tenant"`
	Quota      v1.ResourceList `json:"quota"`
	Finalizers []FinalizerName `json:"finalizers,omitempty"`
}

type PartitionStatus struct {
	Phase      PartitionPhase       `json:"phase"`
	Conditions []PartitionCondition `json:"conditions"`
	Used       v1.ResourceList      `json:"used"`
	// Hard is always equal with resource quota
	// When partition spec is updated and exceeded tenant quota
	// partition status will be blocked to update
	Hard v1.ResourceList `json:"hard"`
}

type FinalizerName string

const (
	FinalizerCaicloud FinalizerName = "caicloud"
)

type PartitionPhase string

const (
	PartitionActive      PartitionPhase = "Active"
	PartitionTerminating PartitionPhase = "Terminating"
)

type PartitionCondition struct {
	Type    ConditionType      `json:"type"`
	Status  v1.ConditionStatus `json:"status"`
	Reason  string             `json:"reason,omitempty"`
	Message string             `json:"message,omitempty"`
}

type ConditionType string

const (
	// ExceedsQuota will be true if used resource exceeds quota
	ExceedsQuota ConditionType = "ExceedsQuota"

	// TerminatingTimeout will be true if terminating partition costs too long
	TerminatingTimeout ConditionType = "TerminatingTimeout"
)

const (
	// DeletionTimeKey defines deletion time of resource
	// Now update DeletionTime in metadata is not supported
	// TODO(liubog2008): change to use metadata.deletionTime
	DeletionTimeKey string = "deletionTimestamp"
)
