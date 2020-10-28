package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
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

	// Spec defines a specification of a cluster.
	// Provisioned by an administrator.
	Spec ClusterSpec `json:"spec"`

	// Status represents the current information/status for the cluster.
	// Populated by the system.
	Status ClusterStatus `json:"status"`
}

// ClusterSpec is a description of a cluster.
type ClusterSpec struct {
	// DisplayName is the human-readable name for cluster.
	DisplayName string `json:"displayName"`

	Provider         CloudProvider              `json:"provider"`
	ProviderConfig   ClusterCloudProviderConfig `json:"providerConfig"`
	IsControlCluster bool                       `json:"isControlCluster"`
	Network          ClusterNetwork             `json:"network"`
	IsHighAvailable  bool                       `json:"isHighAvailable"`
	MastersVIP       string                     `json:"mastersVIP"`
	Auth             ClusterAuth                `json:"auth"`
	Versions         *ClusterVersions           `json:"versions,omitempty"`
	Masters          []string                   `json:"masters"`
	Nodes            []string                   `json:"nodes"`
	Etcds            []string                   `json:"etcds"`
	// deploy
	DeployToolsExternalVars map[string]string `json:"deployToolsExternalVars"`
	// adapt expired
	ClusterToken string       `json:"clusterToken"`
	Ratio        ClusterRatio `json:"ratio"`
}

// ClusterStatus represents information about the status of a cluster.
type ClusterStatus struct {
	Phase          ClusterPhase                       `json:"phase"`
	Conditions     []ClusterCondition                 `json:"conditions"`
	Masters        []MachineThumbnail                 `json:"masters"`
	Nodes          []MachineThumbnail                 `json:"nodes"`
	Etcds          []MachineThumbnail                 `json:"etcds"`
	Capacity       map[ResourceName]resource.Quantity `json:"capacity"`
	OperationLogs  []OperationLog                     `json:"operationLogs,omitempty"`
	AutoScaling    ClusterAutoScalingStatus           `json:"autoScaling,omitempty"`
	ProviderStatus ClusterProviderStatus              `json:"providerStatus,omitempty"`
	Versions       *ClusterVersions                   `json:"versions,omitempty"`
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
// +genclient:nonNamespaced
// +genclient:noStatus
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

// MachineTaint is a type alias for Kubernetes Taint struct.
type MachineTaint = corev1.Taint

// MachineSpec is a description of a machine.
type MachineSpec struct {
	Provider         CloudProvider              `json:"provider"`
	ProviderConfig   MachineCloudProviderConfig `json:"providerConfig,omitempty"`
	Address          []NodeAddress              `json:"address"`
	SshPort          string                     `json:"sshPort"`
	Auth             MachineAuth                `json:"auth"`
	Versions         MachineVersions            `json:"versions,omitempty"`
	Cluster          string                     `json:"cluster"`
	IsMaster         bool                       `json:"isMaster"`
	IsEtcd           bool                       `json:"isEtcd"`
	HostnameReadonly bool                       `json:"hostnameReadonly,omitempty"`
	Tags             map[string]string          `json:"tags"`

	// Taints is a list of taints of a machine. All taints here will be synced to a corresponding Node object when adding a machine to a cluster, and will be cleared when the machine is removed from a cluster.
	// This field is used internally via machine controller to persistently represent taints when a machine is added to a cluster (when Node object doesn't exist); otherwise the information will be lost across component restart.
	Taints []MachineTaint `json:"taints,omitempty"`
}

// MachineStatus represents information about the status of a machine.
type MachineStatus struct {
	// NOTE: this field will been deprecated in v2.11,
	// should use the value calculated from spec, conditions, refers, status
	Phase MachinePhase `json:"phase"`

	// env
	// has value if got any
	Environment *MachineEnvironment `json:"environment,omitempty"`

	// role
	// machine is a master in cluster
	IsMaster bool `json:"isMaster"`

	// IsEtcd represents if the node is etcd-specify node
	IsEtcd bool `json:"isEtcd"`

	// reference

	// refer cluster name
	// has value if assigned to a specific cluster
	ClusterRefer string `json:"clusterRefer"`

	// refer node claim name
	// has value if assigned to a specific cluster and node claim created
	NodeClaimRefer string `json:"nodeClaimRefer"`

	// refer node name
	// has value if assigned to a specific cluster and node created
	NodeRefer string `json:"nodeRefer"`

	// machine capacity in resource list format
	// Deprecated: node about will added in admin, not store in crd
	Capacity map[ResourceName]resource.Quantity `json:"capacity"`
	// sync from node status
	// Deprecated: split with node
	NodeStatus MachineNodeStatus `json:"nodeStatus"`

	// other
	OperationLogs []OperationLog `json:"operationLogs,omitempty"`

	// Current service state of machine.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []MachineCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`

	// cloud status
	ProviderStatus MachineProviderStatus `json:"providerStatus,omitempty"`
}

// MachineProviderStatus represents information about the cloud status of a machine.
type MachineProviderStatus struct {
	Aliyun *AliyunMachineCloudProviderStatus `json:"aliyun,omitempty"`
}

// AliyunMachineCloudProviderStatus represents information about the aliyun machine status.
type AliyunMachineCloudProviderStatus struct {
	Status          AliyunInstanceStatusType `json:"status,omitempty"`
	AutoReleaseTime string                   `json:"autoReleaseTime,omitempty"`
}

// MachineConditionType is a valid value for MachineCondition.Type
type MachineConditionType string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// MachineCondition contains details for the current condition of this machine.
type MachineCondition struct {
	// Type is the type of the condition.
	// Currently only Ready.
	Type MachineConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=MachineConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
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
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeClaim describes claim of node
//
// NodeClaim are non-namespaced; the id of the node claim
// according to etcd is in ObjectMeta.Name.
type NodeClaim struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   NodeClaimSpec   `json:"spec"`
	Status NodeClaimStatus `json:"status"`
}

// NodeClaimSpec is a description of a node claim.
type NodeClaimSpec struct {
	// template
	Template MachineTemplate `json:"template"`
}

type MachineTemplate struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the machine.
	Spec MachineSpec `json:"spec"`
}

