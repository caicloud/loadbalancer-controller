package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	FlavorPlural   = "flavors"
	ProjectPlural  = "projects"
	MLNeuronPlural = "mlneurons"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Project is a specification for a Project resource.
type Project struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Project.
	Spec ProjectSpec `json:"spec"`

	// Most recently observed status of the Project.
	Status ProjectStatus `json:"status"`
}

// ProjectSpec is the spec for a Project resource.
type ProjectSpec struct {
	// Steps represent the topology order in which MLNeuron are executed in Project.
	Steps []Step `json:"steps"`

	Storage StorageConfig `json:"storage"`

	// Resources represent the resource quota of Project
	Resources ResourceConfig `json:"resources"`

	// Preset repository of the platform.
	Repository *RepositoryConfig `json:"repository,omitempty"`
}

// The type of repository
type RepositoryType string

const (
	// CustomRepositoryType represents a user-defined repository.
	CustomRepositoryType RepositoryType = "custom"

	// PresetRepositoryType represents a preset repository of the platform.
	PresetRepositoryType RepositoryType = "preset"

	// PresetPublicRepositoryType represents a public preset repository of the platform.
	PresetPublicRepositoryType RepositoryType = "preset-public"
)

type RepositoryConfig struct {
	ID       string         `json:"id"`
	URL      string         `json:"url"`
	Name     string         `json:"name"`
	Username string         `json:"username"`
	Password string         `json:"password"`
	Tag      string         `json:"tag"`
	Type     RepositoryType `json:"type"`
	Alias    string         `json:"alias"`
}

type StorageConfig struct {
	//Size is the storage size
	Size string `json:"size"`
	// Alias is user-defined name of class name
	Alias string `json:"alias"`
	// ClassName is the storage class name.
	ClassName string `json:"className"`
	// PersistentVolumeClaimName is name of storage
	PersistentVolumeClaimName *string `json:"persistentVolumeClaimName,omitempty"`
}

// ProjectPhase is the state of Project.
type ProjectPhase string

const (
	// ProjectEmpty is the empty state of Project.
	// Project state is empty when just created it.
	ProjectEmpty ProjectPhase = ""

	// ProjectReady is the ready state of Project
	// Project state will be transform to "Ready" when the serial processes at all MLNeurons execute successfully
	// Workflow can only be published after a successful execution
	ProjectReady ProjectPhase = "Ready"

	// ProjectNotReady indicates that the current project cannot be published to the pipeline
	ProjectNotReady ProjectPhase = "NotReady"
)

// ProjectStatus is the status for a Project resource
type ProjectStatus struct {
	// Project status updates are made after project execution
	Phase ProjectPhase `json:"phase"`
}

// Step is the order in which MLNeuron are executed in Project
type Step struct {
	// Name of step name define by user.
	Name string `json:"name"`

	// Unique ID of step.
	StepID string `json:"stepId"`

	// A list of Neurons belong to this step.
	MLNeuronRefs []MLNeuronRef `json:"NeuronRefs"`

	// CreationTime is a timestamp representing the server time when this object was created.
	CreationTime *metav1.Time `json:"creationTime"`
}

