package v1alpha2

// Framework type of application
type FrameworkType string

// Framework defines include machine learning framework and programming language
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

	// Programming language
	JavaScript FrameworkType = "javascript"
	Python     FrameworkType = "python"
	Golang     FrameworkType = "golang"
	Scala      FrameworkType = "scala"
	Java       FrameworkType = "java"
	Cpp        FrameworkType = "cpp"
	C          FrameworkType = "c"
	R          FrameworkType = "r"

	// Programming development tools
	Jupyter            FrameworkType = "jupyter"
	Zeppelin           FrameworkType = "zeppelin"
	JupyterLab         FrameworkType = "jupyterlab"
	TensorBoard        FrameworkType = "tensorboard"
	SparkHistoryServer FrameworkType = "sparkhistoryserver"

	// ModelEvaluation
	Evaluation FrameworkType = "evaluation"
)

// Framework group type of framework
type FrameworkGroupType string

// Framework group defines
const (
	FeatureEnginnering = "FeatureEngineering"
	DevelopmentTools   = "DevelopmentTools"
	ParameterTuning    = "ParameterTuning"
	ModelEvaluation    = "ModelEvaluation"
	ModelTraining      = "ModelTraining"
	ModelServing       = "ModelServing"
	DataCleaning       = "DataCleaning"
)

type DataType string

const (
	// Model set type
	Model            DataType = "Model"
	Code             DataType = "Code"
	VersionedData    DataType = "VersionedData"
	NonVersionedData DataType = "NonVersionedData"
)