// NodeClaimStatus represents information about the status of a node claim.
type NodeClaimStatus struct {
	// node reference
	// has value if node created
	NodeRefer string `json:"nodeRefer"`
	// role
	// node claim is a master in cluster
	IsMaster bool `json:"isMaster"`

	// IsEtcd represents node claim is a etcd in cluster
	IsEtcd bool `json:"isEtcd"`
	// Current service state of node claim.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []NodeClaimCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

// NodeClaimCondition contains details for the current condition of this node claim.
type NodeClaimCondition MachineCondition

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeClaimList is a collection of node claim.
type NodeClaimList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of NodeClaims
	Items []NodeClaim `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// MachineCloudProviderConfig is a description of cloud machine config.
type MachineCloudProviderConfig struct {
	// auto scaling group name which this machine belongs to
	// empty means not belongs to any auto scaling group
	AutoScalingGroup string `json:"autoScalingGroup,omitempty"`

	// Azure is the specific config of Azure VM
	Azure *AzureMachineCloudProviderConfig `json:"azure,omitempty"`

	// Aliyun is the specific config of Aliyun ECS
	Aliyun *AliyunMachineCloudProviderConfig `json:"aliyun,omitempty"`
}

// ClusterCloudProviderConfig is a description of cloud cluster config.
type ClusterCloudProviderConfig struct {
	// cluster level auto scaling setting
	// maybe nil in old or control cluster, need to be inited with a default setting by controller
	AutoScalingSetting *ClusterAutoScalingSetting `json:"autoScalingSetting,omitempty"`

	// Azure is the specific config of Azure Cluster
	Azure *AzureClusterCloudProviderConfig `json:"azure,omitempty"`

	// AzureAks is the specific config of Azure Aks Cluster
	AzureAks *AzureAksClusterCloudProviderConfig `json:"azureAks,omitempty"`

	// Aliyun is the specific config of aliyun cluster
	Aliyun *AliyunClusterCloudProviderConfig `json:"aliyun,omitempty"`
}

// AliyunInstanceType is a label for the aliyun instance type
type AliyunInstanceType struct {
	Name   string `json:"name,omitempty"`
	Family string `json:"family,omitempty"`
}

// AliyunInstanceNetworkType is a label for the aliyun instance network type
type AliyunInstanceNetworkType string

const (
	// AliyunInstanceNetworkClassic for classic network, doesn't support
	AliyunInstanceNetworkClassic AliyunInstanceNetworkType = "Classic"
	// AliyunInstanceNetworkVPC for vpc network, only support
	AliyunInstanceNetworkVPC AliyunInstanceNetworkType = "Vpc"
)

// AliyunInstanceChargeType is a label for the aliyun instance charge type
type AliyunInstanceChargeType string

const (
	// AliyunInstanceChargePostPaid for paid by quantity
	AliyunInstanceChargePostPaid AliyunInstanceChargeType = "PostPaid"
	// AliyunInstanceChargePrePaid for paid before use
	AliyunInstanceChargePrePaid AliyunInstanceChargeType = "PrePaid"
)

// AliyunChargePeriodType is a label for the aliyun charge unit
type AliyunChargePeriodType string

const (
	// AliyunChargePeriodWeek for week
	AliyunChargePeriodWeek AliyunChargePeriodType = "Week"
	// AliyunChargePeriodMonth for month
	AliyunChargePeriodMonth AliyunChargePeriodType = "Month"
	// AliyunChargePeriodYear for year
	AliyunChargePeriodYear AliyunChargePeriodType = "Year"
)

// AliyunInstanceChargeAttr is a describe of aliyun instance charge
type AliyunInstanceChargeAttr struct {
	PaidType        AliyunInstanceChargeType `json:"paidType,omitempty"`
	Period          int                      `json:"period,omitempty"`
	PeriodUnit      AliyunChargePeriodType   `json:"periodUnit,omitempty"`
	AutoRenew       bool                     `json:"autoRenew,omitempty"`
	AutoRenewPeriod int                      `json:"autoRenewPeriod,omitempty"`
}

// AliyunInternetChargeType is a label for the aliyun internet charge type
type AliyunInternetChargeType string

const (
	// AliyunInstanceStatusRunning for running
	AliyunInstanceStatusRunning AliyunInstanceStatusType = "Running"
	// AliyunInstanceStatusStarting for starting
	AliyunInstanceStatusStarting AliyunInstanceStatusType = "Starting"
	// AliyunInstanceStatusStopping for stopping
	AliyunInstanceStatusStopping AliyunInstanceStatusType = "Stopping"
	// AliyunInstanceStatusStopped for stopped
	AliyunInstanceStatusStopped AliyunInstanceStatusType = "Stopped"
)

const (
	// AliyunInternetChargeByTraffic for pay by traffic
	AliyunInternetChargeByTraffic AliyunInternetChargeType = "PayByTraffic"
	// AliyunInternetChargeByBandwidth for pay by bandwidth
	AliyunInternetChargeByBandwidth AliyunInstanceChargeType = "PayByBandwidth"
)

// AliyunInstanceStatusType is a label for the aliyun instance status
type AliyunInstanceStatusType string

// AliyunObjectMeta is the meta description of aliyun object
type AliyunObjectMeta struct {
	// ID is the resource id, unique, for example: i-bp1dfo67k1mxxp5obrjp, rg-aekznzdm55kjvpy
	ID string `json:"id,omitempty"`

	// Name is the display name of object, it can be changed
	Name string `json:"name,omitempty"`

	// RegionID is the Region where the object locate, for example: cn-hangzhou
	RegionID string `json:"regionID,omitempty"`

	// ResourceGroupID is the resource group which the object belong, it may be empty if the resource group is default group
	ResourceGroupID string `json:"resourceGroupID,omitempty"`
}

// AliyunImage is a description of aliyun image
type AliyunImage struct {
	AliyunObjectMeta `json:",inline"`
}

// AliyunDiskCategoryType is a label for aliyun disk category
type AliyunDiskCategoryType string

const (
	// AliyunDiskCategoryCloud for basic cloud disk, default
	AliyunDiskCategoryCloud AliyunDiskCategoryType = "cloud"
	// AliyunDiskCategoryCloudEfficiency for efficiency cloud disk
	AliyunDiskCategoryCloudEfficiency AliyunDiskCategoryType = "cloud_efficiency"
	// AliyunDiskCategoryCloudSSD for cloud ssd disk
	AliyunDiskCategoryCloudSSD AliyunDiskCategoryType = "cloud_ssd"
	// AliyunDiskCategoryEphemeralSSD for ephemeral ssd disk
	AliyunDiskCategoryEphemeralSSD AliyunDiskCategoryType = "ephemeral_ssd"
	// AliyunDiskCategoryCloudESSD for cloud essd disk
	AliyunDiskCategoryCloudESSD AliyunDiskCategoryType = "cloud_essd"
)

// AliyunDisk is a description of aliyun disk
type AliyunDisk struct {
	AliyunObjectMeta `json:",inline"`
	// Size is the size of disk, GiB
	Size               int    `json:"size,omitempty"`
	Category           string `json:"category,omitempty"`
	SnapshotID         string `json:"snapshotID,omitempty"`
	DeleteWithInstance bool   `json:"deleteWithInstance,omitempty"`
}

// AliyunVPC is a description of aliyun vpc
type AliyunVPC struct {
	AliyunObjectMeta `json:",inline"`
	CidrBlock        string `json:"cidrBlock,omitempty"`
	IsDefault        bool   `json:"isDefault,omitempty"`
	NatGatewayID     string `json:"natGatewayID,omitempty"`
}

// AliyunVSwitch is a description of aliyun vswitch
type AliyunVSwitch struct {
	AliyunObjectMeta `json:",inline"`
	VPCID            string `json:"vpcID,omitempty"`
	CidrBlock        string `json:"cidrBlock,omitempty"`
	ZoneID           string `json:"zoneID,omitempty"`
	IsDefault        bool   `json:"isDefault,omitempty"`
}

// AliyunSecurityGroup is a desription of aliyun security group
type AliyunSecurityGroup struct {
	AliyunObjectMeta `json:",inline"`
	VPCID            string `json:"vpcID,omitempty"`
}

// AliyunNetworkInterface is a description of aliyun network interface
type AliyunNetworkInterface struct {
	AliyunObjectMeta `json:",inline"`
	PrimaryIpAddress string `json:"primaryIpAddress,omitempty"`
}

// AliyunAddressType is a label for aliyun address type
type AliyunAddressType string

const (
	// AliyunAddressInternet for public
	AliyunAddressInternet AliyunAddressType = "internet"
	// AliyunAddressIntranet for private
	AliyunAddressIntranet AliyunAddressType = "intranet"
)

// AliyunLoadBalancer is a description of aliyun load balancer
type AliyunLoadBalancer struct {
	AliyunObjectMeta `json:",inline"`
	Address          string                            `json:"address,omitempty"`
	AddressType      AliyunAddressType                 `json:"addressType,omitempty"`
	VSwitch          AliyunVSwitch                     `json:"vswitch,omitempty"`
	VPC              AliyunVPC                         `json:"vpc,omitempty"`
	MasterZoneID     string                            `json:"masterZoneID,omitempty"`
	SlaveZoneID      string                            `json:"slaveZoneID,omitempty"`
	ClientToken      string                            `json:"clientToken,omitempty"`
	BackendServers   []AliyunLoadBalancerBackendServer `json:"backendServers,omitempty"`
	Listeners        []AliyunLoadBalancerListener      `json:"listeners,omitempty"`
}

// AliyunLoadBalancerBackendServer is a description of aliyun load balancer backend server
type AliyunLoadBalancerBackendServer struct {
	ServerID string `json:"serverID,omitempty"`
	Port     int    `json:"port,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Type     string `json:"type,omitempty"`
}

