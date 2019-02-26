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

	Spec   ClusterSpec   `json:"spec"`
	Status ClusterStatus `json:"status"`
}

type ClusterSpec struct {
	DisplayName      string                     `json:"displayName"`
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
	// deploy
	DeployToolsExternalVars map[string]string `json:"deployToolsExternalVars"`
	// adapt expired
	ClusterToken string       `json:"clusterToken"`
	Ratio        ClusterRatio `json:"ratio"`
}

type ClusterStatus struct {
	Phase          ClusterPhase                       `json:"phase"`
	Conditions     []ClusterCondition                 `json:"conditions"`
	Masters        []MachineThumbnail                 `json:"masters"`
	Nodes          []MachineThumbnail                 `json:"nodes"`
	Capacity       map[ResourceName]resource.Quantity `json:"capacity"`
	OperationLogs  []OperationLog                     `json:"operationLogs,omitempty"`
	AutoScaling    ClusterAutoScalingStatus           `json:"autoScaling,omitempty"`
	ProviderStatus ClusterProviderStatus              `json:"providerStatus,omitempty"`
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

type MachineSpec struct {
	Provider         CloudProvider              `json:"provider"`
	ProviderConfig   MachineCloudProviderConfig `json:"providerConfig,omitempty"`
	Address          []NodeAddress              `json:"address"`
	SshPort          string                     `json:"sshPort"`
	Auth             MachineAuth                `json:"auth"`
	Versions         MachineVersions            `json:"versions,omitempty"`
	Cluster          string                     `json:"cluster"`
	IsMaster         bool                       `json:"isMaster"`
	HostnameReadonly bool                       `json:"hostnameReadonly,omitempty"`
	Tags             map[string]string          `json:"tags"`
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

	// Current service state of machine.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []MachineCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,2,rep,name=conditions"`
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

// cloud provider

type MachineCloudProviderConfig struct {
	// auto scaling

	// auto scaling group name which this machine belongs to
	// empty means not belongs to any auto scaling group
	AutoScalingGroup string `json:"autoScalingGroup,omitempty"`

	Azure *AzureMachineCloudProviderConfig `json:"azure,omitempty"`
}

type ClusterCloudProviderConfig struct {
	// auto scaling

	// cluster level auto scaling setting
	// maybe nil in old or control cluster, need to be inited with a default setting by controller
	AutoScalingSetting *ClusterAutoScalingSetting `json:"autoScalingSetting,omitempty"`

	Azure    *AzureClusterCloudProviderConfig    `json:"azure,omitempty"`
	AzureAks *AzureAksClusterCloudProviderConfig `json:"azureAks,omitempty"`
}

// azure

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
	// LoadBalancer - ha cluster vip is azure lb
	LoadBalancer *AzureLoadBalancer `json:"loadBalancer,omitempty"`
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
	VnetSubnetID string      `json:"vnetSubnetID,omitempty`
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

type AzureVirtualNetwork struct {
	AzureObjectMeta
}

type AzureSubnet struct {
	AzureObjectMeta
}

type AzureSecurityGroup struct {
	AzureObjectMeta
}

type AzureVMSize struct {
	AzureObjectMeta
}

type AzureImageReference struct {
	AzureObjectMeta
	Publisher string `json:"publisher,omitempty"`
	Offer     string `json:"offer,omitempty"`
	Sku       string `json:"sku,omitempty"`
	Version   string `json:"version,omitempty"`
}

type AzureNetworkInterface struct {
	AzureObjectMeta
	Primary          bool               `json:"primary"`
	SecurityGroup    AzureSecurityGroup `json:"securityGroup"`
	IPConfigurations []AzureIpConfig    `json:"ipConfigurations"`
}

type AzurePublicIP struct {
	AzureObjectMeta
	PublicIPAddress string `json:"publicIPAddress"`
}

type AzureIpConfig struct {
	AzureObjectMeta
	Primary          bool           `json:"primary,omitempty"`
	Subnet           AzureSubnet    `json:"subnet"`
	PrivateIPAddress string         `json:"privateIPAddress"`
	PublicIP         *AzurePublicIP `json:"publicIPAddress,omitempty"`
}

type AzureDisk struct {
	AzureObjectMeta
	SizeGB  int32  `json:"sizeGB"`
	SkuName string `json:"skuName"`         // ssd/hdd and theirs upper type
	Owner   string `json:"owner,omitempty"` // when cleanup, only controller created can be delete
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
	Type        NetworkType `json:"type"`
	ClusterCIDR string      `json:"clusterCIDR"`
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

// AzureAksNetworkConfigMode for the user network config mode in aks
type AzureAksNetworkConfigMode string

type CloudProvider string

type ClusterPhase string
type MachinePhase string
type PodPhase string
type MASGPhase string // machine auto scaling group phase

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
}

// MASGProviderConfig describe MachineAutoScalingGroup provider config
type MASGProviderConfig struct {
	Azure *AzureMASGProviderConfig `json:"azure"`
}

// MASGProviderAzureConfig describe machine auto scaling group provider config for azure
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
