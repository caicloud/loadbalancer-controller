package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ServingPlural = "servings"
)

// Serving defines a serving deployment.
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=serving
// +kubebuilder:subresource:status
type Serving struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServingSpec   `json:"spec,omitempty"`
	Status            ServingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=servings

// ServingList describes an array of Serving instances
type ServingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of Servings
	Items []Serving `json:"items"`
}

// ServingType is the type of serving jobs.
type ServingType string

const (
	// GraphPipe is the type to serve with GraphPipe
	// This is the serving type in scene in 1.4.0 and 1.4.1.
	GraphPipe ServingType = "GraphPipe"
	// TensorRT is the type to serve with TensorRT.
	// This is the serving type in scene in 1.4.2 and decrypted after 1.5.0.
	TensorRT ServingType = "TensorRT"
	// TRTIS is the type to serve with TensorRT Server Inference.
	// This is the serving type in 1.5.0.
	TRTIS ServingType = "TRTIS"
	// GPUSharing is custom serving implemented with TRTIS + Owl.
	// By default, GPUSharing is turned off
	GPUSharing ServingType = "GPUSharing"
	// Custom is the type to serve with Custom Images.
	// To mount multiple models on the same GPU, use TRTIS + Owl image with Custom type
	Custom ServingType = "Custom"
)

// ServingSpec defines the specification of serving deployment.
type ServingSpec struct {
	// Resource requirements for serving instance.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Number of replicas for a serving instance. This is fixed to 1 for GPUSharing serving type.
	Replicas int32 `json:"replicas"`
	// Scaling is the configuration about how to scale the serving service.
	// If the scalingConfig is defined, replicas will not work.
	Scaling *ScalingConfig `json:"scalingConfig,omitempty"`
	// UserInfo is the information about the user.
	UserInfo *UserInformation `json:"userInfo,omitempty"`
	// StorageConfig is the configuration about the storage.
	Storage *StorageConfig `json:"storageConfig,omitempty"`
	// Models is the list of models to be served via the serving deployment.
	// In 1.5.0, the size of the slice will be 1.
	Models []ServingModel `json:"models,omitempty"`
	// Type of the Serving
	Type ServingType `json:"type,omitempty"`
	// Image used by the Serving
	Image string `json:"image,omitempty"`
	// Command used by custom serving
	Command []string `json:"command,omitempty"`
	// Args used by custom serving
	Args []string `json:"args,omitempty"`
	// PostStart script used by the Serving
	PostStart []string `json:"poststart,omitempty"`
	// Ports that are used for Custom serving
	Ports []int `json:"ports,omitempty"`
}

// StorageConfig is the type for the configuration about storage.
type StorageConfig struct {
	// PersistentVolumeClaim is shared by all Pods in the same Serving.
	// The PVC must be ReadWriteMany in order to be used for multiple serving instances.
	// Size is the size for the storage.
	Size string `json:"size,omitempty"`
	// ClassName is the storageclass name.
	ClassName string `json:"className,omitempty"`
	// PersistentVolumeClaimName is name of storage.
	// When the field is empty, evaluation controller will create one; otherwise, given storage
	// will be used, and all other fields are ignored.
	PersistentVolumeClaimName *string `json:"persistentVolumeClaimName,omitempty"`
	// In custom serving, user may specify the mount path of the pvc
	Path string `json:"path,omitempty"`
}

// UserInformation is the type to store the user-related information.
// The information will be used to compose the PodSpec in the deployment.
type UserInformation struct {
	// Group defines the group that the user belongs to.
	Group string `json:"group,omitempty"`
	// Username is the user's ID.
	Username string `json:"username,omitempty"`
	// Tenant is the tenant in Clever.
	Tenant string `json:"tenant"`
}

// ScalingConfig defines the configuration about how to scale the serving service.
type ScalingConfig struct {
	MinReplicas     *int32           `json:"minReplicas,omitempty"`
	MaxReplicas     int32            `json:"maxReplicas"`
	ResourceMetrics []ResourceMetric `json:"resourceMetric,omitempty"`
	CustomMetrics   []CustomMetric   `json:"customMetric,omitempty"`
}

// ResourceMetric specifies how to scale based on a single metric.
type ResourceMetric struct {
	Name corev1.ResourceName `json:"name,omitempty"`
	// At least one fields below should be set.
	// Value is not supported in v2beta1, thus we do not support value now.
	// Value              *resource.Quantity `json:"value,omitempty"`
	AverageValue       *resource.Quantity `json:"averageValue,omitempty"`
	AverageUtilization *int32             `json:"averageUtilization,omitempty"`
}