// AliyunLoadBalancerVServerGroup is a description of aliyun load balancer virtual server group
type AliyunLoadBalancerVServerGroup struct {
	AliyunObjectMeta `json:",inline"`
	BackendServers   []AliyunLoadBalancerBackendServer `json:"backendServers,omitempty"`
}

// AliyunLoadBalancerListener is a description of aliyun load balancer listener
type AliyunLoadBalancerListener struct {
	ListenPort        int    `json:"listenPort,omitempty"`
	BackendServerPort int    `json:"backendServerPort,omitempty"`
	Status            string `json:"status,omitempty"`
	VServerGroupID    string `json:"vserverGroupID,omitempty"`
}

// AliyunMachineEipAddress for aliyun instance eip attr
type AliyunMachineEipAddress struct {
	IPAddress          string                   `json:"ipAddress,omitempty"`
	AllocationID       string                   `json:"allocationID,omitempty"`
	InternetChargeType AliyunInternetChargeType `json:"internetChargeType,omitempty"`
}

// AliyunMachineCloudProviderConfig is a description of a aliyun ecs
type AliyunMachineCloudProviderConfig struct {
	AliyunObjectMeta        `json:",inline"`
	Image                   AliyunImage               `json:"image,omitempty"`
	InstanceType            AliyunInstanceType        `json:"instanceType,omitempty"`
	InstanceNetworkType     AliyunInstanceNetworkType `json:"instanceNetworkType,omitempty"`
	InstanceCharge          AliyunInstanceChargeAttr  `json:"instanceCharge,omitempty"`
	ZoneID                  string                    `json:"zoneID,omitempty"`
	InternetChargeType      AliyunInternetChargeType  `json:"internetChargeType,omitempty"`
	InternetMaxBandwidthIn  int                       `json:"internetMaxBandwidthIn,omitempty"`
	InternetMaxBandwidthOut int                       `json:"internetMaxBandwidthOut,omitempty"`
	CPU                     int                       `json:"cpu,omitempty"`
	Memory                  int                       `json:"memory,omitempty"`
	OSType                  string                    `json:"osType,omitempty"`
	IOOptimized             bool                      `json:"ioOptimized,omitempty"`
	KeyPairName             string                    `json:"keyPairName,omitempty"`
	ClientToken             string                    `json:"clientToken,omitempty"`
	User                    string                    `json:"user,omitempty"`
	Password                string                    `json:"password,omitempty"`
	OSDisk                  AliyunDisk                `json:"osDisk,omitempty"`
	DataDisks               []AliyunDisk              `json:"dataDisks,omitempty"`
	VPC                     AliyunVPC                 `json:"vpc,omitempty"`
	VSwitch                 AliyunVSwitch             `json:"vswitch,omitempty"`
	SecurityGroup           AliyunSecurityGroup       `json:"securityGroup,omitempty"`
	NetworkInterfaces       []AliyunNetworkInterface  `json:"networkInterfaces,omitempty"`
}

