package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster describes a cluster with kubernetes and addon
//
// Cluster are non-namespaced; the id of the cluster
// according to etcd is in ObjectMeta.Name.
type Cluster struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ClusterSpec   `json:"spec"`
	Status ClusterStatus `json:"status"`
}

type ClusterSpec struct {
	DisplayName      string           `json:"displayName"`
	Provider         CloudProvider    `json:"provider"`
	IsControlCluster bool             `json:"isControlCluster"`
	Network          ClusterNetwork   `json:"network"`
	IsHighAvailable  bool             `json:"isHighAvailable"`
	MastersVIP       string           `json:"mastersVIP"`
	Auth             ClusterAuth      `json:"auth"`
	Versions         *ClusterVersions `json:"versions,omitempty"`
	Masters          []string         `json:"masters"`
	Nodes            []string         `json:"nodes"`
	// adapt expired
	ClusterToken string       `json:"clusterToken"`
	Ratio        ClusterRatio `json:"ratio"`
}

type ClusterStatus struct {
	Phase         ClusterPhase                       `json:"phase"`
	Conditions    []ClusterCondition                 `json:"conditions"`
	Masters       []MachineThumbnail                 `json:"masters"`
	Nodes         []MachineThumbnail                 `json:"nodes"`
	Capacity      map[ResourceName]resource.Quantity `json:"capacity"`
	OperationLogs []OperationLog                     `json:"operationLogs,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList is a collection of clusters
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Clusters
	Items []Cluster `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Machine describes real machines
//
// Machine are non-namespaced; the id of the machine
// according to etcd is in ObjectMeta.Name.
type Machine struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   MachineSpec   `json:"spec"`
	Status MachineStatus `json:"status"`
}

type MachineSpec struct {
	Provider       CloudProvider       `json:"provider"`
	ProviderConfig CloudProviderConfig `json:"providerConfig"`
	Address        []NodeAddress       `json:"address"`
	SshPort        string              `json:"sshPort"`
	Auth           MachineAuth         `json:"auth"`
	Versions       MachineVersions     `json:"versions,omitempty"`
	Cluster        string              `json:"cluster"`
	IsMaster       bool                `json:"isMaster"`
	Tags           map[string]string   `json:"tags"`
}

type MachineStatus struct {
	Phase MachinePhase `json:"phase"`
	// env
	Environment MachineEnvironment `json:"environment"`
	// node about
	NodeRefer  string                             `json:"nodeRefer"`
	Capacity   map[ResourceName]resource.Quantity `json:"capacity"`
	NodeStatus MachineNodeStatus                  `json:"nodeStatus"`
	// other
	OperationLogs []OperationLog `json:"operationLogs,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineList is a collection of machine.
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Machines
	Items []Machine `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Tag describes machine tags history
//
// Tag are non-namespaced; the id of the tag
// according to etcd is in ObjectMeta.Name.
type Tag struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Values []string `json:"values"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TagList is a collection of tag.
type TagList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Tags
	Items []Tag `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Config describes login information, ssh keys
//
// Config are non-namespaced; the id of the login
// according to etcd is in ObjectMeta.Name.
type Config struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Values map[string][]byte `json:"values"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigList is a collection of login.
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Configs
	Items []Config `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// outside

type ClusterNetwork struct {
	Type NetworkType `json:"type"`
	// TODO
}

type ClusterAuth struct {
	KubeUser     string `json:"kubeUser"`
	KubePassword string `json:"kubePassword"`
	KubeToken    string `json:"kubeToken"`
	KubeCertPath string `json:"kubeCertPath"`
	KubeCAData   []byte `json:"kubeCAData"`
	EndpointIP   string `json:"endpointIP"`
	EndpointPort string `json:"endpointPort"`
}

type ClusterVersions struct {
	MasterSets map[string]string
	NodeSets   MachineVersions
}

type ClusterCondition struct {
	Type               ClusterConditionType `json:"type"`
	Status             ConditionStatus      `json:"status"`
	LastHeartbeatTime  metav1.Time          `json:"lastHeartbeatTime"`
	LastTransitionTime metav1.Time          `json:"lastTransitionTime"`
	Reason             string               `json:"reason"`
	Message            string               `json:"message"`
}

type ClusterRatio struct {
	CpuOverCommitRatio    float64 `json:"cpuOverCommitRatio"`
	MemoryOverCommitRatio float64 `json:"memoryOverCommitRatio"`
}

type MachineThumbnail struct {
	Name   string       `json:"name"`
	Status MachinePhase `json:"status"`
}

type MachineAuth struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Key      string `json:"key"`
}

