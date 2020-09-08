package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Coordinate is a common type for all solutions
// Represents the position of a point on the picture
type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// SolutionConfig contains configs of all solutions,
type SolutionConfig struct {
	SurveillanceVideoAnalysis *SurveillanceVideoAnalysis `json:"surveillanceVideoAnalysis"`
	CustomTemplateOCR         *CustomTemplateOCR         `json:"customTemplateOCR"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient:noStatus

// Solution contains all information of solutions supported by clever
// Only one information of solution exists at the same time
type Solution struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SolutionSpec `json:"spec"`
}

type SolutionSpec struct {
	MediaType  MediaType      `json:"mediaType"`
	Technology TechnologyType `json:"technology"`
	Type       SolutionType   `json:"type"`
	Config     SolutionConfig `json:"config"`
}

// MediaType represent media type, which contain Text, Image, Audio and Video
type MediaType string

const (
	MediaTypeText  MediaType = "Text"
	MediaTypeImage MediaType = "Image"
	MediaTypeAudio MediaType = "Audio"
	MediaTypeVideo MediaType = "Video"
)

type TechnologyType string

const (
	// Speech technology
	Speech TechnologyType = "Speech"
	// Computer Vision
	CV TechnologyType = "CV"
	// Natural Language Processing
	NLP TechnologyType = "NLP"
	// Machine Learning
	ML TechnologyType = "ML"
)

// SolutionType represent clever solution type, which contain
// SolutionTypeSurveillance and SolutionTypeCustomTemplateOCR
type SolutionType string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SolutionList is a list of Solution
type SolutionList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Solution `json:"items"`
}