// CustomMetric defines customized resource.
type CustomMetric struct {
	// TBD
}

// +k8s:deepcopy-gen=true

// ServingStatus defines the status of serving deployment.
type ServingStatus struct {
	// Total number of non-terminated pods targeted by this serving (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Total number of non-terminated pods targeted by this serving that have the desired template spec.
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty"`

	// Total number of ready pods targeted by this serving.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this serving.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// Total number of unavailable pods targeted by this serving. This is the total number of
	// pods that are still required for the serving to have 100% available capacity. They may
	// either be pods that are running but not yet available or pods that still have not been created.
	// +optional
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`

	// ModelStatuses store the model statuses of one serving, which key is
	// model-version<-alias>, the alias was only for GPUSharing.
	ModelStatuses map[string]*ModelStatus `json:"modelStatuses,omitempty"`

	// VolumeName is the name of the Volume.
	VolumeName string `json:"volumeName,omitempty"`

	// Conditions is an array of current observed job conditions.
	Conditions []ServingCondition `json:"conditions,omitempty"`

	// Represents time when the job was acknowledged by the job controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// The generation observed by the deployment controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents last time when the spec is updated. The field is only updated by clever-admin.
	LastModifiedTime *metav1.Time `json:"lastModifiedTime,omitempty"`

	// Represents last time when the job was reconciled. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}

// ModelStatus stores the model status in one serving.
// In GPUSharing, replicas will be 1, in other servings types, replicas >= 1.
type ModelStatus struct {
	Replicas            int32 `json:"replicas,omitempty"`
	AvailableReplicas   int32 `json:"availableReplicas,omitempty"`
	UnavailableReplicas int32 `json:"unavailableReplicas,omitempty"`
}

// +k8s:deepcopy-gen=true

// ServingCondition describes the state of the serving service at a certain point.
type ServingCondition struct {
	// Type of job condition.
	Type ServingConditionType `json:"type"`
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

// ServingConditionType is the type of ServingCondition.
type ServingConditionType string

const (
	// ServingHPAScalingActive indicates that the HPA controller is able to scale if necessary:
	// it's correctly configured, can fetch the desired metrics, and isn't disabled.
	ServingHPAScalingActive ServingConditionType = "HorizontalPodAutoscalerScalingActive"
	// ServingHPAAbleToScale indicates a lack of transient issues which prevent scaling from occurring,
	// such as being in a backoff window, or being unable to access/update the target scale.
	ServingHPAAbleToScale ServingConditionType = "HorizontalPodAutoscalerAbleToScale"
	// ServingHPAScalingLimited indicates that the calculated scale based on metrics would be above or
	// below the range for the HPA, and has thus been capped.
	ServingHPAScalingLimited        ServingConditionType = "HorizontalPodAutoscalerScalingLimited"
	ServingDeploymentAvailable      ServingConditionType = "DeploymentAvailable"
	ServingDeploymentProgressing    ServingConditionType = "DeploymentProgressing"
	ServingDeploymentReplicaFailure ServingConditionType = "DeploymentReplicaFailure"
	ServingHealth                   ServingConditionType = "ModelHealth"
	ServingPVCResizing              ServingConditionType = "PersistentVolumeClaimResizing"
	ServingPVCPending               ServingConditionType = "PersistentVolumeClaimPending"
	ServingPVCBound                 ServingConditionType = "PersistentVolumeClaimBound"
)

// +k8s:deepcopy-gen=true

// ServingModel is a model [with version] served with a serving instance.
// A model is uniquely identified via name and version.
type ServingModel struct {
	// Name and version, version is optional
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`

	// Time when the model is added to the serving instance.
	AdditionTime *metav1.Time `json:"additionTime,omitempty"`

	// Servering framework specific configurations
	TensorRTModelConfig *TensorRTModelConfig `json:"tensorRTConfig,omitempty"`
}

// TensorRTModelConfig is the configuration related to TensorRT.
type TensorRTModelConfig struct {
	// Number of instance per model.
	InstanceNum int32 `json:"instanceNum,omitempty"`
	// URLAlias in /api/infer/<URLAlias>/[Version(=1 in 1.4.0)]
	// URLAlias is not used in Scene Serving but TensorRT Inference Serve
	URLAlias string `json:"urlAlias,omitempty"`
}

