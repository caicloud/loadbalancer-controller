package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Snapshot is the snapshot object of the specified PVC.
type Snapshot struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotSpec   `json:"spec,omitempty"`
	Status SnapshotStatus `json:"status,omitempty"`
}

// SnapshotSpec defines the information about the PVC that needs to create a snapshot.
type SnapshotSpec struct {
	StorageSource `json:",inline"`
	PVCName       string `json:"pvcName"`
	// the unique snapshot name
	SnapshotName string `json:"snapshotName"`
}

// SnapshotStatus represents the current information about snapshot.
type SnapshotStatus struct {
	Phase   SnapshotPhase `json:"phase,omitempty"`
	Message string        `json:"message,omitempty"`
}

// StorageSource represents the source of a storage backend.
type StorageSource struct {
	RBD *RBDSource `json:"rbd,omitempty"`
}

// RBDSource represents the RBD information of this Snapshot object.
type RBDSource struct {
	CephMonitors []string `json:"monitors"`
	RadosUser    string   `json:"user"`
	// Key is the authentication secret for RBDUser, we don't use the keyring file path
	Key   string `json:"Key"`
	Pool  string `json:"pool"`
	Image string `json:"image"`
}

// SnapshotPhase defines the type of snapshot phase.
type SnapshotPhase string

const (
	// SnapshotPending represents the snapshot is creating.
	SnapshotPending SnapshotPhase = "Pending"
	// SnapshotSuccess represents the snapshot is normal.
	SnapshotSuccess SnapshotPhase = "Success"
	// SnapshotFailed represents the snapshot creation failed.
	SnapshotFailed SnapshotPhase = "Failed"
	// SnapshotInRollback represents the snapshot is in rollback.
	SnapshotInRollback SnapshotPhase = "Rollback"
	// SnapshotRollbackFailed represents the snapshot rollback failed.
	SnapshotRollbackFailed SnapshotPhase = "RollbackFailed"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnapshotList is a collection of snapshots.
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Snapshots
	Items []Snapshot `json:"items"`
}