// MLNeuronRef is the reference of MLNeuron crd
type MLNeuronRef struct {
	// Name of reference MLNeuron crd name.
	Name string `json:"name"`

	// The original name the current neuron come from
	Origin string `json:"origin"`

	// Time when the MLNeuron is added to this Step
	AdditionTime *metav1.Time `json:"additionTime"`

	// Used for show simple status in project-step page with request every MLNeuron
	Phase MLNeuronPhase `json:"phase"`

	// TODO @codeflitting Add more fields according to product display style
	// Can only add fields that will not be changed after creation
	// Should add MLNeuron info when created from Project
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProjectList is the list of Projects.
type ProjectList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	metav1.ListMeta `json:"metadata"`

	// List of Projects.
	Items []Project `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MLNeuron is a specification of a MLNeuron resource.
// MLNeuron is the minimal management and execution unit in Clever,
// which defines a set resources (e.g. code, dataset, etc) around a single framework like TensorFlow or Spark.
// The meaning of MLNeuron varies according to use cases, for example, model training, feature engineering, etc.
// All use cases as grouped into 'FrameworkGroupType'.
type MLNeuron struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the MLNeuron.
	Spec MLNeuronSpec `json:"spec"`

	// Most recently observed status of the MLNeuron.
	Status MLNeuronStatus `json:"status"`
}

// MLNeuronSpec is a desired state description of the Neuron.
type MLNeuronSpec struct {
	// Creator who create this MLNeuron
	Creator string `json:"creator"`

	// Config of storageClass and storageSize
	Storage StorageConfig `json:"storage"`

	// Volumes is the list of Kubernetes volumes that can be mounted by the Neuron.
	// Optional.
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// VolumeMounts specifies the volumes listed in ".spec.volumes" to mount into NeuronJob.
	// Optional.
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`

	// List of NeuronReplica
	// Replicas is used to convert to Neuron job replica, like PS, Worker in tfjob
	// Only one replica when it is stand-alone training
	// More than one replica when it is distributed training
	Replicas []NeuronReplica `json:"replicas,omitempty"`

	// Repository information
	Repository *RepositoryConfig `json:"repository,omitempty"`

	// The default directory when user start work
	WorkDir string `json:"workDir"`

	// The main code file will run first
	MainCode string `json:"mainCode"`

	// List of input dataSource.
	Inputs []DataSource `json:"inputs,omitempty"`

	// List of output dataSource.
	Outputs []DataSource `json:"outputs,omitempty"`

	// Image is the image config.
	Image ImageFlavor `json:"image"`

	// Env is the MLNeuron env
	Env []corev1.EnvVar `json:"env,omitempty"`

	// May used when this MLNeuron runs
	Command string `json:"command"`

	// Tasks is the job task created by controller
	Task *MLNeuronTask `json:"task,omitempty"`

	// Some frame-specific fields that require manual input by the user
	// Should be store here，
	NeuronConf MLNeuronConfig `json:"neuronConf"`
}

// Replication when Neuron runs
type NeuronReplica struct {
	// Used to specify type among MLNeuronConfig
	Type string `json:"type"`
	// Quantity of replication
	Count        int32          `json:"count"`
	ResourceConf ResourceConfig `json:"resourceConf"`
}

// Training framework like: horovod, stand-alone, PS, multi-worker.
type TrainingType string

const (
	StandAloneTraining   TrainingType = "Stand-Alone"
	MultiWorkerTraining  TrainingType = "Multi-Worker"
	PSWorkerTraining     TrainingType = "PS-Worker"
	MasterWorkerTraining TrainingType = "Master-Worker"
	HorovodTraining      TrainingType = "Horovod"
)

type MLNeuronConfig struct {
	// Type defines the framework group.
	Type FrameworkGroupType `json:"type"`

	// Framework defines the type of the machine learning framework or programming language.
	Framework FrameworkType `json:"framework"`

	// xxxConf is the special fields required when converting to xxxJob struct
	TensorFlowConf  *TensorFlowConfig  `json:"tensorFlowConf,omitempty"`
	PyTorchConf     *PyTorchConfig     `json:"pyTorchConf,omitempty"`
	SparkConf       *SparkConfig       `json:"sparkConf,omitempty"`
	JupyterConf     *JupyterConfig     `json:"jupyterConf,omitempty"`
	TensorBoardConf *TensorBoardConfig `json:"tensorBoardConf,omitempty"`
}

type TensorBoardConfig struct {
	// ResourceConf represent the resource quota of TensorBoard
	ResourceConf ResourceConfig `json:"resourceConf"`

	// Image is the image config.
	Image ImageFlavor `json:"image"`
}

type JupyterConfig struct {
	// The last time user interact with Jupyter
	// Including running some task with out watching
	LastActiveTime *metav1.Time `json:"lastActiveTime"`
	// Time in seconds to kill Jupyter since LastActiveTime
	IdleTimeout *int32 `json:"idleTimeout"`
}

type SparkSchedule struct {
	// Crontab string like: "0 0 9 * * *"
	Crontab string `json:"crontab"`
	// Crontab works only when Enable is true
	Enable bool `json:"enable"`
}

type SparkApplicationType string

// Different types of Spark applications.
const (
	JavaApplication  SparkApplicationType = "Java"
	ScalaApplication SparkApplicationType = "Scala"
)

// DeployMode describes the type of deployment of a Spark application.
type DeployMode string