// AliyunClusterCloudProviderConfig is a description of a aliyun ecs cluster
type AliyunClusterCloudProviderConfig struct {
	RegionID string    `json:"regionID,omitempty"`
	VPC      AliyunVPC `json:"vpc,omitempty"`
	// the eip use for snat
	EIP          string              `json:"eip,omitempty"`
	LoadBalancer *AliyunLoadBalancer `json:"loadBalancer,omitempty"`
}

// azure

// AzureObjectMeta is the meta description of azure object
type AzureObjectMeta struct {
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
	// Name - Resource name.
	Name string `json:"name,omitempty"`
	// Location - Resource location.
	Location string `json:"location,omitempty"`
	// ResourceGroupName - Resource group name
	GroupName string `json:"groupName,omitempty"`
}

type AzureObjectReference = AzureObjectMeta

type AzureClusterCloudProviderConfig struct {
	// Location - cluster azure resource location.
	Location string `json:"location,omitempty"`
	// VirtualNetwork - cluster azure virtual network
	VirtualNetwork AzureVirtualNetwork `json:"virtualNetwork"`
	// Deprecated: ha cluster need more than 1 lb
	// LoadBalancer - ha cluster vip is azure lb
	LoadBalancer *AzureLoadBalancer `json:"loadBalancer,omitempty"`
	// LoadBalancers - ha cluster need 2 azure lb
	LoadBalancers []AzureLoadBalancer `json:"loadBalancers,omitempty"`
}

// AzureAksClusterCloudProviderConfig for azure aks cluster config
type AzureAksClusterCloudProviderConfig struct {
	// NetworkConfigMode - aks network user config mode
	NetworkConfigMode AzureAksNetworkConfigMode `json:"networkConfigMode"`
	// VirtualNetwork - aks virtual network
	VirtualNetwork AzureVirtualNetwork `json:"virtualNetwork"`
	// Subnet - ask node subnet
	Subnet AzureSubnet `json:"subnet"`
	// Owner - the owner name, used for hosting aks deleting
	Owner string `json:"owner,omitempty"`
	// Aks - the aks cluster config
	Aks AzureManagedCluster `json:"aks"`
}

// AzureManagedCluster is azure aks properties
type AzureManagedCluster struct {
	AzureObjectMeta
	KubernetesVersion string `json:"kubernetesVersion"`
	DNSPrefix         string `json:"dnsPrefix"`
	// Fqdn is the api server url, it will return by the azure
	Fqdn string `json:"fqdn"`
	// the agent pool profiles, only one pool is required for now
	AgentPoolProfiles []AzureManagedClusterAgentPoolProfile `json:"agentPoolProfiles"`
	LinuxProfile      AzureLinuxProfile                     `json:"linuxProfile"`
	// the azure secret, is required
	ServicePrincipalProfile AzureManagedClusterServicePrincipalProfile `json:"servicePrincipalProfile"`
	// the resource group witch hold the all the agent, it will return by the azure, and can't be configed
	NodeResourceGroup string                            `json:"nodeResourceGroup"`
	EnableRBAC        bool                              `json:"enableRBAC"`
	NetworkProfile    AzureManagedClusterNetworkProfile `json:"networkProfile"`
}

// AzureManagedClusterAgentPoolProfile for azure aks agent pool properties
type AzureManagedClusterAgentPoolProfile struct {
	Name         string      `json:"name"`
	Count        int32       `json:"count"`
	VMSize       AzureVMSize `json:"vmSize"`
	OSDisk       AzureDisk   `json:"osDisk"`
	VnetSubnetID string      `json:"vnetSubnetID,omitempty"`
	MaxPods      int32       `json:"maxPods"`
	// OSType for the agent os, only linux is available
	OSType string `json:"osType"`
}

// AzureLinuxProfile for azure aks agent linux profile
type AzureLinuxProfile struct {
	AdminUsername string                `json:"adminUsername"`
	SSH           AzureSSHConfiguration `json:"ssh"`
}

// AzureSSHConfiguration for azure ssh config
type AzureSSHConfiguration struct {
	// only one key is required for now
	PublicKeys []AzurePublicKey `json:"publicKeys"`
}

// AzurePublicKey for azure ssh key
type AzurePublicKey struct {
	KeyData string `json:"keyData"`
}

// AzureManagedClusterServicePrincipalProfile for azure aks service account profile
type AzureManagedClusterServicePrincipalProfile struct {
	ClientID string `json:"clientID,omitempty"`
	Secret   string `json:"secret,omitempty"`
}

// AzureManagedClusterNetworkProfile for azure aks cluster network profile
type AzureManagedClusterNetworkProfile struct {
	NetworkPlugin string `json:"networkPlugin"`
	// doesn't support for now
	NetworkPolicy    string `json:"networkPolicy"`
	PodCIDR          string `json:"podCIDR"`
	ServiceCIDR      string `json:"serviceCIDR"`
	DNSServiceIP     string `json:"dnsServiceIP"`
	DockerBridgeCIDR string `json:"dockerBridgeCIDR"`
}

