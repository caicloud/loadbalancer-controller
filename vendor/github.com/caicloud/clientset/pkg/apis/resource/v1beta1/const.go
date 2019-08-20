package v1beta1

// cluster/machine status

const (
	// cluster install status
	ClusterStatusNew           ClusterPhase = "New"
	ClusterStatusInstallMaster ClusterPhase = "InstallMaster"
	ClusterStatusInstallAddon  ClusterPhase = "InstallAddon"
	ClusterStatusReady         ClusterPhase = "Ready"
	ClusterStatusFailed        ClusterPhase = "Failed"
	ClusterStatusDeleting      ClusterPhase = "Deleting"
	// cluster virtual status
	ClusterStatusNotReady ClusterPhase = "NotReady"

	// machine install status
	MachineStatusNew        MachinePhase = "New"
	MachineStatusInited     MachinePhase = "Inited"
	MachineStatusBinding    MachinePhase = "Binding"
	MachineStatusReady      MachinePhase = "Ready"
	MachineStatusUnbinding  MachinePhase = "Unbinding"
	MachineStatusSetOffline MachinePhase = "SetOffline"
	MachineStatusOffline    MachinePhase = "Offline"
	MachineStatusSetOnline  MachinePhase = "SetOnline"
	MachineStatusRemoving   MachinePhase = "Removing"
	MachineStatusFailed     MachinePhase = "Failed"
	// machine virtual status
	MachineStatusNodeNotReady MachinePhase = "NodeNotReady"
	MachineStatusFree         MachinePhase = "Free"
	MachineStatusAllocate     MachinePhase = "Allocate"

	// machine auto scaling group status
	MASGStatusEnabled  MASGPhase = "Enabled"
	MASGStatusDisabled MASGPhase = "Disabled"
	MASGStatusRemoving MASGPhase = "Removing"

	// machine delete about

	// MachineDeletePolicyNone for only delete from compass
	MachineDeletePolicyNone = "none"
	// MachineDeletePolicyInstance for delete vm(with interface/osdisk) from cloud
	MachineDeletePolicyInstance = "instance"
	// MachineDeletePolicyAll for delete all resource(with datadisk) from cloud
	MachineDeletePolicyAll = "all"
)

// cloud provider about

const (
	// cloud provider
	ProviderBaremetal CloudProvider = "caicloud-baremetal"
	ProviderAnchnet   CloudProvider = "caicloud-anchnet"
	ProviderAnsible   CloudProvider = "caicloud-ansible"
	ProviderAliyun    CloudProvider = "caicloud-aliyun"
	ProviderFake      CloudProvider = "caicloud-fake"
	ProviderAzure     CloudProvider = "caicloud-azure"
	ProviderAzureAks  CloudProvider = "caicloud-azure-aks"
)

// cluster/machine/node about annotations & labels about

