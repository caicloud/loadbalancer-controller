package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Framework group type of framework
type FrameworkGroupType string

// Framework group defines
const (
	FeatureEnginnering FrameworkGroupType = "FeatureEngineering"
	DevelopmentTools   FrameworkGroupType = "DevelopmentTools"
	ParameterTuning    FrameworkGroupType = "ParameterTuning"
	ModelEvaluation    FrameworkGroupType = "ModelEvaluation"
	ModelTraining      FrameworkGroupType = "ModelTraining"
	ModelServing       FrameworkGroupType = "ModelServing"
	DataCleaning       FrameworkGroupType = "DataCleaning"
	PresetAlgorithm    FrameworkGroupType = "PresetAlgorithm"
	AutoML             FrameworkGroupType = "AutoML"
	DataDevelopment    FrameworkGroupType = "DataDevelopment"
	Custom             FrameworkGroupType = "Custom"
	DataSync           FrameworkGroupType = "DataSync"
	ModelCompressing   FrameworkGroupType = "ModelCompressing"
)

// Framework type of application
type FrameworkType string

const (
	// Algorithm framework
	Tensorflow FrameworkType = "tensorflow"
	TFserving  FrameworkType = "tfserving"
	Pytorch    FrameworkType = "pytorch"
	SKLearn    FrameworkType = "sklearn"
	Chainer    FrameworkType = "chainer"
	Caffe2     FrameworkType = "caffe2"
	Caffe      FrameworkType = "caffe"
	MXNet      FrameworkType = "mxnet"
	Spark      FrameworkType = "spark"
	Keras      FrameworkType = "keras"
	Onnx       FrameworkType = "onnx"
	Horovod    FrameworkType = "horovod"
	TensorRT   FrameworkType = "tensorrt"

	// HyperparameterTuning is only used for template in clever admin.
	HyperparameterTuning FrameworkType = "hyperparameter tuning"

	// Programming language
	JavaScript FrameworkType = "javascript"
	Python     FrameworkType = "python"
	Golang     FrameworkType = "golang"
	Scala      FrameworkType = "scala"
	Java       FrameworkType = "java"
	Cpp        FrameworkType = "cpp"
	C          FrameworkType = "c"
	R          FrameworkType = "r"

	// DataSync
	DataX FrameworkType = "datax"

	// ModelEvaluation
	Evaluation FrameworkType = "evaluation"
)

// Training framework like: horovod, stand-alone, PS, multi-worker.
type TrainingType string

const (
	StandAloneTraining   TrainingType = "StandAlone"
	MultiWorkerTraining  TrainingType = "MultiWorker"
	PSWorkerTraining     TrainingType = "PSWorker"
	MasterWorkerTraining TrainingType = "MasterWorker"
	HorovodTraining      TrainingType = "Horovod"
	Distributed          TrainingType = "Distributed"
)

type MLConfig struct {
	// Type defines the framework group.
	Type FrameworkGroupType `json:"type"`

	// Framework defines the type of the machine learning framework or programming language.
	Framework FrameworkType `json:"framework"`

	// xxxConf is the special fields required when converting to xxxJob struct
	TensorFlowConf *TensorFlowConfig `json:"tensorFlowConf,omitempty"`
	PyTorchConf    *PyTorchConfig    `json:"pyTorchConf,omitempty"`
	SparkConf      *SparkConfig      `json:"sparkConf,omitempty"`
	JupyterConf    *JupyterConfig    `json:"jupyterConf,omitempty"`
	AutoMLConf     *AutoMLConfig     `json:"automlConf,omitempty"`
	Horovod        *HorovodConfig    `json:"horovodConf,omitempty"`
	DataXConf      *DataXConfig      `json:"dataxConf,omitempty"`
}

type DataXConfig struct {
	// "From" indicates the data source for data synchronization
	Reader Reader `json:"reader,omitempty"`
	// "To" indicates the destination of data synchronization
	Writer Writer `json:"writer,omitempty"`
}

type ReaderType string

const (
	FTPReader   ReaderType = "ftpreader"
	HDFSReader  ReaderType = "hdfsreader"
	HiveReader  ReaderType = "hivereader"
	MysqlReader ReaderType = "mysqlreader"
)

type Reader struct {
	Type ReaderType `json:"type"`

	// DataSourceSecret carries the name of the Secret containing DataSource configuration files.
	DataSourceSecret string `json:"dataSourceSecret,omitempty"`
	// xxxReaderPlugin is used to read data from the source data source
	HDFSReaderPlugin  *HDFSReaderPlugin  `json:"hdfsReaderPlugin,omitempty"`
	HiveReaderPlugin  *HiveReaderPlugin  `json:"hiveReaderPlugin,omitempty"`
	FTPReaderPlugin   *FTPReaderPlugin   `json:"ftpReaderPlugin,omitempty"`
	MysqlReaderPlugin *MysqlReaderPlugin `json:"mysqlReaderPlugin,omitempty"`
}

