package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MLProjectPlural is the Plural for MLProject.
	MLProjectPlural = "mlprojects"

	// MLNeuronPlural is the Plural for MLNeuron.
	MLNeuronPlural = "mlneurons"

	// MLTaskPlural is the Plural for MLTask.
	MLTaskPlural = "mltasks"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MLProject is a specification for a MLProject resource.
type MLProject struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the MLProject.
	Spec MLProjectSpec `json:"spec"`

	// Most recently observed status of the MLProject.
	Status MLProjectStatus `json:"status"`
}

// MLProjectSpec is the spec for a MLProject resource.
type MLProjectSpec struct {
	// Steps represent the topology order in which MLNeuron are executed in MLProject.
	Steps []Step `json:"steps"`

	// Storage contains the storage information of the project,
	// which can be the storage class or the name of the data volume.
	// The storage of the project is used to store the temporary data of the MLNeuron runtime.
	Storage StorageConfig `json:"storage"`
}

type StorageConfig struct {
	// Size is the storage size
	Size string `json:"size"`
	// Alias is user-defined name of class name
	Alias string `json:"alias"`
	// ClassName is the storage class name.
	ClassName string `json:"className"`
	// PersistentVolumeClaimName is name of storage
	PersistentVolumeClaimName *string `json:"persistentVolumeClaimName,omitempty"`
}

// MLProjectPhase is the state of MLProject.
type MLProjectPhase string

const (
	// ProjectEmpty is the empty state of MLProject.
	// MLProject state is empty when just created it.
	ProjectEmpty MLProjectPhase = ""

	// ProjectReady is the ready state of MLProject
	// MLProject state will be transformed to "Ready" when the serial processes at all MLNeurons execute successfully
	// Pipeline can only be published after a successful execution
	ProjectReady MLProjectPhase = "Ready"

	// ProjectNotReady indicates that the current project cannot be published to the pipeline
	ProjectNotReady MLProjectPhase = "NotReady"
)

// MLProjectStatus is the status for a MLProject resource
type MLProjectStatus struct {
	// MLProject phase updates are made after project execution
	Phase MLProjectPhase `json:"phase"`
}

// Step is the order in which MLNeuron are executed in MLProject
type Step struct {
	// Name of step name as defined by user.
	Name string `json:"name"`

	// Unique ID of step.
	StepID string `json:"stepId"`

	// A list of Neurons belonging to this step.
	MLNeuronRefs []*MLNeuron `json:"mlNeuronRefs"`

	// CreationTime is a timestamp representing the server time when this object was created.
	CreationTime *metav1.Time `json:"creationTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MLProjectList is the list of Projects.
type MLProjectList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	metav1.ListMeta `json:"metadata"`

	// List of Projects.
	Items []MLProject `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MLNeuron is a specification of a MLNeuron resource.
// MLNeuron is the minimal management and execution unit in Clever,
// which defines a set resources (e.g. code, dataset, etc) around a single framework like TensorFlow or Spark.
// The meaning of MLNeuron varies according to use cases, for example, model training, feature engineering, etc.
// All use cases are grouped into 'FrameworkGroupType'.
type MLNeuron struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the MLNeuron.
	Spec MLNeuronSpec `json:"spec"`

	// Most recently observed status of the MLNeuron.
	Status MLNeuronStatus `json:"status"`
}

// VolumeType used to determine whether a data volume is from a non-versioned data volume.
type VolumeType string

const (
	// Represents volume created by clever platform
	// CleverVolume was called non-versioned-data
	CleverVolume VolumeType = "CleverVolume"
	// Represents volume not belong to clever platform
	OtherVolume VolumeType = "OtherVolume"
)