// AzureMachineCloudProviderConfig is a description of a azure vm
type AzureMachineCloudProviderConfig struct {
	AzureObjectMeta
	VirtualNetwork    AzureVirtualNetwork     `json:"virtualNetwork"`
	LoginUser         string                  `json:"loginUser"`
	LoginPassword     string                  `json:"loginPassword"`
	VMSize            AzureVMSize             `json:"vmSize"`
	ImageReference    AzureImageReference     `json:"imageReference"`
	OSDisk            AzureDisk               `json:"osDisk"`
	DataDisks         []AzureDisk             `json:"dataDisks"`
	NetworkInterfaces []AzureNetworkInterface `json:"networkInterfaces"`
	AvailabilitySet   *AzureObjectReference   `json:"availabilitySet,omitempty"`
}

// AzureVirtualNetwork is the describe of azure virtual network
type AzureVirtualNetwork struct {
	AzureObjectMeta
}

// AzureSubnet is the describe of azure subnet
type AzureSubnet struct {
	AzureObjectMeta
}

// AzureSecurityGroup is the describe of azure security group
type AzureSecurityGroup struct {
	AzureObjectMeta
}

// AzureVMSize is the describe of azure vm size
type AzureVMSize struct {
	AzureObjectMeta
	NumberOfCores         int `json:"numberOfCores"`
	NumberOfPhysicalCores int `json:"numberOfPhysicalCores"`
	MemoryInMB            int `json:"memoryInMB"`
}

// AzureImageReference is the describe of azure image
type AzureImageReference struct {
	AzureObjectMeta
	Publisher string `json:"publisher,omitempty"`
	Offer     string `json:"offer,omitempty"`
	Sku       string `json:"sku,omitempty"`
	Version   string `json:"version,omitempty"`
}

// AzureNetworkInterface is the describe of azure network interface
type AzureNetworkInterface struct {
	AzureObjectMeta
	Primary          bool               `json:"primary"`
	SecurityGroup    AzureSecurityGroup `json:"securityGroup"`
	IPConfigurations []AzureIpConfig    `json:"ipConfigurations"`
}

// AzurePublicIP is the describe of azure public ip
type AzurePublicIP struct {
	AzureObjectMeta
	PublicIPAddress string `json:"publicIPAddress"`
}

// AzureIpConfig is the describe of azure public ip config
type AzureIpConfig struct {
	AzureObjectMeta
	Primary          bool           `json:"primary,omitempty"`
	Subnet           AzureSubnet    `json:"subnet"`
	PrivateIPAddress string         `json:"privateIPAddress"`
	PublicIP         *AzurePublicIP `json:"publicIPAddress,omitempty"`
}

// AzureDisk is the describe of azure disk object
type AzureDisk struct {
	AzureObjectMeta
	// Lun - Specifies the logical unit number of the data disk. This value is used to identify data disks within the VM and therefore must be unique for each data disk attached to a VM.
	// This value should be [0-63]
	Lun int32 `json:"lun"`
	// CreateOption - Specifies how the virtual machine should be created.<br><br> Possible values are:<br><br> **Attach** \u2013 This value is used when you are using a specialized disk to create the virtual machine.<br><br> **FromImage** \u2013 This value is used when you are using an image to create the virtual machine. If you are using a platform image, you also use the imageReference element described above. If you are using a marketplace image, you  also use the plan element previously described. Possible values include: 'DiskCreateOptionTypesFromImage', 'DiskCreateOptionTypesEmpty', 'DiskCreateOptionTypesAttach'
	CreateOption string `json:"createOption"`
	SizeGB       int32  `json:"sizeGB"`
	SkuName      string `json:"skuName"`         // ssd/hdd and theirs upper type
	Owner        string `json:"owner,omitempty"` // when cleanup, only controller created can be delete
}

type AzureAvailabilitySet struct { // for future? not used now
	AzureObjectMeta
	// PlatformUpdateDomainCount - Update Domain count.
	PlatformUpdateDomainCount int32 `json:"platformUpdateDomainCount,omitempty"`
	// PlatformFaultDomainCount - Fault Domain count.
	PlatformFaultDomainCount int32 `json:"platformFaultDomainCount,omitempty"`
	// VirtualMachines - A list of references to all virtual machines in the availability set.
	VirtualMachines []AzureObjectReference `json:"virtualMachines"`
	// SkuName - Sku name of the availability set, only name is required to be set. See AvailabilitySetSkuTypes for possible set of values. Use 'Aligned' for virtual machines with managed disks and 'Classic' for virtual machines with unmanaged disks. Default value is 'Classic'.
	SkuName string `json:"skuName"`
}

type AzureFrontendIPConfiguration struct {
	AzureObjectMeta
	// PrivateIPAddress - The private IP address of the IP configuration.
	PrivateIPAddress string `json:"privateIPAddress,omitempty"`
	// PublicIP - The reference of the Public IP resource.
	PublicIP *AzurePublicIP `json:"publicIP,omitempty"`
	// Subnet - The reference of the subnet resource.
	Subnet AzureSubnet `json:"subnet,omitempty"`
}

type AzureBackendAddressPool struct {
	AzureObjectMeta
	// BackendIPConfigurations - Gets collection of references to IP addresses defined in network interfaces.
	BackendIPConfigurations []AzureIpConfig `json:"backendIPConfigurations"`
}

type AzureLoadBalancerProbe struct {
	AzureObjectMeta
	// Protocol - The protocol of the end point. Possible values are: 'Http' or 'Tcp'. If 'Tcp' is specified, a received ACK is required for the probe to be successful. If 'Http' is specified, a 200 OK response from the specifies URI is required for the probe to be successful. Possible values include: 'ProbeProtocolHTTP', 'ProbeProtocolTCP'
	Protocol string `json:"protocol,omitempty"`
	// Port - The port for communicating the probe. Possible values range from 1 to 65535, inclusive.
	Port int32 `json:"port,omitempty"`
	// IntervalInSeconds - The interval, in seconds, for how frequently to probe the endpoint for health status. Typically, the interval is slightly less than half the allocated timeout period (in seconds) which allows two full probes before taking the instance out of rotation. The default value is 15, the minimum value is 5.
	IntervalInSeconds int32 `json:"intervalInSeconds,omitempty"`
	// NumberOfProbes - The number of probes where if no response, will result in stopping further traffic from being delivered to the endpoint. This values allows endpoints to be taken out of rotation faster or slower than the typical times used in Azure.
	NumberOfProbes int32 `json:"numberOfProbes,omitempty"`
}