const (
	// AnnotationKeyMachineDeletePolicy for delete-policy annotation
	AnnotationKeyMachineDeletePolicy = "machine.resource.caicloud.io/delete-policy"

	// AnnotationKeyAksLastResizeRequestTime record the aks last resize req time, when the request has been sync to azure, the key should be deleted(or set empty)
	AnnotationKeyAksLastResizeRequestTime    = "cluster.resource.caicloud.io/aks-last-resize-request-time"
	AnnotationKeyAksPrevNetworkSecurityGroup = "cluster.resource.caicloud.io/aks-prev-network-security-group"
	AnnotationKeyAksPrevRouteTable           = "cluster.resource.caicloud.io/aks-prev-route-table"

	AnnotationKeyAlias       = "resource.caicloud.io/alias"
	AnnotationKeyDescription = "resource.caicloud.io/description"

	// node annotations
	//   node from exist k8s like aks may not have these
	// node reference
	AnnotationKeyNodeMachine = "reference.caicloud.io/machine"
	AnnotationKeyNodeCluster = "reference.caicloud.io/cluster"
	// node machine status
	AnnotationKeyNodeMachineStatus = "machine.resource.caicloud.io/status"
	// node machine info
	AnnotationKeyNodeMachineIsMaster = "machine.resource.caicloud.io/is-master"

	// AnnotationKeyNodeMachineIsEtcd for etcd node annotation key, it's value is true when node is etcd node
	AnnotationKeyNodeMachineIsEtcd = "machine.resource.caicloud.io/is-etcd"

	AnnotationKeyNodeMachineDiskTotal = "machine.resource.caicloud.io/disk-total"
	AnnotationKeyNodeMASG             = "machine.resource.caicloud.io/auto-scaling-group"
	// node machine cloud provider about
	AnnotationKeyNodeMachineCloudProvider    = "machine.resource.caicloud.io/cloud-provider"
	AnnotationKeyNodeMachineVirtualMachineID = "machine.resource.caicloud.io/virtual-machine-id"
	// node affinity system prefix
	LabelKeyNodeSystemPrefixKube     = "kubernetes.io"
	LabelKeyNodeSystemPrefixCaicloud = "caicloud.io"
	LabelKeyNodeSystemPrefixAzure    = "azure.com"

	// finalizers
	FinalizerNameClusterController = "cluster.resource.caicloud.io/cluster-finalizer"

	// TaintsKeyNodeEtcd represents the value of etcd node taints key
	TaintsKeyNodeEtcd = "cluster.resource.caicloud.io/etcd"
)

// cluster/machine/node conditions about

const (
	// cdt type
	// cluster
	ClusterConditionNodeNumMismatch ClusterConditionType = "NodeNumMismatch"
	ClusterConditionMastersReady    ClusterConditionType = "MasterReady"
	ClusterConditionApiServerReady  ClusterConditionType = "ApiServerReady"

	// ClusterConditionEtcdsReady represent the cluster etcd nodes status
	ClusterConditionEtcdsReady      ClusterConditionType = "EtcdReady"
	ClusterConditionAddonReady      ClusterConditionType = "AddonReady"
	ClusterConditionAddonTimeout    ClusterConditionType = "AddonTimeout"
	ClusterConditionDeletion        ClusterConditionType = "Deletion"
	ClusterConditionInstallProgress ClusterConditionType = "InstallProgress"
	ClusterConditionProviderFailure ClusterConditionType = "ProviderFailure"
	ClusterConditionCCManagerWorks  ClusterConditionType = "CloudControllerManagerWorks"
	// machine
	MachineConditionFailed MachineConditionType = "Failed"

	// cdt reason
	ConditionReasonPrefix = "ConditionReason:"
	// cluster
	ClusterConditionReasonClusterCreated      = ConditionReasonPrefix + "ClusterCreated"
	ClusterConditionReasonInstallRetry        = ConditionReasonPrefix + "InstallRetry"
	ClusterConditionReasonMasterMarkDone      = ConditionReasonPrefix + "MasterMarkDone"
	ClusterConditionReasonMasterMarkFailed    = ConditionReasonPrefix + "MasterMarkFailed"
	ClusterConditionReasonMasterInstallDone   = ConditionReasonPrefix + "MasterInstallDone"
	ClusterConditionReasonMasterInstallFailed = ConditionReasonPrefix + "MasterInstallFailed"
	ClusterConditionReasonAddonInstallFailed  = ConditionReasonPrefix + "AddonInstallFailed"
	ClusterConditionReasonAddonInstallDone    = ConditionReasonPrefix + "AddonInstallDone"
	ClusterConditionReasonUserOperation       = ConditionReasonPrefix + "UserOperation"
	ClusterConditionReasonControllerCheck     = ConditionReasonPrefix + "ControllerCheck"
	ClusterConditionReasonResourceNotFound    = ConditionReasonPrefix + "ResourceNotFound"
	// machine
	MachineConditionReasonNew        = ConditionReasonPrefix + "MachineNew"
	MachineConditionReasonBinding    = ConditionReasonPrefix + "MachineBinding"
	MachineConditionReasonUnbinding  = ConditionReasonPrefix + "MachineUnbinding"
	MachineConditionReasonRemoving   = ConditionReasonPrefix + "MachineRemoving"
	MachineConditionReasonSetOffline = ConditionReasonPrefix + "SetOffline"
	MachineConditionReasonSetOnline  = ConditionReasonPrefix + "SetOnline"
)

