package v1alpha1

import "fmt"

const (
	// SourceOrchestration marked the release from orchestration
	SourceOrchestration = "orchestration"
	// SourceMicroservice marked the release from microservice
	SourceMicroservice = "microservice"
)

var (
	// LabelKeySource - release.caicloud.io/source
	LabelKeySource = fmt.Sprintf("%s/source", GroupName)
	// LabelKeyKind - release.caicloud.io/kind = deployments statefulsets cronjobs jobs
	LabelKeyKind = fmt.Sprintf("%s/kind", GroupName)
)