type AzureLoadBalancingRule struct {
	AzureObjectMeta
	// Protocol - Possible values include: 'TransportProtocolUDP', 'TransportProtocolTCP', 'TransportProtocolAll'
	Protocol string `json:"protocol,omitempty"`
	// LoadDistribution - The load distribution policy for this rule. Possible values are 'Default', 'SourceIP', and 'SourceIPProtocol'. Possible values include: 'Default', 'SourceIP', 'SourceIPProtocol'
	LoadDistribution string `json:"loadDistribution,omitempty"`
	// FrontendPort - The port for the external endpoint. Port numbers for each rule must be unique within the Load Balancer. Acceptable values are between 0 and 65534. Note that value 0 enables "Any Port"
	FrontendPort int32 `json:"frontendPort,omitempty"`
	// BackendPort - The port used for internal connections on the endpoint. Acceptable values are between 0 and 65535. Note that value 0 enables "Any Port"
	BackendPort int32 `json:"backendPort,omitempty"`
	// IdleTimeoutInMinutes - The timeout for the TCP idle connection. The value can be set between 4 and 30 minutes. The default value is 4 minutes. This element is only used when the protocol is set to TCP.
	IdleTimeoutInMinutes int32 `json:"idleTimeoutInMinutes,omitempty"`
	// EnableFloatingIP - Configures a virtual machine's endpoint for the floating IP capability required to configure a SQL AlwaysOn Availability Group. This setting is required when using the SQL AlwaysOn Availability Groups in SQL server. This setting can't be changed after you create the endpoint.
	EnableFloatingIP bool `json:"enableFloatingIP,omitempty"`
	// DisableOutboundSnat - Configures SNAT for the VMs in the backend pool to use the publicIP address specified in the frontend of the load balancing rule.
	DisableOutboundSnat bool `json:"disableOutboundSnat,omitempty"`
	// ProvisioningState - Gets the provisioning state of the PublicIP resource. Possible values are: 'Updating', 'Deleting', and 'Failed'.
	ProvisioningState string `json:"provisioningState,omitempty"`
}

type AzureLoadBalancer struct {
	AzureObjectMeta
	// SkuName - Name of a load balancer SKU. Possible values include: 'LoadBalancerSkuNameBasic', 'LoadBalancerSkuNameStandard'
	SkuName string `json:"skuName"`
	// FrontendIPConfigurations - Object representing the frontend IPs to be used for the load balancer
	FrontendIPConfigurations []AzureFrontendIPConfiguration `json:"frontendIPConfigurations"`
	// BackendAddressPools - Collection of backend address pools used by a load balancer
	BackendAddressPools []AzureBackendAddressPool `json:"backendAddressPools"`
	// Probes - Collection of probe objects used in the load balancer
	Probes []AzureLoadBalancerProbe `json:"probes"`
	// LoadBalancingRules - Object collection representing the load balancing rules Gets the provisioning
	LoadBalancingRules []AzureLoadBalancingRule `json:"loadBalancingRules,omitempty"`
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
	Type         NetworkType       `json:"type"`
	ClusterCIDR  string            `json:"clusterCIDR"`
	ServiceCIDR  string            `json:"serviceCIDR"`
	DNSServiceIP string            `json:"dnsServiceIP"`
	Default      NetworkTemplate   `json:"default"`
	Extras       []NetworkTemplate `json:"extras,omitempty"`
}

type NetworkTemplate struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NetworkSpec `json:"spec"`
}

type ClusterAuth struct {
	KubeUser     string `json:"kubeUser,omitempty"`
	KubePassword string `json:"kubePassword,omitempty"`
	KubeToken    string `json:"kubeToken,omitempty"`
	KubeCertPath string `json:"kubeCertPath,omitempty"`
	KubeCAData   []byte `json:"kubeCAData,omitempty"`
	EndpointIP   string `json:"endpointIP,omitempty"`
	EndpointPort string `json:"endpointPort,omitempty"`

	KubeConfig *clientcmdapi.Config `json:"kubeConfig,omitempty"`
}