type DataVolume struct {
	Type VolumeType `json:"type"`
	// Volume represents a named volume in a pod that may be accessed by any container in the pod.
	Volume corev1.Volume `json:"volumes,omitempty"`
	// VolumeMount describes a mounting of a Volume within a container.
	VolumeMount corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// ReplicaSpec is a description of the replica
// Copy from https://github.com/kubeflow/common/blob/master/job_controller/api/v1/types.go#L65
type ReplicaSpec struct {
	// Replicas is the desired number of replicas of the given template.
	// If unspecified, defaults to 1.
	Replicas *int32 `json:"replicas,omitempty"`

	// Template is the object that describes the pod that
	// will be created for this replica. RestartPolicy in PodTemplateSpec
	// will be overide by RestartPolicy in ReplicaSpec
	Template corev1.PodTemplateSpec `json:"template,omitempty"`

	// Restart policy for all replicas within the job.
	// One of Always, OnFailure, Never and ExitCode.
	// Default to Never.
	RestartPolicy corev1.RestartPolicy `json:"restartPolicy,omitempty"`
}
type DataSourceType string

const (
	// Indicate the data come from clever-model
	DataTypeModel DataSourceType = "Model"
	// Indicate the data come from clever-dataset
	DataTypeDataset DataSourceType = "Dataset"
	// Indicate the data come from github or gitlab
	DataTypeCode DataSourceType = "Code"
)

// Data store information of many kinds of data
type Data struct {
	// Name used to specify a data
	Name string `json:"name"`
	// Enum value: code data model
	Type DataSourceType `json:"type"`
	// Local path to save the data
	LocalPath string `json:"localPath"`
	// DataSource represents the source and type of the mounted volume.
	DataSource `json:",inline"`
}

type DataSource struct {
	Model   *ModelDataSource   `json:"model,omitempty"`
	Dataset *DatasetDataSource `json:"dataset,omitempty"`
	Code    *CodeDataSource    `json:"code,omitempty"`
}

// ModelDataSource represents the data come from a clever-model
type ModelDataSource struct {
	Version       string `json:"version"`
	FrameworkType string `json:"frameworkType"`
	FormatType    string `json:"formatType"`
	Tenant        string `json:"tenant"`
	Input         string `json:"input"`
	Output        string `json:"output"`
	GPU           string `json:"gpu"`
}

// DatasetDataSource represents the data come from clever-dataset
type DatasetDataSource struct {
	Version string `json:"version"`
}

// The type of repository
type RepositoryType string

const (
	// CustomRepositoryType represents a user-defined repository.
	CustomRepositoryType RepositoryType = "Custom"

	// PresetRepositoryType represents a preset repository of the platform.
	PresetRepositoryType RepositoryType = "Preset"

	// PresetPublicRepositoryType represents a public preset repository of the platform.
	PresetPublicRepositoryType RepositoryType = "PresetPublic"
)

// CodeDataSource represents the data come from a code repository
type CodeDataSource struct {
	Type RepositoryType `json:"type"`
	// Branch like `master`
	Branch string `json:"branch"`
	// Tag like `v0.1.2`
	Tag string `json:"tag"`
	// A group name or a user name
	Owner string `json:"owner"`
	// URL for code like `https://github.com/caicloud/clientset.git`
	URL   string `json:"url"`
	Token string `json:"token"`
	// Gitlab: project ID
	ProjectID string `json:"id"`
}

type MLCommon struct {
	// List of input data.
	Inputs []Data `json:"inputs,omitempty"`
	// List of output data.
	Outputs []Data `json:"outputs,omitempty"`
	// Mountable volumes for runtime
	DataVolumes []DataVolume `json:"dataVolumes,omitempty"`

	// List of NeuronReplica
	// Replicas is used to convert to Neuron job replica, like PS, Worker in tfjob
	// Only one replica when it is stand-alone training
	// More than one replica when it is distributed training
	Replicas map[string]*ReplicaSpec `json:"replicas,omitempty"`

	// Some frame-specific fields that require manual input by the user
	// Should be store here，
	MLConf MLConfig `json:"mlConf"`
}

// MLTaskSpec is the specification of a task
type MLTaskSpec struct {
	// The MLNeuron which task creates from
	MLNeuronRef corev1.LocalObjectReference `json:"mlNeuronRef"`
	MLCommon    `json:",inline"`
}

// MLNeuronSpec is a desired state description of the Neuron.
type MLNeuronSpec struct {
	TaskTemplate MLCommon `json:"taskTemplate"`

	// Config of storageClass and storageSize
	// @TODO @tskdsb  useless or not ?
	// Storage StorageConfig `json:"storage"`

	// Just for display and path-copy in project
	MainCode string `json:"mainCode"`

	// To specify same working directory for every replica
	WorkingDir string `json:"workingDir"`
}

type MLNeuronPhase string

const (
	// MLNeuron state is empty when just created it
	MLNeuronEmpty MLNeuronPhase = ""

	// When pulling data, code, model
	MLNeuronPulling MLNeuronPhase = "Pulling"

	// When pushing data, code, model
	MLNeuronPushing MLNeuronPhase = "Pushing"

	// MLNeuron state will be transformed to "Running" by controller when
	// there has ongoing task
	MLNeuronRunning MLNeuronPhase = "Running"

	// MLNeuron state will be transformed to "Success" by controller when
	// all tasks executed successfully
	MLNeuronSucceed MLNeuronPhase = "Succeeded"

	// MLNeuron state will be transformed to "Failed" by controller when
	// one of task of tasks executed Failed
	MLNeuronFailed MLNeuronPhase = "Failed"
)

type MLNeuronStatus struct {
	// Phase is the state of the MLNeuron
	Phase MLNeuronPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MLNeuronList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	metav1.ListMeta `json:"metadata"`

	// List of MLNeuron
	Items []MLNeuron `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MLTask is the actual execution unit in a project, which stems its configuration from
// MLNeuron and Spawns MLJob (i.e. TFJob, PyTorchJob, etc).
type MLTask struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the MLNeuron.
	Spec MLTaskSpec `json:"spec"`

	// Most recently observed status of the MLNeuron.
	Status MLTaskStatus `json:"status"`
}

// MLNeuron Task Type of MLTask
type MLJobActionType string

const (
	// Pull task is used to pull input and output to MLNeuron pvc
	PullTask MLJobActionType = "Pull"
	// Submit task is used to submit MLNeuronJob
	SubmitTask MLJobActionType = "Submit"
	// Push data task is used to push output dataset to remote
	PushTask MLJobActionType = "Push"
)

// MLNeuron Task Type of MLTask
type JobType string

const (
	JobTypeJob                       JobType = "Job"
	JobTypeTFJob                             = "TFJob"
	JobTypePyTorchJob                        = "PyTorchJob"
	JobTypeMPIJob                            = "MPIJob"
	JobTypeSparkApplication                  = "SparkApplication"
	JobTypeScheduledSparkApplication         = "ScheduledSparkApplication"
	JobTypeAutoML                            = "Experiment"
)

type MLJob struct {
	// Name of job CR
	JobID string `json:"jobId"`

	// The purpose of this job
	// Enum value:
	// Pull: pull data, code, model from repository to local for runtime
	// Push: push local data, code, model to repository for save and share
	// Submit: usually a train process
	Type MLJobActionType `json:"type"`

	// Enum value: Job TFJob PyTorchJob MPIJob
	JobType JobType `json:"jobType"`

	// Pod list created by this job
	Replicas []ReplicaPod `json:"replicas"`

	// MLNeuron task's job state
	Status MLNeuronPhase `json:"status"`

	// Human readable message indicating the reason for Failure
	Message string `json:"message"`

	StartTime *metav1.Time `json:"startTime"`
	EndTime   *metav1.Time `json:"endTime"`
}

type ReplicaPod struct {
	Name  string          `json:"name"`
	Phase corev1.PodPhase `json:"phase"`
	// More information in the future
	// ...
}

type MLTaskStatus struct {
	// Jobs which MLTask created
	// Jobs run sequentially
	Jobs []MLJob `json:"jobs"`

	// Task start time
	StartTime *metav1.Time `json:"startTime"`

	// Task end time
	EndTime *metav1.Time `json:"endTime"`

	// MLNeuronStatus.Phase will update by this value
	Phase MLNeuronPhase `json:"phase"`

	// Human readable message indicating the reason for Failure
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MLTaskList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	metav1.ListMeta `json:"metadata"`

	// List of MLTask
	Items []MLTask `json:"items"`
}