// type of deployment of a Spark application.
const (
	ClusterMode DeployMode = "cluster"
)

type SparkConfig struct {
	// Type tells the type of the Spark application.
	Type SparkApplicationType `json:"type"`
	// Mode is the deployment mode of the Spark application.
	Mode DeployMode `json:"mode,omitempty"`
	// MainClass is the fully-qualified main class of the Spark application.
	// This only applies to Java/Scala Spark applications.
	// Optional.
	MainClass string `json:"mainClass,omitempty"`
	// MainFile is the path to a bundled JAR, Python, or R file of the application.
	// Optional.
	MainApplicationFile string `json:"mainApplicationFile"`
	// Arguments is a list of arguments to be passed to the application.
	// Optional.
	Arguments []string `json:"arguments,omitempty"`
	// SparkConf carries user-specified Spark configuration properties as they would use the  "--conf" option in
	// spark-submit.
	// Optional.
	SparkConf map[string]string `json:"sparkConf,omitempty"`
	// HadoopConf carries user-specified Hadoop configuration properties as they would use the  the "--conf" option
	// in spark-submit.  The SparkApplication controller automatically adds prefix "spark.hadoop." to Hadoop
	// configuration properties.
	// Optional.
	HadoopConf map[string]string `json:"hadoopConf,omitempty"`
	// SparkConfigMap carries the name of the ConfigMap containing Spark configuration files such as log4j.properties.
	// The controller will add environment variable SPARK_CONF_DIR to the path where the ConfigMap is mounted to.
	// Optional.
	SparkConfigMap string `json:"sparkConfigMap,omitempty"`
	// HadoopConfigMap carries the name of the ConfigMap containing Hadoop configuration files such as core-site.xml.
	// The controller will add environment variable HADOOP_CONF_DIR to the path where the ConfigMap is mounted to.
	// Optional.
	HadoopConfigMap string `json:"hadoopConfigMap,omitempty"`
	// Deps captures all possible types of dependencies of a Spark application.
	Deps Dependencies `json:"deps"`
	// Schedule determines how the scheduled task is run, whether it is using scheduleSparkApplication or cyclone
	Schedule SparkSchedule `json:"schedule"`
}

// Dependencies specifies all possible types of dependencies of a Spark application.
type Dependencies struct {
	// Jars is a list of JAR files the Spark application depends on.
	// Optional.
	Jars []string `json:"jars,omitempty"`
	// Files is a list of files the Spark application depends on.
	// Optional.
	Files []string `json:"files,omitempty"`
	// PyFiles is a list of Python files the Spark application depends on.
	// Optional.
	PyFiles []string `json:"pyFiles,omitempty"`
}

type TensorFlowConfig struct {
	// CleanPodPolicy defines the policy to kill pods after TFJob is
	// succeeded.
	// Default to Running.
	CleanPodPolicy string `json:"cleanPodPolicy,omitempty"`

	// TTLSecondsAfterFinished is the TTL to clean up tf-jobs (temporary
	// before kubernetes adds the cleanup controller).
	// It may take extra ReconcilePeriod seconds for the cleanup, since
	// reconcile gets called periodically.
	// Default to infinite.
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`

	TrainingType TrainingType `json:"trainingType"`
}

type PyTorchConfig struct {
	// CleanPodPolicy defines the policy to kill pods after PyTorchJob is
	// succeeded.
	// Default to Running.
	CleanPodPolicy string `json:"cleanPodPolicy,omitempty"`

	// TTLSecondsAfterFinished is the TTL to clean up pytorch-jobs (temporary
	// before kubernetes adds the cleanup controller).
	// It may take extra ReconcilePeriod seconds for the cleanup, since
	// reconcile gets called periodically.
	// Default to infinite.
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`

	TrainingType TrainingType `json:"trainingType"`
}

// DataSource store information of data set
type DataSource struct {
	// Name used to specify a data
	Name string `json:"name"`
	// Type can be `model`, `data`, ...
	Type DataType `json:"type"`
	// A data may have many version
	Version string `json:"version"`
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

	// MLNeuron state will be transform to "Success" by controller when
	// all tasks executed successfully
	MLNeuronSucceed MLNeuronPhase = "Succeeded"

	// MLNeuron state will be transform to "Failed" by controller when
	// one of task of tasks executed Failed
	MLNeuronFailed MLNeuronPhase = "Failed"
)