// -----------------------------------------------------------------

const (
	ScenePlural = "scenes"
)

// Scene defines a scene service
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type Scene struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SceneSpec   `json:"spec,omitempty"`
	Status            SceneStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SceneList describes an array of Scenes
type SceneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Scenes
	Items []Scene `json:"items"`
}

type SceneStatus struct {
	// Total number of servings.
	// +optional
	Servings int32 `json:"servings,omitempty"`

	// Total number of ready servings.
	// +optional
	AvailableServings int32 `json:"availableServings,omitempty"`

	// Total number of unavailable servings.
	// +optional
	UnavailableServings int32 `json:"unavailableServings,omitempty"`

	// Conditions is an array of current observed job conditions.
	Conditions []SceneCondition `json:"conditions,omitempty"`

	// VolumeName is the name of the Volume.
	VolumeName string `json:"volumeName,omitempty"`

	// Represents time when the job was acknowledged by the job controller.
	// It is not guaranteed to be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// The generation observed by the deployment controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents last time when the spec is updated. The field is only updated by clever-admin.
	LastModifiedTime *metav1.Time `json:"lastModifiedTime,omitempty"`

	// Represents last time when the job was reconciled. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	LastReconcileTime *metav1.Time `json:"lastReconcileTime,omitempty"`
}

type SceneCondition struct {
	// Type of job condition.
	Type SceneConditionType `json:"type"`
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

type SceneConditionType string

const (
	SceneRunning SceneConditionType = "Running"
	SceneHealth  SceneConditionType = "SceneHealth"

	ScenePVCResizing SceneConditionType = "PersistentVolumeClaimResizing"
	ScenePVCPending  SceneConditionType = "PersistentVolumeClaimPending"
	ScenePVCBound    SceneConditionType = "PersistentVolumeClaimBound"
)

// ServingTemplateSpec describes a template to create a serving deployment.
type ServingTemplateSpec struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServingSpec `json:"spec,omitempty"`
}

type SceneSpec struct {
	// UserInfo is the information about the user.
	// If UserInfo is defined here, it will override the corresponding fields in servings.
	UserInfo *UserInformation `json:"userInfo,omitempty"`
	// A list of serving deployments under a scene.
	Servings []ServingTemplateSpec `json:"servings"`
	// A list of route configuration of the scene.
	// To match Istio-Admin struct definition:
	// https://github.com/caicloud/platform/blob/master/docs/api/istio-admin.md#trafficrule
	Http []*HTTPRoute `json:"httpRoute"`
	// Name of default serving deployment.
	DefaultServing string `json:"defaultServing,omitempty"`
}

type HTTPRoute struct {
	// A list of rules to match requests. All matches are ORed.
	Match []*HTTPMatchRequest `json:"match,omitempty"`
	// A list of route information for matched requests.
	Route []*HTTPRouteServing `json:"route"`
	// Path sending to Pod in Scene
	Rewrite *HTTPRewrite `json:"rewrite,omitempty"`
}

type HTTPRewrite struct {
	// In Clever 1.4.0, the rewrite path is "/predict"
	Uri string `json:"uri,omitempty"`
}

// HTTPMatchRequest specify rules to match requests. All rules are ANDed.
type HTTPMatchRequest struct {
	// Match headers of a request.
	Headers map[string]*StringMatch `json:"headers"`
	// Path that external users use
	Uri *StringMatch `json:"uri,omitempty"`
	// Port that external users use
	Port uint32 `json:"port,omitempty"`
}

type HTTPRouteServing struct {
	// Name of serving defined in []SceneSpec.Servings.
	// Compatible with https://github.com/caicloud/platform/blob/master/docs/api/istio-admin.md#httproute
	Destination Destination `json:"destination"`
	// Traffic weight of the serving.
	Weight int32 `json:"weight,omitempty"`
}

// Destination describes the name of the Serving where traffic lands
type Destination struct {
	Subset string `json:"subset"`
	Port   uint32 `json:"port,omitempty"`
}

// StringMatch defines 3 different types of matching strategy, i.e. only match prefix,
// exact string match, and regular expression match.
type StringMatch struct {
	Prefix  string `json:"prefix,omitempty"`
	Exact   string `json:"exact,omitempty"`
	Regex   string `json:"regex,omitempty"`
	Include string `json:"include,omitempty"`
	Exclude string `json:"exclude,omitempty"`
}
