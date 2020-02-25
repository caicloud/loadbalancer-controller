/*
Copyright 2019 CaiCloud.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EvalJobSpec defines the desired state of EvalJob
type EvalJobSpec struct {
	// CleanPodPolicy defines the policy to kill pods after EvalJob is
	// succeeded.
	// Default to None.
	CleanPodPolicy *CleanPodPolicy `json:"cleanPodPolicy,omitempty"`

	// Resource requirements for evaluation instance.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Creator who create the evalJob
	Creator string `json:"creator"`

	// StorageConfig is the configuration about the storage.
	Storage *StorageConfig `json:"storageConfig,omitempty"`

	// Models contains all models that will be evaluated.
	Models []EvalModel `json:"models,omitempty"`

	// Functions contains all evaluation functions.
	Functions []EvalFunc `json:"functions,omitempty"`

	// Parallelism
	Parallelism *int `json:"parallelism,omitempty"`

	// Template specified the pod template.
	// This v1alpha1 version, we only care about 'labels','annotations','env' and 'resources'.
	// We will overwrite the image, command, args when really creating pod.
	Template corev1.PodTemplateSpec `json:"template,omitempty"`
}

// +k8s:deepcopy-gen=true

// EvalJobStatus represents the current observed state of the EvalJob.
type EvalJobStatus struct {
	// The phase of EvalJob.
	Phase JobPhaseType `json:"phase,omitempty"`
	// VolumeName is the name of the volume.
	VolumeName string `json:"volumeName,omitempty"`
	// The status of each worker pod.
	WorkerStatuses []EvalWorkerStatus `json:"workerStatuses,omitempty"`
	// Represents the lastest available observations of a EvalJob's current state.
	Conditions []EvalJobCondition `json:"conditions,omitempty"`
	// The time that worker pods are created.
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// The end time of this evaljob finished.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

type EvalJobConditionType string

// These are valid conditions of a EvalJob
const (
	// Processing means the evaljob has been created correctly
	// and it is processing tasks in the right way.
	JobProcessing EvalJobConditionType = "Processing"
)

// +k8s:deepcopy-gen=true
// EvalJobCondition describe the state of a evaljob at a certain point.
type EvalJobCondition struct {
	// Type of job condition.
	Type EvalJobConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EvalJob is the Schema for the evaljobs API
// +k8s:openapi-gen=true
type EvalJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EvalJobSpec   `json:"spec,omitempty"`
	Status EvalJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EvalJobList contains a list of EvalJob
type EvalJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EvalJob `json:"items"`
}

// EvalModel is the descriptor of a model
type EvalModel struct {
	// Name of model.
	Name string `json:"name"`
	// Version of model.
	Version string `json:"version"`
	// Format of model
	Format string `json:"format"`
	// Framework of model.
	Framework string `json:"framework"`
}

// EvalFunc is the descriptor of an evaluation function
type EvalFunc struct {
	// Name of evaluation function.
	Name string `json:"name"`
	// URL of function registry
	URL string `json:"url"`
}

// StorageConfig is the type for the configuration about storage.
type StorageConfig struct {
	// PersistentVolumeClaim is shared by all Pods in the same Evaluation Job.
	// The PVC must be ReadWriteMany in order to be used for multiple evaluation instances.
	// Size is the size for the storage.
	Size string `json:"size,omitempty"`
	// ClassName is the storageclass name.
	ClassName string `json:"className,omitempty"`
	// PersistentVolumeClaimName is name of storage.
	// When the field is empty, evaluation controller will create one; otherwise, given storage
	// will be used, and all other fields are ignored.
	PersistentVolumeClaimName *string `json:"persistentVolumeClaimName,omitempty"`
}

// EvalWorkerStatus represents the current observed state of the specified worker pod.
type EvalWorkerStatus struct {
	// model information
	Model EvalModel `json:"model"`
	// function information
	Function EvalFunc `json:"function"`
	// phase of the worker pod
	Phase WorkerPhaseType `json:"phase"`
	// name of the worker pod
	PodName string `json:"podName"`
	// scheduled represents if the pod had been scheduled.
	Scheduled bool `json:"scheduled"`
	// message for phase
	Message string `json:"message"`
}

type WorkerPhaseType string

// These are valid phases of an Worker pod.
const (
	// WorkerPending means the worker is waiting to be created.
	WorkerPending WorkerPhaseType = "Pending"

	// WorkerRunning means the worker is prepared to be created.
	WorkerRunning WorkerPhaseType = "Running"

	// WorkerSucceeded means the pod of this worker
	// had been completed successfully.
	WorkerSucceeded WorkerPhaseType = "Succeeded"

	// WorkerFailed means the worker has failed.
	WorkerFailed WorkerPhaseType = "Failed"
)

type JobPhaseType string

// These are valid phases of an EvalJob.
const (
	// JobPending means the job is creating pvc.
	JobPending JobPhaseType = "Pending"

	// JobLauncing means launcher pod has been created.
	JobLaunching JobPhaseType = "Launching"

	// JobLaunced means launcher pod has been completed.
	JobLaunched JobPhaseType = "Launched"

	// JobRunning means one or more worker pods had been running.
	JobRunning JobPhaseType = "Running"

	// JobCompleted means all pods of this job
	// had been completed.
	JobCompleted JobPhaseType = "Completed"

	// JobFailed means launcher pod has been failed.
	JobFailed JobPhaseType = "Failed"

	// JobAborted means the job has been aborted.
	JobAborted JobPhaseType = "Aborted"
)

// CleanPodPolicy describes how to deal with pods when the EvalJob is finished.
type CleanPodPolicy string

const (
	// User hasn't defined the cleanPodPolicy, do nothing.
	CleanPodPolicyUndefined CleanPodPolicy = ""
	// CleanPodPolcicyAll means the controller will delete all pods after it finished.
	CleanPodPolicyAll CleanPodPolicy = "All"
	// CleanPodPolicyRunning means the controller only delete pods that are still running after it finished.
	CleanPodPolicyRunning CleanPodPolicy = "Running"
)