type MLNeuronStatus struct {
	// Save status for MLNeuron when it runs every time
	Histories []MLNeuronTask `json:"histories"`
	// Phase is the state of the MLNeuron
	Phase MLNeuronPhase `json:"phase"`
}

// MLNeuron Task Type of MLNeuronTask
type MLNeuronTaskJobType string

const (
	// Pull  task is used to pull input and output to MLNeuron pvc
	PullTask MLNeuronTaskJobType = "Pull"

	// Submit  task is used to submit MLNeuronJob
	SubmitTask MLNeuronTaskJobType = "Submit"

	// Push data task is used to push output dataset to remote
	PushTask MLNeuronTaskJobType = "Push"
)

// Data sets
type MLNeuronTaskJobData struct {
	Inputs     []DataSource      `json:"inputs"`
	Outputs    []DataSource      `json:"outputs"`
	Repository *RepositoryConfig `json:"repository,omitempty"`
}

type MLNeuronTaskJob struct {
	// MLNeuron task's job name
	JobID string `json:"jobId"`

	// MLNeuron task type
	Type MLNeuronTaskJobType `json:"type"`

	// Data sets and model sets and code repository that need to be pulled or pushed
	Data *MLNeuronTaskJobData `json:"data,omitempty"`

	// MLNeuron task's job state
	Status MLNeuronPhase `json:"status"`

	StartTime *metav1.Time `json:"startTime"`
	EndTime   *metav1.Time `json:"endTime"`
}

type MLNeuronTask struct {
	// Task id is unique in mlneuron
	TaskID string `json:"taskId"`

	// Submitter of this MLNeuron task
	Submitter string `json:"submitter"`

	// Task start time
	StartTime *metav1.Time `json:"startTime"`

	// Task end time
	EndTime *metav1.Time `json:"endTime"`

	// Job list of MLNeuron
	Jobs []MLNeuronTaskJob `json:"jobs"`

	// MLNeuronStatus.Phase will update by this value
	Phase MLNeuronPhase `json:"phase"`

	// Human readable message indicating the reason for Failure
	Message string `json:"message"`

	// remove input,output when task succeed or failed
	CleanAfterFinish bool `json:"cleanAfterFinish"`
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
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Flavor struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the Flavor.
	Spec FlavorSpec `json:"spec"`
}

// FlavorSpec is a desired state description of the Flavor.
type FlavorSpec struct {
	// A list of images recommended to the user for selection.
	Images []ImageFlavor `json:"images"`

	// A list of resources recommended to the user for selection.
	ResourceList []ResourceConfig `json:"resourceList"`
}

// The type of image
type ImageType string

const (
	// CustomImageType represents a user-defined image.
	CustomImageType ImageType = "custom"

	// PresetImageType represents a preset image of the platform.
	PresetImageType ImageType = "preset"

	// PresetImageType represents a preset pure image of the platform.
	PresetPureImageType ImageType = "preset-pure"
)

// ImageFlavor is the image configuration.
type ImageFlavor struct {
	// The name of the image shown to user.
	Name string `json:"name"`

	// Docker registry of the image mirror.
	Image string `json:"image"`

	// Type of image
	Type ImageType `json:"type"`
}

// Recommended resource configuration.
type ResourceFlavor struct {
	// CPU, in cores. (500m = .5 cores).
	CPU string `json:"cpu"`

	// NVIDIA GPU, in devices.
	GPU string `json:"gpu"`

	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024).
	Memory string `json:"memory"`
}

// Recommended resource configuration with limit
type ResourceConfig struct {
	// Name of recommend resource configuration.
	Name string `json:"name"`

	// Requests of resource quota
	Requests ResourceFlavor `json:"requests"`

	// Limits of resource quota
	Limits ResourceFlavor `json:"limits"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FlavorList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	metav1.ListMeta `json:"metadata"`

	// List of Flavor
	Items []Flavor `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// User for cascading delete
type MLNeuronTaskOwner struct {
	metav1.TypeMeta `json:",inline"`

	// Standard object's metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type MLNeuronTaskOwnerList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	metav1.ListMeta `json:"metadata"`

	Items []MLNeuronTaskOwner `json:"items"`
}
