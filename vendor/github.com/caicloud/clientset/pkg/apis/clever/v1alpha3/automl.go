package v1alpha3

import v1 "k8s.io/api/core/v1"

// AutoMLConfig is the configuration for AutoML.
// Ref https://github.com/kubeflow/katib/blob/master/pkg/apis/controller/experiments/v1alpha3/experiment_types.go
type AutoMLConfig struct {
	// List of hyperparameter configurations.
	Parameters []ParameterSpec `json:"parameters,omitempty"`

	// Describes the objective of the experiment.
	Objective *ObjectiveSpec `json:"objective,omitempty"`

	// Describes the suggestion algorithm.
	Algorithm *AlgorithmSpec `json:"algorithm,omitempty"`

	// How many trials can be processed in parallel.
	// Defaults to 3
	ParallelTrialCount *int32 `json:"parallelTrialCount,omitempty"`

	// Max completed trials to mark experiment as succeeded
	MaxTrialCount *int32 `json:"maxTrialCount,omitempty"`

	// Max failed trials to mark experiment as failed.
	MaxFailedTrialCount *int32 `json:"maxFailedTrialCount,omitempty"`

	// For v1alpha3 we will keep the metrics collector implementation same as v1alpha1.
	MetricsCollectorSpec *MetricsCollectorSpec `json:"metricsCollectorSpec,omitempty"`
}

// ParameterSpec defines the search space and type of one parameter.
// Ref https://github.com/kubeflow/katib/blob/master/pkg/apis/controller/common/v1alpha3/common_types.go
type ParameterSpec struct {
	Name          string        `json:"name,omitempty"`
	ParameterType ParameterType `json:"parameterType,omitempty"`
	FeasibleSpace FeasibleSpace `json:"feasibleSpace,omitempty"`
}

// ParameterType defines the type of the parameter.
type ParameterType string

const (
	// ParameterTypeUnknown defines unknown type.
	ParameterTypeUnknown ParameterType = "unknown"
	// ParameterTypeDouble defines double type.
	ParameterTypeDouble ParameterType = "double"
	// ParameterTypeInt defines integer type.
	ParameterTypeInt ParameterType = "int"
	// ParameterTypeDiscrete defines discrete type, which is a discrete
	// array of numbers. For example, [1, 0.1, 0.001].
	ParameterTypeDiscrete ParameterType = "discrete"
	// ParameterTypeCategorical defines categorical type, which is a list of strings.
	// For example, ["sgd", "adam", "ftcl"].
	ParameterTypeCategorical ParameterType = "categorical"
)

// FeasibleSpace defines the search space of one parameter.
// `double` should have Max, Min, and Step. For example, Min = 1, Max = 5, Step = 0.1.
// `Int` should have Max, Min.
// `discrete` and `categorical` should have List.
type FeasibleSpace struct {
	Max  string   `json:"max,omitempty"`
	Min  string   `json:"min,omitempty"`
	List []string `json:"list,omitempty"`
	Step string   `json:"step,omitempty"`
}

// ObjectiveSpec defines the objective of the experiment.
// Ref https://github.com/kubeflow/katib/blob/master/pkg/apis/controller/common/v1alpha3/common_types.go
type ObjectiveSpec struct {
	Type                ObjectiveType `json:"type,omitempty"`
	Goal                *float64      `json:"goal,omitempty"`
	ObjectiveMetricName string        `json:"objectiveMetricName,omitempty"`
	// This can be empty if we only care about the objective metric.
	// Note: If we adopt a push instead of pull mechanism, this can be omitted completely.
	AdditionalMetricNames []string `json:"additionalMetricNames,omitempty"`
}

// ObjectiveType defines the direction of the optimization.
type ObjectiveType string

const (
	// ObjectiveTypeUnknown defines unknown direction.
	ObjectiveTypeUnknown ObjectiveType = ""
	// ObjectiveTypeMinimize implies that lower metric is better.
	ObjectiveTypeMinimize ObjectiveType = "minimize"
	// ObjectiveTypeMaximize implies that higher metric is better.
	ObjectiveTypeMaximize ObjectiveType = "maximize"
)

// AlgorithmSpec defines the optimization algorithm configuration of the experiment.
type AlgorithmSpec struct {
	AlgorithmName string `json:"algorithmName,omitempty"`
	// Key-value pairs representing settings for suggestion algorithms.
	AlgorithmSettings []AlgorithmSetting `json:"algorithmSettings"`
	EarlyStopping     *EarlyStoppingSpec `json:"earlyStopping,omitempty"`
}

// AlgorithmSetting defines the configuration of the algorithm.
type AlgorithmSetting struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// EarlyStoppingSpec defines the configuration of Early Stopping.
type EarlyStoppingSpec struct {
	// TODO
}

// MetricsCollectorSpec defines the configuration of metrics collector, which is
// a component used to collect training metrics.
// Ref https://github.com/kubeflow/katib/blob/master/pkg/apis/controller/common/v1alpha3/common_types.go
type MetricsCollectorSpec struct {
	Source    *SourceSpec    `json:"source,omitempty"`
	Collector *CollectorSpec `json:"collector,omitempty"`
}

// SourceSpec defines the source of the training metrics.
// +k8s:deepcopy-gen=true
type SourceSpec struct {
	// Model-train source code can expose metrics by http, such as HTTP endpoint in
	// prometheus metric format
	HttpGet *v1.HTTPGetAction `json:"httpGet,omitempty"`
	// During training model, metrics may be persisted into local file in source
	// code, such as tfEvent use case
	FileSystemPath *FileSystemPath `json:"fileSystemPath,omitempty"`
	// Default metric output format is {"metric": "<metric_name>",
	// "value": <int_or_float>, "epoch": <int>, "step": <int>}, but if the output doesn't
	// follow default format, please extend it here
	Filter *FilterSpec `json:"filter,omitempty"`
}

// FilterSpec defines how to filter the metrics from training logs.
// +k8s:deepcopy-gen=true
type FilterSpec struct {
	// When the metrics output follows format as this field specified, metricsCollector
	// collects it and reports to metrics server, it can be "<metric_name>: <float>" or else
	MetricsFormat []string `json:"metricsFormat,omitempty"`
}

// FileSystemKind defines how to get traning logs from the file system.
type FileSystemKind string

const (
	// DirectoryKind defines Directory kind.
	DirectoryKind FileSystemKind = "Directory"
	// FileKind defines File kind.
	FileKind FileSystemKind = "File"
	// InvalidKind is used when the kind is invalid.
	InvalidKind FileSystemKind = "Invalid"
)

// FileSystemPath is the path to the training logs in the file system.
// +k8s:deepcopy-gen=true
type FileSystemPath struct {
	Path string         `json:"path,omitempty"`
	Kind FileSystemKind `json:"kind,omitempty"`
}

// CollectorKind defines the kind of the collector.
type CollectorKind string

// CollectorSpec is the configuration of the collector.
type CollectorSpec struct {
	Kind CollectorKind `json:"kind"`
	// When kind is "customCollector", this field will be used
	CustomCollector *v1.Container `json:"customCollector,omitempty"`
}

const (
	// StdOutCollector is used to collect metrics from STDOUT.
	StdOutCollector CollectorKind = "StdOut"
	// FileCollector is used to collect metrics from File.
	FileCollector CollectorKind = "File"
)