type ClusterVersions struct {
	// ClusterSets save the versions of component that run in cluster scope
	//   like kube version ...
	ClusterSets map[string]string `json:"clusterSets"`
	// MasterSets save the versions of component that run in masters
	//   like apiserver ...
	MasterSets map[string]string `json:"masterSets"`
	// MasterSets save the versions of component that run in all nodes
	//   like kubelet, kube-proxy
	NodeSets MachineVersions `json:"nodeSets"`
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

// AzureAksNetworkConfigMode for the user network config mode in aks
type AzureAksNetworkConfigMode string

type CloudProvider string

type ClusterPhase string
type MachinePhase string
type PodPhase string
type MASGPhase string // machine auto scaling group phase
type InfraNetworkPhase string

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
	DeviceType string `json:"deviceType"`
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

// auto scaling

// ClusterScaleUpSetting describe cluster scale up setting
type ClusterScaleUpSetting struct {
	Algorithm            string `json:"algorithm"`
	IsQuotaUpdateEnabled bool   `json:"isQuotaUpdateEnabled"`
	// Deprecated: the scale-up has been switched to pre-scheduling, no longer need the cooldown time anymore.
	// cool down time after any cluster scale up action, waiting for pods schedule
	CoolDown metav1.Duration `json:"coolDown"`
}

// ClusterScaleDownSetting describe cluster scale down setting
type ClusterScaleDownSetting struct {
	// is scale down enabled
	IsEnabled bool `json:"isEnabled"`
	// cool down time after any cluster scale up action
	CoolDown metav1.Duration `json:"coolDown"`
	// machine continues idle time threshold
	IdleTime metav1.Duration `json:"idleTime"`
	// machine idle threshold, percent of cpu/mem usage
	IdleThreshold int `json:"idleThreshold"`
}

// AutoScalingNotifySetting describe notify about setting
type AutoScalingNotifySetting struct {
	// notify methods
	Methods []string `json:"methods"`
	// notify group ids
	// is int in notify-admin, but in case of it changes, use string
	Groups []string `json:"groups"`
}

// ClusterAutoScalingSetting describe a cluster auto scaling setting
// maybe nil in old or control cluster, need to be inited with a default setting by controller
type ClusterAutoScalingSetting struct {
	// scale up setting
	ScaleUpSetting ClusterScaleUpSetting `json:"scaleUpSetting"`
	// scale down setting
	ScaleDownSetting ClusterScaleDownSetting `json:"scaleDownSetting"`
	// cluster level warning message notify setting
	NotifySetting AutoScalingNotifySetting `json:"notifySetting"`
}

// ClusterAutoScalingStatus describe cluster auto scaling operate status
type ClusterAutoScalingStatus struct {
	// last scale up operation time
	LastScaleUpTime metav1.Time `json:"lastScaleUpTime,omitempty"`
	// last selected scale up group name
	LastScaleUpGroup string `json:"lastScaleUpGroup,omitempty"`
	// last scale down operation time
	LastScaleDownTime metav1.Time `json:"lastScaleDownTime,omitempty"`
	// last selected scale down group name
	LastScaleDownGroup string `json:"lastScaleDownGroup,omitempty"`
}

// ClusterProviderStatus describe cluster cloud provider operate status
type ClusterProviderStatus struct {
	AzureAks *AzureAksClusterCloudProviderStatus `json:"azureAks,omitempty"`
}

// AzureAksClusterCloudProviderStatus for azure aks status
type AzureAksClusterCloudProviderStatus struct {
	// the aks state get from azure
	ProvisioningState string `json:"provisioningState,omitempty"`
	// AgentPoolCount - the agent count, use one value because there is just one pool available for now
	AgentPoolCount int32 `json:"agentPoolCount"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineAutoScalingGroup describe a machine auto scaling group
type MachineAutoScalingGroup struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   MASGSpec   `json:"spec"`
	Status MASGStatus `json:"status"`
}

// MASGSpec describe MachineAutoScalingGroup spec
type MASGSpec struct {
	// is this auto scaling group enabled
	IsEnabled bool `json:"isEnabled"`
	// machine auto scaling group cloud provider
	Provider CloudProvider `json:"provider"`
	// machine auto scaling group cloud provider config
	ProviderConfig MASGProviderConfig `json:"providerConfig"`
	// tags of scaled machine
	Tags map[string]string `json:"tags"`
	// cluster name which group belongs to
	Cluster string `json:"cluster"`
	// group min machine num
	MinNum int `json:"minNum"`
	// group max machine num
	MaxNum int `json:"maxNum"`
	// group level warning message notify setting
	NotifySetting AutoScalingNotifySetting `json:"notifySetting"`

	// Taints is the taint list of machine. It will been copied to the machine created by the MASG.
	Taints []MachineTaint `json:"taints,omitempty"`
}

// MASGProviderConfig describe MachineAutoScalingGroup provider config
type MASGProviderConfig struct {
	Azure  *AzureMASGProviderConfig  `json:"azure,omitempty"`
	Aliyun *AliyunMASGProviderConifg `json:"aliyun,omitempty"`
}

// AliyunMASGProviderConifg is description for aliyun scaling template
type AliyunMASGProviderConifg struct {
	AliyunObjectMeta `json:",inline"`
	VPC              AliyunVPC           `json:"vpc,omitempty"`
	VSwitch          AliyunVSwitch       `json:"vswitch,omitempty"`
	LoginUser        string              `json:"loginUser,omitempty"`
	LoginPassword    string              `json:"loginPassword,omitempty"`
	InstanceType     AliyunInstanceType  `json:"instanceType,omitempty"`
	Image            AliyunImage         `json:"image,omitempty"`
	OSDisk           AliyunDisk          `json:"osDisk,omitempty"`
	DataDisks        []AliyunDisk        `json:"dataDisks,omitempty"`
	SecurityGroup    AliyunSecurityGroup `json:"securityGroup,omitempty"`
	CPU              int                 `json:"cpu,omitempty"`
	Memory           int                 `json:"memory,omitempty"`
	ZoneID           string              `json:"zoneID,omitempty"`
}

// AzureMASGProviderConfig describe machine auto scaling group provider config for azure
// similar with AzureMachineCloudProviderConfig, but no nic inside
type AzureMASGProviderConfig struct {
	AzureObjectMeta
	VirtualNetwork AzureVirtualNetwork `json:"virtualNetwork"`
	Subnet         AzureSubnet         `json:"subnet"`
	LoginUser      string              `json:"loginUser"`
	LoginPassword  string              `json:"loginPassword"`
	VMSize         AzureVMSize         `json:"vmSize"`
	ImageReference AzureImageReference `json:"imageReference"`
	OSDisk         AzureDisk           `json:"osDisk"`
	DataDisks      []AzureDisk         `json:"dataDisks"`
	SecurityGroup  AzureSecurityGroup  `json:"securityGroup"`
}

// MASGStatus describe MachineAutoScalingGroup status
type MASGStatus struct {
	// machine auto scaling group status phase
	Phase MASGPhase `json:"phase"`
	// info of machines belong to this group
	Machines []MASGMachineInfo `json:"machines"`
	// last scale up operation time
	LastScaleUpTime metav1.Time `json:"lastScaleUpTime,omitempty"`
	// last selected scale up machine name
	LastScaleUpMachine string `json:"lastScaleUpMachine,omitempty"`
	// last scale down operation time
	LastScaleDownTime metav1.Time `json:"lastScaleDownTime,omitempty"`
	// last selected scale down machine name
	LastScaleDownMachine string `json:"lastScaleDownMachine,omitempty"`
}

// MASGMachineInfo saves info of machine which belongs to this MachineAutoScalingGroup
type MASGMachineInfo struct {
	// name of related machine
	Name string `json:"name"`

	// machine provider config
	ProviderConfig MASGMachineProviderConfig `json:"providerConfig"`

	// scaling up about

	// timestamp when vm created
	// if nil or 0, means machine not created yet
	CreatedTime *metav1.Time `json:"createdTime,omitempty"`
	// timestamp when vm bounded to cluster
	// if nil or 0, means machine not bound to cluster yet
	BoundTime *metav1.Time `json:"boundTime,omitempty"`

	// timestamp when vm ready in cluster
	// if nil or 0, means machine not ready to cluster yet
	ReadyTime *metav1.Time `json:"readyTime,omitempty"`
	// timestamp when vm failed
	// if nil or 0, means machine not failed yet
	FailedTime *metav1.Time `json:"failedTime,omitempty"`
	// scaling down about

	// lastBusyTime mark the last time when enough pods run on this machine
	// ignore if nil or not greater than boundTime
	LastBusyTime *metav1.Time `json:"lastBusyTime,omitempty"`

	// timestamp when set unbound
	// if nil or 0, means machine still bound, creating or failed
	UnboundTime *metav1.Time `json:"unboundTime,omitempty"`
}

// MASGMachineProviderConfig saves inited provider config of machine in auto scaling group
type MASGMachineProviderConfig struct {
	// provider config for azure
	Azure *MASGMachineAzureProviderConfig `json:"azure,omitempty"`
}

// MASGMachineAzureProviderConfig is MASGMachineProviderConfig in azure
type MASGMachineAzureProviderConfig struct {
	// azure machine object name is generated by vm resource group and name, so we need save vm name first
	VMName string `json:"vmName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MachineAutoScalingGroupList is a collection of machine auto scaling groups
type MachineAutoScalingGroupList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of MachineAutoScalingGroups
	Items []MachineAutoScalingGroup `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfraNetwork describe a real exist infrastructure level network
type InfraNetwork struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   InfraNetworkSpec   `json:"spec"`
	Status InfraNetworkStatus `json:"status"`
}

type InfraNetworkSpec struct {
	// IsControl mark is control cluster inside
	IsControl bool `json:"isControl"`
	// Provider describe cloud provider type of this infra network
	Provider CloudProvider `json:"provider"`
	// ProviderConfig saves cloud provider detail config
	ProviderConfig InfraNetworkProviderConfig `json:"providerConfig"`
}

type InfraNetworkProviderConfig struct {
	// bare metal config
	BareMetal *InfraNetworkBareMetalProviderConfig `json:"bareMetal,omitempty"`
	// azure config
	Azure *InfraNetworkAzureProviderConfig `json:"azure,omitempty"`
}

type InfraNetworkBareMetalProviderConfig struct {
	// simple cidr
	CIDRs []string `json:"cidrs"`
}

type InfraNetworkAzureProviderConfig struct {
	// simple virtual network reference
	VirtualNetwork AzureVirtualNetwork `json:"virtualNetwork"`
}

type InfraNetworkStatus struct {
	// calculated CIDRs
	CIDRs []string `json:"cidrs"`
	// status phase of this infra network, describe if it is in use
	Phase InfraNetworkPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfraNetworkList is the list result of infra network
type InfraNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of InfraNetworks
	Items []InfraNetwork `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// oem for SUNNY-62

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterResourceScale describe a scale cluster node & partition resource information
type ClusterResourceScale struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ClusterResourceScaleSpec   `json:"spec"`
	Status ClusterResourceScaleStatus `json:"status"`
}

type ClusterResourceScaleSpec struct {
	Cluster string              `json:"cluster"`
	Machine ClusterMachineScale `json:"machine"`
	Quota   ClusterQuotaScale   `json:"quota"`
}

type ClusterMachineScale struct {
	// Provider describe the scale machine cloud information,
	// and this will not equal machine.spec.Provider
	Provider CloudProvider `json:"provider"`

	// Addresses describe the ip of every machine,
	// every item respresent a scale machine.
	Addresses []NodeAddress `json:"addresses"`

	SSHPort string `json:"sshPort"`

	Auth MachineAuth       `json:"auth"`
	Tags map[string]string `json:"tags"`
}

type ClusterQuotaScale struct {
	Quotas []ClusterQuotaScaleItem `json:"quotas"`
}

type ClusterQuotaScaleItem struct {
	Tenant    string              `json:"tenant"`
	Partition string              `json:"partition"`
	Quota     corev1.ResourceList `json:"quota"`
}

type ClusterResourceScalePhase = string

const (
	ClusterResourceScalePhasePending ClusterResourceScalePhase = "Pending"
	ClusterResourceScalePhaseScaling ClusterResourceScalePhase = "Scaling"
	ClusterResourceScalePhaseDone    ClusterResourceScalePhase = "Done"
)

type ClusterResourceScaleStatus struct {
	Phase ClusterResourceScalePhase `json:"phase"`

	Machines []ClusterMachineScaleStatusItem `json:"machines"`
	Quotas   []ClusterQuotaScaleStatusItem   `json:"quotas"`

	Conditions []ClusterResourceScaleCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
}

type ClusterMachineScaleStatusItem struct {
	Name        string       `json:"name"`
	Phase       MachinePhase `json:"phase"`
	CreatedTime *metav1.Time `json:"createdTime,omitempty"`
	BoundTime   *metav1.Time `json:"boundTime,omitempty"`
	ReadyTime   *metav1.Time `json:"readyTime,omitempty"`
	FailedTime  *metav1.Time `json:"failedTime,omitempty"`
}

type ClusterQuotaScaleStatusItem struct {
	Tenant      string       `json:"tenant"`
	Partition   string       `json:"partition"`
	UpdatedTime *metav1.Time `json:"updatedTime,omitempty"`
}

type ClusterResourceScaleConditionType = string

// ClusterResourceScaleCondition contains details for the current condition of this cluster-resource-scale.
type ClusterResourceScaleCondition struct {
	// Type is the type of the condition.
	// Currently only Ready.
	Type ClusterResourceScaleConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=MachineConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterResourceScaleList is the list result of cluster-resource-scale
type ClusterResourceScaleList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of ClusterResourceScale
	Items []ClusterResourceScale `json:"items" protobuf:"bytes,2,rep,name=items"`
}
