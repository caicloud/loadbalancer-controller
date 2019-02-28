package v1alpha1

import "fmt"

const (
	// ApplicationPlural is the plural of application resources
	ApplicationPlural = "applications"
	// ApplicationDraftPlural is the plural of applicationDraft resources
	ApplicationDraftPlural = "applicationdrafts"
)

var (
	// LabelKeyCreatedBy - orchestration.caicloud.io/created-by
	LabelKeyCreatedBy = fmt.Sprintf("%s/created-by", GroupName)
)