type NodeAddress struct {
	Type    NodeAddressType `json:"type"`
	Address string          `json:"address"`
}

type MachineNodeStatus struct {
	Unschedulable bool            `json:"unschedulable"`
	Conditions    []NodeCondition `json:"conditions"`
}

type NodeCondition = corev1.NodeCondition

type OperationLog struct {
	Time     metav1.Time       `json:"time"`
	Operator string            `json:"operator"`
	Type     OperationLogType  `json:"type"`
	Field    OperationLogField `json:"field"`
	Value    string            `json:"value"`
	Detail   string            `json:"detail"`
}

// little types

type CloudProvider string
type CloudProviderConfig map[string]string

type ClusterPhase string
type MachinePhase string
type PodPhase string

type ClusterConditionType string
type NodeConditionType = corev1.NodeConditionType
type ConditionStatus = corev1.ConditionStatus

type NetworkType string

type NodeAddressType string

type ResourceName string

type OperationLogType string
type OperationLogField string

type MachineVersions map[string]string // TODO
type MachineHardware map[string]string // TODO

// agent

type MachineEnvironment struct {
	SystemInfo         MachineSystemInfo   `json:"systemInfo"`
	HardwareInfo       MachineHardwareInfo `json:"hardwareInfo"`
	DiskInfo           []MachineDiskInfo   `json:"diskInfo"`
	NicInfo            []MachineNicInfo    `json:"nicInfo"`
	GPUInfo            []MachineGPUInfo    `json:"gpuInfo"`
	LastTransitionTime metav1.Time         `json:"lastTransitionTime"`
}

type MachineSystemInfo struct {
	BootTime        uint64 `json:"bootTime"`
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platformFamily"`
	PlatformVersion string `json:"platformVersion"`
	KernelVersion   string `json:"kernelVersion"`
}

type MachineHardwareInfo struct {
	CPUModel         string  `json:"cpuModel"`
	CPUArch          string  `json:"cpuArch"`
	CPUMHz           float64 `json:"cpuMHz"`
	CPUCores         int     `json:"cpuCores"`
	CPUPhysicalCores int     `json:"cpuPhysicalCores"`
	MemoryTotal      uint64  `json:"memoryTotal"`
}

type MachineDiskInfo struct {
	Device     string `json:"device"`
	Capacity   uint64 `json:"capacity"`
	Type       string `json:"type"`
	MountPoint string `json:"mountPoint"`
}

type MachineNicInfo struct {
	Name         string   `json:"name"`
	MTU          string   `json:"mtu"`
	Speed        string   `json:"speed"`
	HardwareAddr string   `json:"hardwareAddr"`
	Status       string   `json:"status"`
	Addrs        []string `json:"addrs"`
}

type MachineGPUInfo struct {
	UUID             string `json:"uuid"`
	ProductName      string `json:"productName"`
	ProductBrand     string `json:"productBrand"`
	PCIeGen          string `json:"pcieGen"`
	PCILinkWidths    string `json:"pciLinkWidths"`
	MemoryTotal      string `json:"memoryTotal"`
	MemoryClock      string `json:"memoryClock"`
	GraphicsAppClock string `json:"graphicsAppClock"`
	GraphicsMaxClock string `json:"graphicsMaxClock"`
}