type HDFSReaderPlugin struct {
	// FieldDelimiter is the field delimiter written, by default ","
	FieldDelimiter string `json:"fieldDelimiter,omitempty"`
	// Path represent the absolute path to a folder or file
	Path string `json:"path,omitempty"`
	// FileType represents the file type, such as csv, text, etc.
	FileType string `json:"fileType,omitempty"`
	// IsFile indicates whether the data synchronization task is to synchronize a single file
	IsFile bool `json:"isFile,omitempty"`
}

type FTPReaderPlugin struct {
	// FieldDelimiter is the field delimiter written, by default ","
	FieldDelimiter string `json:"fieldDelimiter,omitempty"`
	// Path represent the absolute path to a folder or file
	Path string `json:"path,omitempty"`
	// FileType represents the file type, such as csv, text, etc.
	FileType string `json:"fileType,omitempty"`
	// IsFile indicates whether the data synchronization task is to synchronize a single file
	IsFile bool `json:"isFile,omitempty"`
}

type HiveReaderPlugin struct {
	// The table name of the data table
	Table string `json:"table,omitempty"`
}

type MysqlReaderPlugin struct {
	// The table name of the data table
	Table string `json:"table,omitempty"`
}

type WriteMode string

const (
	// truncate, clean up all files with the fileName prefix in the directory before writing.
	TruncateWriteMode WriteMode = "truncate"
	// append, do not do any processing before writing, directly uses filename to write, and ensure that the file name does not conflict.
	AppendWriteMode WriteMode = "append"
)

type WriterType string

const (
	TxtFileWriter WriterType = "txtfilewriter"
)

type Writer struct {
	Type WriterType `json:"type"`

	// xxxWriterPlugin is used to write data to the target data source
	TxtFileWriterPlugin *TxtFileWriterPlugin `json:"txtFileWriterPlugin,omitempty"`
}

type TxtFileWriterPlugin struct {
	// Path indicates the destination of data synchronization
	Path string `json:"path"`
	//FileName is the file name stored in data synchronization
	FileName string `json:"fileName"`
	// WriteMode represents the data cleaning processing mode before writing
	WriteMode WriteMode `json:"writeMode"`
	// FieldDelimiter is the field delimiter written, by default ","
	FieldDelimiter string `json:"fieldDelimiter,omitempty"`
}

type JupyterConfig struct {
	// The last time user interact with Jupyter
	// Including running some task with out watching
	LastActiveTime *metav1.Time `json:"lastActiveTime"`
	// Time in seconds to kill Jupyter since LastActiveTime
	IdleTimeout *int32 `json:"idleTimeout"`
}

// Crontab config of spark
type SparkSchedule struct {
	// Crontab string like: "0 0 9 * * *"
	Crontab string `json:"crontab"`
	// Crontab works only when Enable is true
	Enable bool `json:"enable"`
}

// SparkApplicationType describes the type of a Spark application.
type SparkApplicationType string

// Different types of Spark applications.
const (
	JavaApplicationType   SparkApplicationType = "Java"
	ScalaApplicationType  SparkApplicationType = "Scala"
	PythonApplicationType SparkApplicationType = "Python"
	RApplicationType      SparkApplicationType = "R"
)

// DeployMode describes the type of deployment of a Spark application.
type DeployMode string

// Different types of deployments.
const (
	ClusterMode         DeployMode = "cluster"
	ClientMode          DeployMode = "client"
	InClusterClientMode DeployMode = "in-cluster-client"
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

type HorovodConfig struct {
	// Specifies the number of slots per worker used in hostfile.
	// Defaults to 1.
	// +optional
	SlotsPerWorker *int32 `json:"slotsPerWorker,omitempty"`

	// TODO: Move this to `RunPolicy` in common operator, see discussion in https://github.com/kubeflow/tf-operator/issues/960
	// Specifies the number of retries before marking this job failed.
	// Defaults to 6.
	// +optional
	BackoffLimit *int32 `json:"backoffLimit,omitempty"`

	// TODO: Move this to `RunPolicy` in common operator, see discussion in https://github.com/kubeflow/tf-operator/issues/960
	// Specifies the duration in seconds relative to the start time that
	// the job may be active before the system tries to terminate it.
	// Note that this takes precedence over `BackoffLimit` field.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// CleanPodPolicy defines the policy that whether to kill pods after the job completes.
	// Defaults to Running.
	CleanPodPolicy string `json:"cleanPodPolicy,omitempty"`
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
