package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StorageType describes the parameters for a class of storage for
// which PersistentVolumes can be dynamically provisioned.
//
// StorageTypes are non-namespaced; the name of the storage type
// according to etcd is in ObjectMeta.Name.
type StorageType struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Provisioner indicates the type of the provisioner.
	Provisioner string `json:"provisioner" protobuf:"bytes,2,opt,name=provisioner"`

	// Parameters holds the parameters for the provisioner that should
	// create volumes of this storage type.
	// Required ones for create storage service.
	// +optional
	RequiredParameters map[string]string `json:"requiredParameters,omitempty" protobuf:"bytes,3,rep,name=requiredParameters"`

	// Parameters holds the parameters for the provisioner that should
	// create volumes of this storage type.
	// Required ones for create storage class.
	// +optional
	OptionalParameters map[string]string `json:"optionalParameters,omitempty" protobuf:"bytes,3,rep,name=classOptionalParameters"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StorageTypeList is a collection of storage types.
type StorageTypeList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of StorageClasses
	Items []StorageType `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StorageService describes the parameters for a class of storage for
// which PersistentVolumes can be dynamically provisioned.
//
// StorageServices are non-namespaced; the name of the storage service
// according to etcd is in ObjectMeta.Name.
type StorageService struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// TypeName indicates the name of the storage type that this service belongs to.
	TypeName string `json:"typeName" protobuf:"bytes,2,opt,name=typeName"`

	// Parameters holds the parameters for the provisioner that should
	// create volumes of this storage class.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty" protobuf:"bytes,3,rep,name=parameters"`

	// StorageMetaData represents the current metadata for each storage backend.
	// +optional
	StorageMetaData StorageMetaData `json:"storageMetaData,omitempty"`
}

// StorageMetaData is the data structure for each storage backend metadata.
type StorageMetaData struct {
	Ceph *CephMetaData `json:"ceph,omitempty"`
}

// CephMetaData is the data structure for Ceph metadata.
type CephMetaData struct {
	Pools []CephPool `json:"pools"`
}

// CephPool is the data structure for single Ceph storage pool.
type CephPool struct {
	Name        string `json:"name"`
	ReplicaSize int    `json:"replicaSize"`
	// total capacity of the current pool, kb
	Capacity int `json:"capacity"`
	// capacity used, kb
	Used int `json:"used"`
	// number of objects in the pool
	Objects int `json:"objects"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StorageServiceList is a collection of storage services.
type StorageServiceList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of StorageClasses
	Items []StorageService `json:"items" protobuf:"bytes,2,rep,name=items"`
}