// storage provisioner

type StorageClassProvisioner = string

const (
	StorageClassProvisionerGlusterfs StorageClassProvisioner = "kubernetes.io/glusterfs"
	StorageClassProvisionerCephRBD   StorageClassProvisioner = "kubernetes.io/rbd"
	StorageClassProvisionerCephfs    StorageClassProvisioner = "ceph.com/cephfs"
	StorageClassProvisionerAzureDisk StorageClassProvisioner = "kubernetes.io/azure-disk"
	StorageClassProvisionerAzureFile StorageClassProvisioner = "kubernetes.io/azure-file"
	StorageClassProvisionerNetappNAS StorageClassProvisioner = "netapp.io/trident"
	StorageClassProvisionerNFS       StorageClassProvisioner = "caicloud.io/nfs"
	StorageClassProvisionerLocal     StorageClassProvisioner = "caicloud.io/local-storage"
)

// storage annotations and labels

const (
	// storage class annotations
	AnnotationKeyStorageType         = "storage.resource.caicloud.io/type"
	AnnotationKeyStorageService      = "storage.resource.caicloud.io/service"
	AnnotationKeyStorageClass        = "storage.resource.caicloud.io/class"
	AnnotationKeyStorageAdminMarkKey = "storage.resource.caicloud.io/process"
	AnnotationKeyStorageAdminMarkVal = "true"
	AnnotationKeyIsSystemKey         = "storage.resource.caicloud.io/system"
	AnnotationKeyIsSystemVal         = "true"

	// annotations for all
	AnnotationKeyStorageAdminAlias       = "storage.resource.caicloud.io/alias"
	AnnotationKeyStorageAdminDescription = "storage.resource.caicloud.io/description"

	// local storage refer on node
	AnnotationKeyNodeRelatedLocalStorageClassNames = "storage.resource.caicloud.io/node-local-storage-class-names"
	AnnotationKeyNodeRelatedLocalStorageClasses    = "storage.resource.caicloud.io/node-local-storage-classes"

	// local storage related
	// AnnotationKeyLocalStorageDisks is the key that stores disk information in the StorageClass annotation.
	AnnotationKeyLocalStorageDisks string = "storage.resource.caicloud.io/local-storage-disks"
	// AnnotationKeyNodeUseableDisks is the key that stores useable disks information in the Node annotation.
	AnnotationKeyNodeUseableDisks string = "storage.resource.caicloud.io/useable-disks"

	// pvc annotations
	AnnotationKeyAppIDs              = "storage.resource.caicloud.io/app-ids"
	AnnotationKeyAppNames            = "storage.resource.caicloud.io/app-names"
	AnnotationKeyPvcIsMountable      = "storage.resource.caicloud.io/mountable"
	AnnotationKeyPVCRollbackPhase    = "storage.resource.caicloud.io/rollback-phase"
	AnnotationKeyRelatedClusters     = "storage.resource.caicloud.io/related-clusters"
	AnnotationKeyPVCEventMessage     = "storage.resource.caicloud.io/pvc-event"
	AnnotationKeyPVCRelatedSnapshots = "storage.resource.caicloud.io/pvc-snapshots"

	// finalizers
	FinalizerNameStorageServiceController = "storage.resource.caicloud.io/storageservice-deletion"
	FinalizerNameStorageClassController   = "storage.resource.caicloud.io/storageclass-deletion"
	FinalizerNameSnapshotController       = "storage.resource.caicloud.io/snapshot-deletion"
	FinalizerNamePV                       = "storage.resource.caicloud.io/pv-deletion"
	FinalizerNameNodeLocalStorage         = "storage.resource.caicloud.io/nls-deletion"
)
