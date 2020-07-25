package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	SolutionTypeSurveillance      SolutionType = "SurveillanceVideoAnalysis"
	SolutionTypeCustomTemplateOCR SolutionType = "CustomTemplateOCR"
)

type AnalysisSceneType string

const (
	// Traffic statistics
	SceneTypeCounting AnalysisSceneType = "counting"
	// Retention detection
	SceneTypeOccupancy AnalysisSceneType = "occupancy"
	// Congestion detection
	SceneTypeCongestion AnalysisSceneType = "congestion"
	// Target Tracking
	SceneTypeTracking AnalysisSceneType = "tracking"
)

type SurveillanceVideoAnalysis struct {
	// Only support `car, person, animal`
	ObjectClass []string          `json:"objectClass"`
	SceneType   AnalysisSceneType `json:"sceneType"`
	// One frame of the video, can be proportional scaled
	FrameImage FrameImage `json:"frameImage"`
	// Region of interest: a region (usually a polygon) consisting of coordinates
	ROI []Coordinate `json:"roi"`
	// Videos to analyze
	FileList       []string                    `json:"fileList"`
	OccupancyArgs  *OccupancyArgs              `json:"occupancyArgs, omitempty"`
	CongestionArgs *CongestionArgs             `json:"congestionArgs, omitempty"`
	Resources      corev1.ResourceRequirements `json:"resources"`
}

type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type FrameImage struct {
	Size ImageSize `json:"size"`
	Path string    `json:"path" `
}

type OccupancyArgs struct {
	TimeLimit string `json:"timeLimit"`
}

type CongestionArgs struct {
	NumLimit  int    `json:"numLimit"`
	TimeLimit string `json:"timeLimit"`
}
