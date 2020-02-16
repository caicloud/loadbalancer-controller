package v1alpha3

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FlavorPlural is the Plural for Flavor.
	FlavorPlural = "flavors"
)

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
	CustomImageType ImageType = "Custom"

	// PresetImageType represents a preset image of the platform.
	PresetImageType ImageType = "Preset"

	// PresetImageType represents a preset pure image of the platform.
	PresetPureImageType ImageType = "PresetPure"
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

// Recommended resource configuration with limit
type ResourceConfig corev1.ResourceRequirements

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type FlavorList struct {
	metav1.TypeMeta `json:",inline"`

	// Standard list metadata.
	metav1.ListMeta `json:"metadata"`

	// List of Flavor
	Items []Flavor `json:"items"`
}
